package parser

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/reflect_util"
	"github.com/selefra/selefra/pkg/modules/module"
	"gopkg.in/yaml.v3"
)

const VariablesBlockName = "variables"

func (x *YamlFileToModuleParser) parseVariablesBlock(variablesBlockKeyNode, variableBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) module.VariablesBlock {

	blockPath := VariablesBlockName

	// variables must be an array element
	if variableBlockValueNode.Kind != yaml.SequenceNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForArrayType(variableBlockValueNode, blockPath))
		return nil
	}

	// Parse each child element
	variables := make(module.VariablesBlock, 0)
	for index, variableNode := range variableBlockValueNode.Content {
		block := x.parseVariableBlock(index, variableNode, diagnostics)
		if block != nil {
			variables = append(variables, block)
		}
	}

	if len(variables) == 0 {
		return nil
	}
	return variables
}

const (
	VariableBlockKeyFieldName         = "key"
	VariableBlockDefaultFieldName     = "default"
	VariableBlockDescriptionFieldName = "description"
	VariableBlockAuthorFieldName      = "author"
)

func (x *YamlFileToModuleParser) parseVariableBlock(index int, node *yaml.Node, diagnostics *schema.Diagnostics) *module.VariableBlock {

	blockPath := fmt.Sprintf("%s[%d]", VariablesBlockName, index)

	toMap, d := x.toMap(node, blockPath)
	diagnostics.AddDiagnostics(d)
	if d != nil && d.HasError() {
		return nil
	}

	variableBlock := module.NewVariableBlock()
	for key, entry := range toMap {
		switch key {

		case VariableBlockKeyFieldName:
			variableBlock.Key = x.parseStringValueWithDiagnosticsAndSetLocation(variableBlock, VariableBlockKeyFieldName, entry, blockPath, diagnostics)

		case VariableBlockDefaultFieldName:
			fieldSelector := fmt.Sprintf("%s.%s", blockPath, VariableBlockDefaultFieldName)
			anyValue, d := x.parseAny(entry.value, fieldSelector)
			diagnostics.AddDiagnostics(d)
			if !reflect_util.IsNil(anyValue) {
				variableBlock.Default = anyValue
			}
			// set location
			x.setLocationKVWithDiagnostics(variableBlock, VariableBlockDefaultFieldName, fieldSelector, entry, diagnostics)

		case VariableBlockDescriptionFieldName:
			variableBlock.Description = x.parseStringValueWithDiagnosticsAndSetLocation(variableBlock, VariableBlockDescriptionFieldName, entry, blockPath, diagnostics)

		case VariableBlockAuthorFieldName:
			variableBlock.Author = x.parseStringValueWithDiagnosticsAndSetLocation(variableBlock, VariableBlockAuthorFieldName, entry, blockPath, diagnostics)

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))
		}
	}

	if variableBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(variableBlock, "", blockPath, newNodeEntry(nil, node), diagnostics)

	return variableBlock
}
