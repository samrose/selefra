package executors

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/planner"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/selefra/selefra/pkg/version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderInstallExecutor_Execute(t *testing.T) {

	providerInstallPlans := []*planner.ProviderInstallPlan{
		planner.NewProviderInstallPlan("alicloud", "v0.0.1"),
		planner.NewProviderInstallPlan("alicloud", "v0.0.2"),
		planner.NewProviderInstallPlan("alicloud", "v0.0.3"),
		planner.NewProviderInstallPlan("alicloud", version.VersionLatest),
		planner.NewProviderInstallPlan("gcp", version.VersionLatest),
	}

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})
	executor, diagnostics := NewProviderInstallExecutor(&ProviderInstallExecutorOptions{
		Plans:             providerInstallPlans,
		MessageChannel:    messageChannel,
		DownloadWorkspace: "./test_download",
	})
	assert.False(t, utils.HasError(diagnostics))
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	d := executor.Execute(context.Background())
	assert.False(t, utils.HasError(d))
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}
	messageChannel.ReceiverWait()
}
