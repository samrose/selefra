package cloud_sdk

import (
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudClient_CreateTask(t *testing.T) {
	client := getAuthedSDKClientForTest()

	project, diagnostics := client.CreateProject("cli-test-project-" + id_util.RandomId())
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, project)

	response, d := client.CreateTask(project.Name)
	assert.False(t, utils.HasError(d))
	assert.NotNil(t, response)
}
