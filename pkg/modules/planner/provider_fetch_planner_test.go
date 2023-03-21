package planner

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/json_util"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderFetchPlanner_MakePlan(t *testing.T) {

	//rootModule := module.NewModule()
	//rootModule.SelefraBlock = module.NewSelefraBlock()
	//rootModule.SelefraBlock.RequireProvidersBlock = []*module.RequireProviderBlock{
	//	{
	//		Name:    "aws",
	//		Source:  "aws",
	//		Version: "latest",
	//	},
	//	{
	//		Name:    "gcp",
	//		Source:  "gcp",
	//		Version: "latest",
	//	},
	//}
	//rootModule.ProvidersBlock = []*module.ProviderBlock{
	//	{
	//		Name:          "aws-001",
	//		Provider:      "aws",
	//		MaxGoroutines: pointer.ToUInt64Pointer(10),
	//	},
	//	{
	//		Name:          "aws-002",
	//		Provider:      "aws",
	//		MaxGoroutines: pointer.ToUInt64Pointer(30),
	//	},
	//}
	//versionWinnerMap := map[string]string{
	//	"aws": "v0.0.1",
	//	"gcp": "v0.0.1",
	//}
	//plan, diagnostics := NewProviderFetchPlanner(&ProviderFetchPlannerOptions{
	//	Module:                       rootModule,
	//	ProviderVersionVoteWinnerMap: versionWinnerMap,
	//}).MakePlan(context.Background())
	//assert.False(t, utils.HasError(diagnostics))
	//assert.Equal(t, 3, len(plan))

	// load module
	moduleDirectory := "./test_data/provider_fetch_planner"
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.String())
		}
	})
	loader, err := module_loader.NewLocalDirectoryModuleLoader(&module_loader.LocalDirectoryModuleLoaderOptions{
		ModuleLoaderOptions: &module_loader.ModuleLoaderOptions{
			Source:            moduleDirectory,
			Version:           "",
			DownloadDirectory: "./test_download",
			ProgressTracker:   nil,
			MessageChannel:    messageChannel,
		},
		ModuleDirectory: moduleDirectory,
	})
	assert.Nil(t, err)
	rootModule, isLoadSuccess := loader.Load(context.Background())
	messageChannel.ReceiverWait()
	assert.True(t, isLoadSuccess)

	// check module
	validatorContext := module.NewValidatorContext()
	d := rootModule.Check(rootModule, validatorContext)
	if utils.IsNotEmpty(d) {
		t.Log(d.String())
	}
	assert.False(t, utils.HasError(d))
	if utils.HasError(d) {
		return
	}

	versionWinnerMap := map[string]string{
		"aws": "v0.0.1",
		"gcp": "v0.0.1",
	}
	messageChannel = message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.String())
		}
	})
	providerFetchPlans, diagnostics := NewProviderFetchPlanner(&ProviderFetchPlannerOptions{
		Module:                       rootModule,
		ProviderVersionVoteWinnerMap: versionWinnerMap,
		DSN:                          env.GetDatabaseDsn(),
		MessageChannel:               messageChannel,
	}).MakePlan(context.Background())
	messageChannel.ReceiverWait()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.String())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotEqual(t, 0, len(providerFetchPlans))

	t.Log(json_util.ToJsonString(providerFetchPlans))

}
