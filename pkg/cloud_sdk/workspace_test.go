package cloud_sdk

import (
	"context"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudClient_UploadWorkspace(t *testing.T) {
	client := getAuthedSDKClientForTest()

	project, diagnostics := client.CreateProject("cli-test-project-" + id_util.RandomId())
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, project)

	d := client.UploadWorkspace(context.Background(), project.Name, "./test_data/sync_workspace")
	assert.False(t, utils.HasError(d))
}
