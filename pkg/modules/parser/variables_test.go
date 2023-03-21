package parser

import (
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlFileToModuleParser_parseVariablesBlock(t *testing.T) {
	module, diagnostics := NewYamlFileToModuleParser("./test_data/test_parse_variables/modules.yaml").Parse()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, module.VariablesBlock)

	variableBLock := module.VariablesBlock[0]
	assert.NotEmpty(t, variableBLock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, variableBLock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, variableBLock.GetNodeLocation("key").ReadSourceString())
	assert.NotEmpty(t, variableBLock.GetNodeLocation("key._key").ReadSourceString())
	assert.NotEmpty(t, variableBLock.GetNodeLocation("key._value").ReadSourceString())

	assert.NotEmpty(t, variableBLock.GetNodeLocation("default").ReadSourceString())
	assert.NotEmpty(t, variableBLock.GetNodeLocation("default._key").ReadSourceString())
	assert.NotEmpty(t, variableBLock.GetNodeLocation("default._value").ReadSourceString())

}
