package parser

import (
	"fmt"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlFileToModuleParser_Parse(t *testing.T) {
	module, diagnostics := NewYamlFileToModuleParser("./test_data/test_modules.yaml").Parse()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))

	location := module.RulesBlock[0].MetadataBlock.GetNodeLocation("tags[0]._value")
	s := location.ReadSourceString()
	fmt.Println(s)
}
