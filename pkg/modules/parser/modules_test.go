package parser

import (
	"fmt"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlFileToModuleParser_parseModulesBlock(t *testing.T) {
	module, diagnostics := NewYamlFileToModuleParser("./test_data/test_parse_modules/modules.yaml").Parse()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, module.ModulesBlock)

	moduleBlock := module.ModulesBlock[1]
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, moduleBlock.GetNodeLocation("name._key").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("name._value").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("name").ReadSourceString())

	assert.NotEmpty(t, moduleBlock.GetNodeLocation("input._key").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("input._value").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("input").ReadSourceString())

	assert.NotEmpty(t, moduleBlock.GetNodeLocation("input.name._key").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("input.name._value").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("input.name").ReadSourceString())

	assert.NotEmpty(t, moduleBlock.GetNodeLocation("uses._key").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("uses._value").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("uses").ReadSourceString())

	assert.NotEmpty(t, moduleBlock.GetNodeLocation("uses[0]").ReadSourceString())
	assert.NotEmpty(t, moduleBlock.GetNodeLocation("uses[0]._value").ReadSourceString())

	moduleBlock = module.ModulesBlock[0]
	for i := 0; i < len(moduleBlock.Uses); i++ {
		assert.NotEmpty(t, moduleBlock.GetNodeLocation(fmt.Sprintf("uses[%d]", i)).ReadSourceString())
		assert.NotEmpty(t, moduleBlock.GetNodeLocation(fmt.Sprintf("uses[%d]._value", i)).ReadSourceString())
	}

}
