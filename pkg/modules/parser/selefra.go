package parser

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/utils"
	"gopkg.in/yaml.v3"
)

// ------------------------------------------------- --------------------------------------------------------------------

const (
	SelefraBlockFieldName             = "selefra"
	SelefraBlockNameFieldName         = "name"
	SelefraBlockCLIVersionFieldName   = "cli_version"
	SelefraBlockLogLevelFieldName     = "log_level"
	SelefraRequiredProvidersBlockName = "providers"
	SelefraConnectionsBlockName       = "connection"
	SelefraCloudBlockName             = "cloud"
)

func (x *YamlFileToModuleParser) parseSelefraBlock(selefraBlockKeyNode, selefraBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) *module.SelefraBlock {

	blockPath := SelefraBlockFieldName

	// type check
	if selefraBlockValueNode.Kind != yaml.MappingNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForMappingType(selefraBlockValueNode, blockPath))
		return nil
	}

	toMap, d := x.toMap(selefraBlockValueNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	selefraBlock := module.NewSelefraBlock()
	for key, entry := range toMap {
		switch key {

		case SelefraBlockNameFieldName:
			selefraBlock.Name = x.parseStringValueWithDiagnosticsAndSetLocation(selefraBlock, SelefraBlockNameFieldName, entry, blockPath, diagnostics)

		case SelefraBlockCLIVersionFieldName:
			selefraBlock.CliVersion = x.parseStringValueWithDiagnosticsAndSetLocation(selefraBlock, SelefraBlockCLIVersionFieldName, entry, blockPath, diagnostics)

		case SelefraBlockLogLevelFieldName:
			selefraBlock.LogLevel = x.parseStringValueWithDiagnosticsAndSetLocation(selefraBlock, SelefraBlockLogLevelFieldName, entry, blockPath, diagnostics)

		case SelefraCloudBlockName:
			selefraBlock.CloudBlock = x.parseCloudBlock(entry.key, entry.value, diagnostics)

		case SelefraRequiredProvidersBlockName:
			selefraBlock.RequireProvidersBlock = x.parseRequiredProvidersBlock(entry.key, entry.value, diagnostics)

		case SelefraConnectionsBlockName:
			selefraBlock.ConnectionBlock = x.parseConnectionBlock(entry.key, entry.value, diagnostics)

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))
		}
	}

	if selefraBlock.IsEmpty() {
		return nil
	}

	// set code location
	x.setLocationKVWithDiagnostics(selefraBlock, "", blockPath, newNodeEntry(selefraBlockKeyNode, selefraBlockValueNode), diagnostics)

	return selefraBlock
}

// ------------------------------------------------- --------------------------------------------------------------------

const (
	CloudBlockProjectFieldName      = "project"
	CloudBlockOrganizationFieldName = "organization"
	CloudBlockHostnameFieldName     = "hostname"
)

func (x *YamlFileToModuleParser) parseCloudBlock(cloudBlockKeyNode, cloudBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) *module.CloudBlock {

	blockPath := fmt.Sprintf("%s.%s", SelefraBlockFieldName, "cloud")

	// type check
	toMap, d := x.toMap(cloudBlockValueNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	cloudBlock := module.NewCloudBlock()
	for key, entry := range toMap {
		switch key {
		case CloudBlockProjectFieldName:
			cloudBlock.Project = x.parseStringValueWithDiagnosticsAndSetLocation(cloudBlock, CloudBlockProjectFieldName, entry, blockPath, diagnostics)

		case CloudBlockOrganizationFieldName:
			cloudBlock.Organization = x.parseStringValueWithDiagnosticsAndSetLocation(cloudBlock, CloudBlockOrganizationFieldName, entry, blockPath, diagnostics)

		case CloudBlockHostnameFieldName:
			cloudBlock.HostName = x.parseStringValueWithDiagnosticsAndSetLocation(cloudBlock, CloudBlockHostnameFieldName, entry, blockPath, diagnostics)

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))

		}
	}

	if cloudBlock.IsEmpty() {
		return nil
	}

	// set code location
	x.setLocationKVWithDiagnostics(cloudBlock, "", blockPath, newNodeEntry(cloudBlockKeyNode, cloudBlockValueNode), diagnostics)

	return cloudBlock
}

// ------------------------------------------------- --------------------------------------------------------------------

const (
	RequiredProviderBlockNameFieldName    = "name"
	RequiredProviderBlockSourceFieldName  = "source"
	RequiredProviderBlockVersionFieldName = "version"
	RequiredProviderBlockPathFieldName    = "path"
)

func (x *YamlFileToModuleParser) parseRequiredProvidersBlock(requiredProviderBlockKeyNode, requiredProviderBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) module.RequireProvidersBlock {

	blockPath := fmt.Sprintf("%s.%s", SelefraBlockFieldName, SelefraRequiredProvidersBlockName)

	if requiredProviderBlockValueNode.Kind != yaml.SequenceNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForArrayType(requiredProviderBlockKeyNode, blockPath))
		return nil
	}

	requiredProvidersBlock := make(module.RequireProvidersBlock, 0)
	for index, requiredProviderNode := range requiredProviderBlockValueNode.Content {
		p := x.parseRequiredProviderBlock(index, requiredProviderNode, diagnostics)
		if p != nil {
			requiredProvidersBlock = append(requiredProvidersBlock, p)
		}
	}

	if len(requiredProvidersBlock) == 0 {
		return nil
	}
	return requiredProvidersBlock
}

func (x *YamlFileToModuleParser) parseRequiredProviderBlock(index int, node *yaml.Node, diagnostics *schema.Diagnostics) *module.RequireProviderBlock {

	blockPath := fmt.Sprintf("%s.%s[%d]", SelefraBlockFieldName, SelefraRequiredProvidersBlockName, index)

	toMap, d := x.toMap(node, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	requiredProviderBlock := module.NewRequireProviderBlock()
	for key, entry := range toMap {
		switch key {

		case RequiredProviderBlockNameFieldName:
			requiredProviderBlock.Name = x.parseStringValueWithDiagnosticsAndSetLocation(requiredProviderBlock, RequiredProviderBlockNameFieldName, entry, blockPath, diagnostics)

		case RequiredProviderBlockSourceFieldName:
			requiredProviderBlock.Source = x.parseStringValueWithDiagnosticsAndSetLocation(requiredProviderBlock, RequiredProviderBlockSourceFieldName, entry, blockPath, diagnostics)

		case RequiredProviderBlockVersionFieldName:
			requiredProviderBlock.Version = x.parseStringValueWithDiagnosticsAndSetLocation(requiredProviderBlock, RequiredProviderBlockVersionFieldName, entry, blockPath, diagnostics)

		case RequiredProviderBlockPathFieldName:
			requiredProviderBlock.Path = x.parseStringValueWithDiagnosticsAndSetLocation(requiredProviderBlock, RequiredProviderBlockPathFieldName, entry, blockPath, diagnostics)

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))

		}
	}

	if requiredProviderBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(requiredProviderBlock, "", blockPath, newNodeEntry(nil, node), diagnostics)

	return requiredProviderBlock
}

// ------------------------------------------------- --------------------------------------------------------------------

const (
	ConnectionBlockTypeFieldName     = "type"
	ConnectionBlockUsernameFieldName = "username"
	ConnectionBlockPasswordFieldName = "password"
	ConnectionBlockHostFieldName     = "host"
	ConnectionBlockPortFieldName     = "port"
	ConnectionBlockDatabaseFieldName = "database"
	ConnectionBlockSSLModeFieldName  = "sslmode"
	ConnectionBlockExtrasFieldName   = "extras"
)

func (x *YamlFileToModuleParser) parseConnectionBlock(connectionBlockKeyNode, connectionBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) *module.ConnectionBlock {

	blockPath := fmt.Sprintf("%s.%s", SelefraBlockFieldName, SelefraConnectionsBlockName)

	// type check
	toMap, d := x.toMap(connectionBlockValueNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	connectionBlock := module.NewConnectionBlock()
	for key, entry := range toMap {
		switch key {

		case ConnectionBlockTypeFieldName:
			connectionBlock.Type = x.parseStringValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockTypeFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockUsernameFieldName:
			connectionBlock.Username = x.parseStringValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockUsernameFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockPasswordFieldName:
			connectionBlock.Password = x.parseStringValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockPasswordFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockHostFieldName:
			connectionBlock.Host = x.parseStringValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockHostFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockPortFieldName:
			connectionBlock.Port = x.parseUintValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockPortFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockDatabaseFieldName:
			connectionBlock.Database = x.parseStringValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockDatabaseFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockSSLModeFieldName:
			connectionBlock.SSLMode = x.parseStringValueWithDiagnosticsAndSetLocation(connectionBlock, ConnectionBlockSSLModeFieldName, entry, blockPath, diagnostics)

		case ConnectionBlockExtrasFieldName:
			connectionBlock.Extras = x.parseStringSliceAndSetLocation(connectionBlock, ConnectionBlockExtrasFieldName, newNodeEntry(nil, entry.value), blockPath, diagnostics)
		}
	}

	if connectionBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(connectionBlock, "", blockPath, newNodeEntry(connectionBlockKeyNode, connectionBlockValueNode), diagnostics)

	return connectionBlock
}

// ------------------------------------------------- --------------------------------------------------------------------
