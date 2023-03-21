package grpc

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"testing"
)

type testGRpcClient struct {
}

var _ StreamClient[*testGRpcRequest, *testGRpcResponse]

func (t testGRpcClient) Send(request *testGRpcRequest) error {
	fmt.Println("send message: " + request.id)
	return nil
}

func (t testGRpcClient) CloseAndRecv() (*testGRpcResponse, error) {
	fmt.Println("close stream")
	return nil, nil
}

func (t testGRpcClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (t testGRpcClient) Trailer() metadata.MD {
	return nil
}

func (t testGRpcClient) CloseSend() error {
	return nil
}

func (t testGRpcClient) Context() context.Context {
	return context.Background()
}

func (t testGRpcClient) SendMsg(m interface{}) error {
	return nil
}

func (t testGRpcClient) RecvMsg(m interface{}) error {
	return nil
}

type testGRpcRequest struct {
	id string
}

type testGRpcResponse struct {
	id string
}

func TestNewStreamUploader(t *testing.T) {

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})
	options := &StreamUploaderOptions[*testGRpcClient, string, *testGRpcRequest, *testGRpcResponse]{
		Name:                      "test-stream-uploader",
		Client:                    &testGRpcClient{},
		WaitSendTaskQueueBuffSize: 1,
		MessageChannel:            messageChannel,
	}
	uploader := NewStreamUploader(options)
	uploader.RunUploaderWorker()

	for i := 0; i < 100; i++ {
		id := id_util.RandomId()
		submitSuccess, diagnostics := uploader.Submit(context.Background(), id, &testGRpcRequest{id: id})
		assert.False(t, utils.HasError(diagnostics))
		assert.True(t, submitSuccess)
	}
	uploader.ShutdownAndWait(context.Background())
	messageChannel.ReceiverWait()

}
