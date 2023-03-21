package parser

import (
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestYamlFileToModuleParser_parseSelefraBlock(t *testing.T) {
	module, diagnostics := NewYamlFileToModuleParser("./test_data/test_parse_selefra/modules.yaml").Parse()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, module.SelefraBlock)

	selefraBlock := module.SelefraBlock
	assert.NotEmpty(t, selefraBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, selefraBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, selefraBlock.GetNodeLocation("name").ReadSourceString())
	assert.NotEmpty(t, selefraBlock.GetNodeLocation("name._key").ReadSourceString())
	assert.NotEmpty(t, selefraBlock.GetNodeLocation("name._value").ReadSourceString())

	assert.NotEmpty(t, selefraBlock.GetNodeLocation("cli_version").ReadSourceString())
	assert.NotEmpty(t, selefraBlock.GetNodeLocation("cli_version._key").ReadSourceString())
	assert.NotEmpty(t, selefraBlock.GetNodeLocation("cli_version._value").ReadSourceString())

	// ------------------------------------------------- --------------------------------------------------------------------

	cloudBlock := selefraBlock.CloudBlock
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("._key").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, cloudBlock.GetNodeLocation("project").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("project._key").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("project._value").ReadSourceString())

	assert.NotEmpty(t, cloudBlock.GetNodeLocation("organization").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("organization._key").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("organization._value").ReadSourceString())

	assert.NotEmpty(t, cloudBlock.GetNodeLocation("hostname").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("hostname._key").ReadSourceString())
	assert.NotEmpty(t, cloudBlock.GetNodeLocation("hostname._value").ReadSourceString())

	// ------------------------------------------------- --------------------------------------------------------------------

	connectionBlock := selefraBlock.ConnectionBlock
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("type").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("type._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("type._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("username").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("username._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("username._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("password").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("password._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("password._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("host").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("host._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("host._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("port").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("port._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("port._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("database").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("database._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("database._value").ReadSourceString())

	assert.NotEmpty(t, connectionBlock.GetNodeLocation("sslmode").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("sslmode._key").ReadSourceString())
	assert.NotEmpty(t, connectionBlock.GetNodeLocation("sslmode._value").ReadSourceString())

	// ------------------------------------------------- --------------------------------------------------------------------

	requireProvidersBlock := selefraBlock.RequireProvidersBlock
	for _, requireProviderBlock := range requireProvidersBlock {

		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("._value").ReadSourceString())

		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("name").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("name._key").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("name._value").ReadSourceString())

		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("source").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("source._key").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("source._value").ReadSourceString())

		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("version").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("version._key").ReadSourceString())
		assert.NotEmpty(t, requireProviderBlock.GetNodeLocation("version._value").ReadSourceString())

	}

	// ------------------------------------------------- --------------------------------------------------------------------

}
