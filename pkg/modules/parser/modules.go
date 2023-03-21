package parser

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/reflect_util"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/utils"
	"gopkg.in/yaml.v3"
)

// ------------------------------------------------ ---------------------------------------------------------------------

const (
	ModulesBlockName = "modules"
)

// Parse modules block
func (x *YamlFileToModuleParser) parseModulesBlock(moduleBlockKeyNode, moduleBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) module.ModulesBlock {

	// modules must be an array element
	if moduleBlockValueNode.Kind != yaml.SequenceNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForArrayType(moduleBlockValueNode, ModulesBlockName))
		return nil
	}

	// Parse each child element
	modulesBlock := make(module.ModulesBlock, 0)
	for index, moduleNode := range moduleBlockValueNode.Content {
		block := x.parseModuleBlock(index, moduleNode, diagnostics)
		if block != nil {
			modulesBlock = append(modulesBlock, block)
		}
	}

	return modulesBlock
}

// ------------------------------------------------ ---------------------------------------------------------------------

const (
	ModuleBlockNameFieldName  = "name"
	ModuleBlockUsesFieldName  = "uses"
	ModuleBlockInputFieldName = "input"
)

// Parse module block
func (x *YamlFileToModuleParser) parseModuleBlock(moduleIndex int, moduleBlockNode *yaml.Node, diagnostics *schema.Diagnostics) *module.ModuleBlock {

	blockPath := fmt.Sprintf("%s[%d]", ModulesBlockName, moduleIndex)

	toMap, d := x.toMap(moduleBlockNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	moduleBlock := module.NewModuleBlock()
	for key, entry := range toMap {
		switch key {
		case ModuleBlockNameFieldName:
			moduleBlock.Name = x.parseStringValueWithDiagnosticsAndSetLocation(moduleBlock, ModuleBlockNameFieldName, entry, blockPath, diagnostics)

		case ModuleBlockUsesFieldName:
			moduleBlock.Uses = x.parseStringSliceAndSetLocation(moduleBlock, ModuleBlockUsesFieldName, entry, blockPath, diagnostics)

		case ModuleBlockInputFieldName:
			inputMap := x.parseModuleInputBlock(moduleBlock, moduleIndex, entry.key, entry.value, diagnostics)
			if len(inputMap) != 0 {
				moduleBlock.Input = inputMap
			}

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))

		}
	}

	if moduleBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(moduleBlock, "", blockPath, newNodeEntry(nil, moduleBlockNode), diagnostics)

	return moduleBlock
}

func (x *YamlFileToModuleParser) parseModuleInputBlock(moduleBlock *module.ModuleBlock, index int, keyNode, valueNode *yaml.Node, diagnostics *schema.Diagnostics) map[string]any {

	blockPath := fmt.Sprintf("%s[%d].%s", ModulesBlockName, index, ModuleBlockInputFieldName)

	if valueNode.Kind != yaml.MappingNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForMappingType(valueNode, blockPath))
		return nil
	}

	toMap, d := x.toMap(valueNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	inputMap := make(map[string]any)
	for key, entry := range toMap {
		parseAny, d := x.parseAny(entry.value, fmt.Sprintf("%s.%s", blockPath, key))
		diagnostics.AddDiagnostics(d)
		if !reflect_util.IsNil(parseAny) {
			inputMap[key] = parseAny

			// set location
			x.setLocationKVWithDiagnostics(moduleBlock, ModuleBlockInputFieldName+"."+key, blockPath, entry, diagnostics)
		}
	}

	if len(inputMap) == 0 {
		return nil
	}

	x.setLocationKVWithDiagnostics(moduleBlock, ModuleBlockInputFieldName, blockPath, newNodeEntry(keyNode, valueNode), diagnostics)

	return inputMap
}

// ------------------------------------------------ ---------------------------------------------------------------------
