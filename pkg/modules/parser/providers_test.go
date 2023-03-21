package parser

import (
	"fmt"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlFileToModuleParser_parseProvidersBlock(t *testing.T) {
	module, diagnostics := NewYamlFileToModuleParser("./test_data/test_parse_providers/modules.yaml").Parse()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, module.ProvidersBlock)

	providerBlock := module.ProvidersBlock[0]
	assert.NotEmpty(t, providerBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, providerBlock.GetNodeLocation("name._key").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("name._value").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("name").ReadSourceString())

	assert.NotEmpty(t, providerBlock.GetNodeLocation("cache._key").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("cache._value").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("cache").ReadSourceString())

	assert.NotEmpty(t, providerBlock.GetNodeLocation("resources._key").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("resources._value").ReadSourceString())
	assert.NotEmpty(t, providerBlock.GetNodeLocation("resources").ReadSourceString())

	for i := 0; i < len(providerBlock.Resources); i++ {
		assert.NotEmpty(t, providerBlock.GetNodeLocation(fmt.Sprintf("resources[%d]", i)).ReadSourceString())
		assert.NotEmpty(t, providerBlock.GetNodeLocation(fmt.Sprintf("resources[%d]._value", i)).ReadSourceString())
	}
}
