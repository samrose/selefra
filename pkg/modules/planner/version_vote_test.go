package planner

import (
	"context"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestNewProviderVersionVoteService(t *testing.T) {

	// case 1: To be able to pick a definitive version
	rootModule := randomModule("v0.0.1, v0.0.2")
	rootModule.SubModules = append(rootModule.SubModules, randomModule("v0.0.2, v0.0.3"))
	service := NewProviderVersionVoteService()
	rootModule.Traversal(context.Background(), func(ctx context.Context, traversalContext *module.TraversalContext) bool {
		d := service.Vote(context.Background(), traversalContext.Module)
		assert.False(t, utils.HasError(d))
		return true
	})
	slice := service.providerVersionVoteMap["aws"].GetWinnersVersionSlice()
	assert.Equal(t, []string{"v0.0.2"}, slice)

	// case 2: The ability to select multiple explicit versions
	rootModule = randomModule("v0.0.1, v0.0.2, v0.0.3")
	rootModule.SubModules = append(rootModule.SubModules, randomModule("v0.0.2, v0.0.3, v0.0.4"))
	service = NewProviderVersionVoteService()
	rootModule.Traversal(context.Background(), func(ctx context.Context, traversalContext *module.TraversalContext) bool {
		d := service.Vote(context.Background(), traversalContext.Module)
		assert.False(t, utils.HasError(d))
		return true
	})
	slice = service.providerVersionVoteMap["aws"].GetWinnersVersionSlice()
	sort.Strings(slice)
	assert.Equal(t, []string{"v0.0.2", "v0.0.3"}, slice)

	// case 3: No definitive version could be selected
	rootModule = randomModule("v0.0.1, v0.0.2")
	rootModule.SubModules = append(rootModule.SubModules, randomModule("v0.0.4"))
	service = NewProviderVersionVoteService()
	rootModule.Traversal(context.Background(), func(ctx context.Context, traversalContext *module.TraversalContext) bool {
		d := service.Vote(context.Background(), traversalContext.Module)
		assert.False(t, utils.HasError(d))
		return true
	})
	slice = service.providerVersionVoteMap["aws"].GetWinnersVersionSlice()
	assert.Equal(t, []string{}, slice)

}

func randomModule(requiredVersion string) *module.Module {

	requireProviderBlock := module.NewRequireProviderBlock()
	requireProviderBlock.Source = "aws"
	requireProviderBlock.Name = "aws"
	requireProviderBlock.Version = requiredVersion

	rootModule := module.NewModule()
	rootModule.SelefraBlock = &module.SelefraBlock{
		RequireProvidersBlock: []*module.RequireProviderBlock{
			requireProviderBlock,
		},
	}

	return rootModule
}
