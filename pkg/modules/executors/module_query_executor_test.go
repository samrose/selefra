package executors

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModuleQueryExecutor_Execute(t *testing.T) {

	projectWorkspace := "./test_data/test_query_module"
	downloadWorkspace := "./test_download"

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})

	Instructions := make(map[string]interface{})
	Instructions["gpt"] = "Please help me analyze the vulnerabilities in AWS S3?"
	d := NewProjectLocalLifeCycleExecutor(&ProjectLocalLifeCycleExecutorOptions{
		Instruction:                          Instructions,
		ProjectWorkspace:                     projectWorkspace,
		DownloadWorkspace:                    downloadWorkspace,
		MessageChannel:                       messageChannel,
		ProjectLifeCycleStep:                 ProjectLifeCycleStepQuery,
		FetchStep:                            FetchStepFetch,
		ProjectCloudLifeCycleExecutorOptions: nil,
		DSN:                                  "host=127.0.0.1 user=postgres password=pass port=5432 dbname=postgres sslmode=disable",
		FetchWorkerNum:                       1,
		QueryWorkerNum:                       1,
	}).Execute(context.Background())
	messageChannel.ReceiverWait()
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}
	assert.False(t, utils.HasError(d))

}
