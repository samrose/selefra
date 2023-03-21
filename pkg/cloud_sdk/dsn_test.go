package cloud_sdk

import (
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudClient_FetchOrgDSN(t *testing.T) {
	client := getAuthedSDKClientForTest()
	dsn, diagnostics := client.FetchOrgDSN()
	assert.False(t, utils.HasError(diagnostics))
	assert.NotEmpty(t, dsn)
}
