package message

import (
	"github.com/selefra/selefra-utils/pkg/reflect_util"
	"sync"
)

// Channel Used to link multiple channels and coordinate messaging in tree invocation relationships
type Channel[Message any] struct {

	// Current channel
	channel chan Message

	// Control of subchannels
	subChannelWg *sync.WaitGroup
	selfWg       *sync.WaitGroup

	// The callback on shutdown
	closeCallbackFunc func()

	// The current channel processes the message
	consumerFunc func(index int, message Message)
}

func NewChannel[Message any](consumerFunc func(index int, message Message), buffSize ...int) *Channel[Message] {

	// can have buff
	var channel chan Message
	if len(buffSize) != 0 {
		channel = make(chan Message, buffSize[0])
	} else {
		channel = make(chan Message)
	}

	x := &Channel[Message]{
		channel:      channel,
		subChannelWg: &sync.WaitGroup{},
		selfWg:       &sync.WaitGroup{},
		consumerFunc: consumerFunc,
	}

	x.selfWg.Add(1)
	go func() {

		// The exit of the channel consumer indicates that the channel is closed, and a callback event is triggered when the channel is closed
		defer func() {
			x.selfWg.Done()
			if x.closeCallbackFunc != nil {
				x.closeCallbackFunc()
			}
		}()

		count := 1
		for message := range x.channel {
			if x.consumerFunc != nil {
				x.consumerFunc(count, message)
			}
		}
	}()

	return x
}

func (x *Channel[Message]) Send(message Message) {
	if !reflect_util.IsNil(message) {
		x.channel <- message
	}
}

func (x *Channel[Message]) MakeChildChannel() *Channel[Message] {

	// Adds a semaphore to the parent channel
	x.subChannelWg.Add(1)

	// Create a child channel and bridge it to the parent channel
	subChannel := NewChannel[Message](func(index int, message Message) {
		x.channel <- message
	})

	// Reduces the semaphore of the parent channel when the child channel is turned off
	subChannel.closeCallbackFunc = func() {
		x.subChannelWg.Done()
	}

	return subChannel
}

func (x *Channel[Message]) ReceiverWait() {
	x.selfWg.Wait()
}

func (x *Channel[Message]) SenderWaitAndClose() {
	x.subChannelWg.Wait()
	close(x.channel)
	x.selfWg.Wait()
}
