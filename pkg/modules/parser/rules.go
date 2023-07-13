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
	RulesBlockName = "rules"
)

func (x *YamlFileToModuleParser) parseRulesBlock(rulesBlockKeyNode, rulesBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) module.RulesBlock {
	if x.instruction != nil && x.instruction["query"] == "" {
		return nil
	}
	// modules must be an array element
	if rulesBlockValueNode.Kind != yaml.SequenceNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForArrayType(rulesBlockValueNode, RulesBlockName))
		return nil
	}

	// Parse each child element
	rulesBlock := make(module.RulesBlock, 0)
	for index, moduleNode := range rulesBlockValueNode.Content {
		block := x.parseRuleBlock(index, moduleNode, diagnostics)
		if block != nil {
			rulesBlock = append(rulesBlock, block)
		}
	}

	return rulesBlock
}

// ------------------------------------------------ ---------------------------------------------------------------------

const (
	RuleBlockNameFieldName      = "name"
	RuleBlockQueryFieldName     = "query"
	RuleBlockLabelsFieldName    = "labels"
	RuleBlockMetadataFieldName  = "metadata"
	RuleBlockMainTableFieldName = "main_table"
	RuleBlockOutputFieldName    = "output"
)

func (x *YamlFileToModuleParser) parseRuleBlock(index int, ruleBlockNode *yaml.Node, diagnostics *schema.Diagnostics) *module.RuleBlock {

	blockPath := fmt.Sprintf("%s[%d]", RulesBlockName, index)

	toMap, d := x.toMap(ruleBlockNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	ruleBlock := module.NewRuleBlock()
	for key, entry := range toMap {
		switch key {

		case RuleBlockNameFieldName:
			ruleBlock.Name = x.parseStringValueWithDiagnosticsAndSetLocation(ruleBlock, RuleBlockNameFieldName, entry, blockPath, diagnostics)

		case RuleBlockQueryFieldName:
			ruleBlock.Query = x.parseStringValueWithDiagnosticsAndSetLocation(ruleBlock, RuleBlockQueryFieldName, entry, blockPath, diagnostics)

		case RuleBlockLabelsFieldName:
			ruleBlock.Labels = x.parseStringMapAndSetLocation(ruleBlock, RuleBlockLabelsFieldName, entry, blockPath, diagnostics)

		case RuleBlockMetadataFieldName:
			ruleBlock.MetadataBlock = x.parseMetadataBlock(index, ruleBlock, entry.key, entry.value, diagnostics)

		case RuleBlockMainTableFieldName:
			ruleBlock.MainTable = x.parseStringValueWithDiagnosticsAndSetLocation(ruleBlock, RuleBlockMainTableFieldName, entry, blockPath, diagnostics)

		case RuleBlockOutputFieldName:
			ruleBlock.Output = x.parseStringValueWithDiagnosticsAndSetLocation(ruleBlock, RuleBlockOutputFieldName, entry, blockPath, diagnostics)

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))
		}
	}

	if ruleBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(ruleBlock, "", blockPath, newNodeEntry(nil, ruleBlockNode), diagnostics)

	return ruleBlock
}

// ------------------------------------------------ ---------------------------------------------------------------------

const (
	RuleMetadataBlockName                 = "metadata"
	RuleMetadataBlockIdFieldName          = "id"
	RuleMetadataBlockSeverityFieldName    = "severity"
	RuleMetadataBlockProviderFieldName    = "provider"
	RuleMetadataBlockTagsFieldName        = "tags"
	RuleMetadataBlockAuthorFieldName      = "author"
	RuleMetadataBlockRemediationFieldName = "remediation"
	RuleMetadataBlockTitleFieldName       = "title"
	RuleMetadataBlockDescriptionFieldName = "description"
)

func (x *YamlFileToModuleParser) parseMetadataBlock(ruleIndex int, ruleBlock *module.RuleBlock, metadataBlockKeyNode, metadataBlockValueNode *yaml.Node, diagnostics *schema.Diagnostics) *module.RuleMetadataBlock {

	blockPath := fmt.Sprintf("%s[%d].%s", RulesBlockName, ruleIndex, RuleMetadataBlockName)

	toMap, d := x.toMap(metadataBlockValueNode, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	ruleMetadataBlock := module.NewRuleMetadataBlock(ruleBlock)
	for key, entry := range toMap {
		switch key {

		case RuleMetadataBlockIdFieldName:
			ruleMetadataBlock.Id = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockIdFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockSeverityFieldName:
			ruleMetadataBlock.Severity = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockSeverityFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockProviderFieldName:
			ruleMetadataBlock.Provider = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockProviderFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockTagsFieldName:
			ruleMetadataBlock.Tags = x.parseStringSliceAndSetLocation(ruleMetadataBlock, RuleMetadataBlockTagsFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockAuthorFieldName:
			ruleMetadataBlock.Author = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockAuthorFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockRemediationFieldName:
			ruleMetadataBlock.Remediation = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockRemediationFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockTitleFieldName:
			ruleMetadataBlock.Title = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockTitleFieldName, entry, blockPath, diagnostics)

		case RuleMetadataBlockDescriptionFieldName:
			ruleMetadataBlock.Description = x.parseStringValueWithDiagnosticsAndSetLocation(ruleMetadataBlock, RuleMetadataBlockDescriptionFieldName, entry, blockPath, diagnostics)

		default:
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForUnSupport(entry.key, entry.value, fmt.Sprintf("%s.%s", blockPath, key)))
		}
	}

	if ruleMetadataBlock.IsEmpty() {
		return nil
	}

	// set location
	x.setLocationKVWithDiagnostics(ruleMetadataBlock, "", blockPath, newNodeEntry(metadataBlockKeyNode, metadataBlockValueNode), diagnostics)

	return ruleMetadataBlock
}

// ------------------------------------------------ ---------------------------------------------------------------------
