package parser

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/utils"
	"gopkg.in/yaml.v3"
)

// ------------------------------------------------ ---------------------------------------------------------------------

const (
	ProvidersBlockName = "providers"
)

func (x *YamlFileToModuleParser) parseProvidersBlock(providersBlockKeyNode, providersBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) module.ProvidersBlock {

	// modules must be an array element
	if providersBlockValueNode.Kind != yaml.SequenceNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForArrayType(providersBlockKeyNode, ProvidersBlockName))
		return nil
	}

	// Parse each child element
	modulesBlock := make(module.ProvidersBlock, 0)
	for index, moduleNode := range providersBlockValueNode.Content {
		block := x.parseProviderBlock(index, moduleNode, diagnostics)
		if block != nil {
			modulesBlock = append(modulesBlock, block)
		}
	}
	return modulesBlock
}

// ------------------------------------------------ ---------------------------------------------------------------------

const (
	ProviderBlockNameFieldName          = "name"
	ProviderBlockCacheFieldName         = "cache"
	ProviderBlockProviderFieldName      = "provider"
	ProviderBlockMaxGoroutinesFieldName = "max_goroutines"
	ProviderBlockResourcesFieldName     = "resources"

	// ProviderBlockProvidersConfigYamlStringFieldName Virtual field
	//ProviderBlockProvidersConfigYamlStringFieldName = ""
)

func (x *YamlFileToModuleParser) parseProviderBlock(index int, providerBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) *module.ProviderBlock {

	blockPath := fmt.Sprintf("%s[%d]", ProvidersBlockName, index)

	toMap, d := x.toMap(providerBlockValueNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	providerBlock := module.NewProviderBlock()

	// name
	entry, exists := toMap[ProviderBlockNameFieldName]
	if exists {
		providerBlock.Name = x.parseStringValueWithDiagnosticsAndSetLocation(providerBlock, ProviderBlockNameFieldName, entry, blockPath, diagnostics)
	}

	// cache
	entry, exists = toMap[ProviderBlockCacheFieldName]
	if exists {
		providerBlock.Cache = x.parseStringValueWithDiagnosticsAndSetLocation(providerBlock, ProviderBlockCacheFieldName, entry, blockPath, diagnostics)
	}

	// provider
	entry, exists = toMap[ProviderBlockProviderFieldName]
	if exists {
		providerBlock.Provider = x.parseStringValueWithDiagnosticsAndSetLocation(providerBlock, ProviderBlockProviderFieldName, entry, blockPath, diagnostics)
	}

	// max_goroutines
	entry, exists = toMap[ProviderBlockMaxGoroutinesFieldName]
	if exists {
		providerBlock.MaxGoroutines = x.parseUintValueWithDiagnosticsAndSetLocation(providerBlock, ProviderBlockMaxGoroutinesFieldName, entry, blockPath, diagnostics)
	}

	// resources
	entry, exists = toMap[ProviderBlockResourcesFieldName]
	if exists {
		providerBlock.Resources = x.parseStringSliceAndSetLocation(providerBlock, ProviderBlockResourcesFieldName, entry, blockPath, diagnostics)
	}

	out, err := yaml.Marshal(providerBlockValueNode)
	if err != nil {
		// TODO build error message
		diagnostics.AddErrorMsg("build error message")
		return nil
	}
	providerBlock.ProvidersConfigYamlString = string(out)

	if providerBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(providerBlock, "", blockPath, newNodeEntry(nil, providerBlockValueNode), diagnostics)

	return providerBlock
}

// ------------------------------------------------ ---------------------------------------------------------------------
