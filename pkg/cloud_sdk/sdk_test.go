package cloud_sdk

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/grpc/pb/issue"
	"github.com/selefra/selefra/pkg/grpc/pb/log"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func getUnAuthSDKClientForTest() *CloudClient {
	client, diagnostics := NewCloudClient(cli_env.GetServerHost())
	if utils.HasError(diagnostics) {
		panic(diagnostics.ToString())
	}
	return client
}

func getAuthedSDKClientForTest() *CloudClient {
	client, diagnostics := NewCloudClient(cli_env.GetServerHost())
	if utils.HasError(diagnostics) {
		panic(diagnostics.ToString())
	}
	token := cli_env.GetCloudToken()
	cloudCredentials, d := client.Login(token)
	if utils.HasError(d) {
		panic(d.ToString())
	}
	if cloudCredentials == nil {
		panic("cloud credentials is nil")
	}
	return client
}

func TestNewCloudClient(t *testing.T) {
	client, diagnostics := NewCloudClient(cli_env.GetServerHost())
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, client)
}

func TestCloudClient_NewIssueStreamUploader(t *testing.T) {
	client := getAuthedSDKClientForTest()

	project, diagnostics := client.CreateProject("cli-test-project-" + id_util.RandomId())
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, project)

	_, d := client.CreateTask(project.Name)
	assert.False(t, utils.HasError(d))
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})
	//client.MessageChannel = messageChannel
	issueStreamUploader, d := client.NewIssueStreamUploader(messageChannel)
	assert.False(t, utils.HasError(d))
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}
	issueStreamUploader.RunUploaderWorker()

	for i := 0; i < 10000; i++ {
		ok, d := issueStreamUploader.Submit(context.Background(), i, &issue.UploadIssueStream_Request{
			Index: int32(i),
			Rule: &issue.UploadIssueStream_Rule{
				Name:     "test-rule",
				Query:    "selefra * from 1",
				Metadata: &issue.UploadIssueStream_Metadata{},
				Output:   "output",
			},
			Provider: nil,
			Module:   nil,
			Context: &issue.UploadIssueStream_Context{
				SrcTableNames: []string{
					"foo", "bar", "test",
				},
				Schema: "test",
			},
		})
		assert.Nil(t, d)
		assert.True(t, ok)
	}

	issueStreamUploader.ShutdownAndWait(context.Background())
	messageChannel.ReceiverWait()

}

func TestCloudClient_NewLogStreamUploader(t *testing.T) {
	client := getAuthedSDKClientForTest()

	project, diagnostics := client.CreateProject("cli-test-project-" + id_util.RandomId())
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, project)

	_, d := client.CreateTask(project.Name)
	assert.False(t, utils.HasError(d))
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})
	//client.MessageChannel = messageChannel
	logClient, logStreamUploader, d := client.NewLogStreamUploader(messageChannel)
	assert.False(t, utils.HasError(d))
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}
	logStreamUploader.RunUploaderWorker()

	status, err := logClient.UploadLogStatus(client.BuildMetaContext(), &log.UploadLogStatus_Request{
		Stage:  log.StageType_STAGE_TYPE_INITIALIZING,
		Status: log.Status_STATUS_SUCCESS,
		Time:   timestamppb.Now(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, status)

	for i := 0; i < 10000; i++ {
		ok, d := logStreamUploader.Submit(context.Background(), i, &log.UploadLogStream_Request{
			Stage: log.StageType_STAGE_TYPE_INITIALIZING,
			Index: uint64(i),
			Msg:   fmt.Sprintf("test %d", i),
			Level: log.Level_LEVEL_DEBUG,
			Time:  timestamppb.Now(),
		})
		assert.Nil(t, d)
		assert.True(t, ok)
	}

	logStreamUploader.ShutdownAndWait(context.Background())
	messageChannel.ReceiverWait()

}
