package planner

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRulePlanner_MakePlan(t *testing.T) {

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})
	loader, err := module_loader.NewLocalDirectoryModuleLoader(&module_loader.LocalDirectoryModuleLoaderOptions{
		ModuleLoaderOptions: &module_loader.ModuleLoaderOptions{
			MessageChannel: messageChannel,
		},
		ModuleDirectory: "./test_data/rule_planner",
	})
	assert.Nil(t, err)
	rootModule, b := loader.Load(context.Background())
	assert.True(t, b)
	messageChannel.ReceiverWait()

	scope := NewScope()
	scope.SetVariable("account_id", "100000875657")
	tableToProviderMap := map[string]string{
		"aws_s3_buckets":           "aws",
		"aws_s3_bucket_cors_rules": "aws",
	}

	options := &RulePlannerOptions{
		ModulePlan:         nil,
		Module:             rootModule,
		ModuleScope:        scope,
		RuleBlock:          rootModule.RulesBlock[0],
		TableToProviderMap: tableToProviderMap,
	}
	plan, diagnostics := NewRulePlanner(options).MakePlan(context.Background())
	t.Log(diagnostics.ToString())
	assert.False(t, utils.HasError(diagnostics))
	//t.Log(json_util.ToJsonString(plan))
	assert.NotEmpty(t, plan.Query)
	assert.NotEmpty(t, plan.BindingProviderName)
	assert.NotEmpty(t, plan.BindingTables)
	assert.Len(t, plan.BindingTables, 2)

}
