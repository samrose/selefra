package parser

import (
	"fmt"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlFileToModuleParser_parseRulesBlock(t *testing.T) {
	module, diagnostics := NewYamlFileToModuleParser("./test_data/test_parse_rules/modules.yaml", make(map[string]interface{})).Parse()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, module.RulesBlock)

	ruleBlock := module.RulesBlock[0]
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, ruleBlock.GetNodeLocation("name._key").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("name._value").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("name").ReadSourceString())

	assert.NotEmpty(t, ruleBlock.GetNodeLocation("query._key").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("query._value").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("query").ReadSourceString())

	assert.NotEmpty(t, ruleBlock.GetNodeLocation("output._key").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("output._value").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("output").ReadSourceString())

	assert.NotEmpty(t, ruleBlock.GetNodeLocation("labels._key").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("labels._value").ReadSourceString())
	assert.NotEmpty(t, ruleBlock.GetNodeLocation("labels").ReadSourceString())

	for key, _ := range ruleBlock.Labels {
		assert.NotEmpty(t, ruleBlock.GetNodeLocation(fmt.Sprintf("labels.%s", key)).ReadSourceString())
		assert.NotEmpty(t, ruleBlock.GetNodeLocation(fmt.Sprintf("labels.%s._key", key)).ReadSourceString())
		assert.NotEmpty(t, ruleBlock.GetNodeLocation(fmt.Sprintf("labels.%s._value", key)).ReadSourceString())
	}

	metadataBlock := ruleBlock.MetadataBlock
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, metadataBlock.GetNodeLocation("author._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("author._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("author").ReadSourceString())

	assert.NotEmpty(t, metadataBlock.GetNodeLocation("description._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("description._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("description").ReadSourceString())

	assert.NotEmpty(t, metadataBlock.GetNodeLocation("id._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("id._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("id").ReadSourceString())

	assert.NotEmpty(t, metadataBlock.GetNodeLocation("provider._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("provider._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("provider").ReadSourceString())

	assert.NotEmpty(t, metadataBlock.GetNodeLocation("remediation._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("remediation._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("remediation").ReadSourceString())

	assert.NotEmpty(t, metadataBlock.GetNodeLocation("severity._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("severity._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("severity").ReadSourceString())

	for i := 0; i < len(metadataBlock.Tags); i++ {
		assert.NotEmpty(t, metadataBlock.GetNodeLocation(fmt.Sprintf("tags[%d]", i)).ReadSourceString())
		assert.NotEmpty(t, metadataBlock.GetNodeLocation(fmt.Sprintf("tags[%d]._value", i)).ReadSourceString())
	}
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("title._key").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("title._value").ReadSourceString())
	assert.NotEmpty(t, metadataBlock.GetNodeLocation("title").ReadSourceString())

}
