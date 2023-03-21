package grpc

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"google.golang.org/grpc"
	"sync"
	"time"
)

// ------------------------------------------------- --------------------------------------------------------------------

type StreamClient[Request, Response any] interface {
	Send(Request) error
	CloseAndRecv() (Response, error)
	grpc.ClientStream
}

// ------------------------------------------------- --------------------------------------------------------------------

type StreamUploaderOptions[Client, ID, Request, Response any] struct {

	// Used to distinguish between different instances
	Name string

	// The client used in the grpc call
	Client StreamClient[Request, Response]

	// The size of the queue of tasks waiting to be sent
	WaitSendTaskQueueBuffSize int

	// Used to send messages to the outside world
	MessageChannel *message.Channel[*schema.Diagnostics]
}

// ------------------------------------------------- --------------------------------------------------------------------

type StreamUploader[Client, ID, Request, Response any] struct {

	//Various options when uploading
	options *StreamUploaderOptions[Client, ID, Request, Response]

	// The queue of logs waiting to be sent
	waitSendTaskQueue chan *UploadTask[ID, Request]

	// Used to coordinate several workers
	workerWg sync.WaitGroup
}

func NewStreamUploader[Client, ID, Request, Response any](options *StreamUploaderOptions[Client, ID, Request, Response]) *StreamUploader[Client, ID, Request, Response] {
	return &StreamUploader[Client, ID, Request, Response]{
		options:           options,
		waitSendTaskQueue: make(chan *UploadTask[ID, Request], options.WaitSendTaskQueueBuffSize),
		workerWg:          sync.WaitGroup{},
	}
}

func (x *StreamUploader[Client, ID, Request, Response]) GetOptions() *StreamUploaderOptions[Client, ID, Request, Response] {
	return x.options
}

// Submit the message to the send queue
func (x *StreamUploader[Client, ID, Request, Response]) Submit(ctx context.Context, id ID, request Request) (bool, *schema.Diagnostics) {

	task := &UploadTask[ID, Request]{
		TaskId:  id,
		Request: request,
	}

	for submitTryTimes := 0; submitTryTimes < 10000; submitTryTimes++ {
		logger.InfoF("stream uploader name %s, id = %s, submit begin, try times = %d", x.options.Name, utils.Strava(id), submitTryTimes)
		select {
		case x.waitSendTaskQueue <- task:
			logger.InfoF("stream uploader name %s, id = %s, submit success, try times = %d", x.options.Name, utils.Strava(id), submitTryTimes)
			return true, nil
		case <-ctx.Done():
			logger.InfoF("stream uploader name %s, id = %s, submit timeout, try times = %d", x.options.Name, utils.Strava(id), submitTryTimes)
			x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("stream uploader name %s, id = %s, submit request timeout, try times = %d", x.options.Name, utils.Strava(id), submitTryTimes))
		}
	}

	logger.InfoF("stream uploader name %s, id = %s, submit final failed", x.options.Name, utils.Strava(id))
	return false, schema.NewDiagnostics().AddErrorMsg("stream uploader name %s, id = %s, submit request timeout", x.options.Name, utils.Strava(id))
}

// ShutdownAndWait Close the task queue while waiting for the remaining messages in the queue to finish sending
func (x *StreamUploader[Client, ID, Request, Response]) ShutdownAndWait(ctx context.Context) *schema.Diagnostics {

	defer func() {
		logger.InfoF("stream uploader %s message channel SenderWaitAndClose begin", x.options.Name)
		x.options.MessageChannel.SenderWaitAndClose()
		logger.InfoF("stream uploader %s message channel SenderWaitAndClose end", x.options.Name)
	}()

	close(x.waitSendTaskQueue)
	logger.InfoF("stream uploader %s close waitSendTaskQueue", x.options.Name)

	logger.InfoF("stream uploader %s wait group begin", x.options.Name)
	x.workerWg.Wait()
	logger.InfoF("stream uploader %s wait group done", x.options.Name)

	return nil
}

//func (x *StreamUploader[Client, ID, Request, Response]) runReceiveWorker() {
//
//	x.workerWg.Add(1)
//
//	go func() {
//
//		defer func() {
//			x.workerWg.Done()
//		}()
//
//		for {
//			var response Response
//			err := x.options.Client.RecvMsg(&response)
//			if err != nil {
//				if errors.Is(err, io.EOF) {
//					x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("stream uploader name %s, cloud ack receiver exit", x.options.Name))
//					return
//				} else {
//					x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("stream uploader name %s, receive response error: %s", x.options.Name, err.Error()))
//				}
//			} else {
//				id, err := x.options.ResponseAckFunc(response)
//				if err != nil {
//					x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("stream uploader name %s, extract ack id error: %s", x.options.Name, err.Error()))
//					x.options.MessageChannel.Send(x.processResponseACKFailed(id, err))
//				} else {
//					x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("stream uploader name %s, cloud ack message id %d", x.options.Name, id))
//					x.options.MessageChannel.Send(x.processResponseACKOk(id))
//				}
//			}
//		}
//	}()
//}
//
//func (x *StreamUploader[Client, ID, Request, Response]) processResponseACKOk(id ID) *schema.Diagnostics {
//
//	x.waitResponseAckTaskSetLock.Lock()
//	defer x.waitResponseAckTaskSetLock.Unlock()
//
//	// Removes the message from the ack collection
//	delete(x.waitResponseAckTaskSet, id)
//
//	return nil
//}
//
//func (x *StreamUploader[Client, ID, Request, Response]) processResponseACKFailed(id ID, err error) *schema.Diagnostics {
//
//	diagnostics := schema.NewDiagnostics()
//
//	x.waitResponseAckTaskSetLock.Lock()
//	defer x.waitResponseAckTaskSetLock.Unlock()
//
//	task, exists := x.waitResponseAckTaskSet[id]
//	if !exists {
//		return diagnostics.AddErrorMsg("stream uploader name %s, cloud ack message id %d not found", x.options.Name, id)
//	}
//
//	if task.TryTimes >= 3 {
//		return diagnostics.AddErrorMsg("stream uploader name %s, send message id %d try times used up", x.options.Name, id)
//	}
//
//}

func (x *StreamUploader[Client, ID, Request, Response]) RunUploaderWorker() {

	x.workerWg.Add(1)

	go func() {

		// Set the exit flag when exiting
		defer func() {
			logger.InfoF("stream uploader %s, begin close stream client", x.options.Name)
			response, err := x.options.Client.CloseAndRecv()
			if err != nil {
				logger.ErrorF("stream uploader %s, close stream client error: %s, response = %v", x.options.Name, err.Error(), response)
			} else {
				logger.InfoF("stream uploader %s, close stream client success, response = %v", x.options.Name, response)
			}
			x.workerWg.Done()
		}()

		timer := time.NewTimer(time.Second)
		defer timer.Stop()

		continueIdleCount := 0
		for {
			timer.Reset(time.Second)
			select {
			case task, ok := <-x.waitSendTaskQueue:

				continueIdleCount = 0

				if !ok {
					logger.InfoF("stream uploader name %s, wait send task queue closed, worker exiting", x.options.Name)
					return
				}

				err := x.options.Client.Send(task.Request)
				if err != nil {
					logger.ErrorF("stream uploader name %s, send message error: %s, id = %s", x.options.Name, err.Error(), utils.Strava(task.TaskId))
					//return
				} else {
					logger.InfoF("stream uploader name %s, send message success, id = %s", x.options.Name, utils.Strava(task.TaskId))
				}

			case <-timer.C:

				continueIdleCount++
				logger.InfoF("stream uploader name %s, wait task, idle count %d", x.options.Name, continueIdleCount)
			}
		}
	}()
}

// ------------------------------------------------- --------------------------------------------------------------------

// UploadTask Represents a task to be uploaded
type UploadTask[ID, Request any] struct {

	// What is the ID of this task
	TaskId ID

	// A request to send
	Request Request

	//TryTimes int
}

// ------------------------------------------------- --------------------------------------------------------------------

//type SyncTaskMap[ID, Request comparable] struct {
//	lock    *sync.RWMutex
//	taskMap map[ID]*UploadTask[ID, Request]
//}
//
//func NewSyncTaskMap[ID, Request comparable]() *SyncTaskMap[ID, Request] {
//	return &SyncTaskMap[ID, Request]{
//		lock:    &sync.RWMutex{},
//		taskMap: make(map[ID]*UploadTask[ID, Request]),
//	}
//}
//
//func (x *SyncTaskMap[ID, Request]) Set(id ID, task *UploadTask[ID, Request]) {
//	x.Run(func(taskMap map[ID]*UploadTask[ID, Request]) {
//		taskMap[id] = task
//	})
//}
//
//func (x *SyncTaskMap[ID, Request]) Delete(id ID, task *UploadTask[ID, Request]) {
//	x.Run(func(taskMap map[ID]*UploadTask[ID, Request]) {
//		delete(taskMap, id)
//	})
//}
//
//func (x *SyncTaskMap[ID, Request]) Get(id ID) (response *UploadTask[ID, Request], exists bool) {
//	x.Run(func(taskMap map[ID]*UploadTask[ID, Request]) {
//		response, exists = taskMap[id]
//	})
//	return
//}
//
//func (x *SyncTaskMap[ID, Request]) Run(runFunc func(taskMap map[ID]*UploadTask[ID, Request])) {
//	x.lock.Lock()
//	defer x.lock.Unlock()
//
//	runFunc(x.taskMap)
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
