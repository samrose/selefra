package parser

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/songzhibin97/gkit/tools/pointer"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

// YamlFileToModuleParser Read a yaml file as a module, but the module is only for program convenience. There is no such file module; a module should at least be a folder
type YamlFileToModuleParser struct {
	yamlFilePath string
	instruction  map[string]interface{}
}

func NewYamlFileToModuleParser(yamlFilePath string, instruction map[string]interface{}) *YamlFileToModuleParser {
	return &YamlFileToModuleParser{
		yamlFilePath: yamlFilePath,
		instruction:  instruction,
	}
}

func (x *YamlFileToModuleParser) Parse() (*module.Module, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// 1. read yaml file
	yamlFileBytes, err := os.ReadFile(x.yamlFilePath)
	if err != nil {
		return nil, diagnostics.AddErrorMsg("YamlParserError, read yaml file %s error: %s", x.yamlFilePath, err.Error())
	}

	documentNode := &yaml.Node{}
	err = yaml.Unmarshal(yamlFileBytes, documentNode)
	if err != nil {
		return nil, diagnostics.AddErrorMsg("yaml file %s unmarshal error: %s", x.yamlFilePath, err.Error())
	}

	if documentNode.Kind != yaml.DocumentNode {
		return nil, diagnostics.AddErrorMsg("yaml file %s unmarshal error, not have document node", x.yamlFilePath)
	}

	yamlFileModule := &module.Module{}
	rootContent := documentNode.Content[0].Content
	for index := 0; index < len(rootContent); index += 2 {
		key := rootContent[index]
		value := rootContent[index+1]
		switch key.Value {
		case SelefraBlockFieldName:
			yamlFileModule.SelefraBlock = x.parseSelefraBlock(key, value, diagnostics)
			if x.instruction != nil {

				if x.instruction["openai_api_key"] != "" {
					yamlFileModule.SelefraBlock.OpenaiApiKey = x.instruction["openai_api_key"].(string)
				}
				if x.instruction["openai_mode"] != "" {
					yamlFileModule.SelefraBlock.OpenaiMode = x.instruction["openai_mode"].(string)
				}
				if x.instruction["openai_limit"] != 0 {
					yamlFileModule.SelefraBlock.OpenaiLimit = x.instruction["openai_limit"].(uint64)
				}
			}
		case VariablesBlockName:
			yamlFileModule.VariablesBlock = x.parseVariablesBlock(key, value, diagnostics)
		case ProvidersBlockName:
			yamlFileModule.ProvidersBlock = x.parseProvidersBlock(key, value, diagnostics)
		case ModulesBlockName:
			yamlFileModule.ModulesBlock = x.parseModulesBlock(key, value, diagnostics)
		case RulesBlockName:
			yamlFileModule.RulesBlock = x.parseRulesBlock(key, value, diagnostics)
		}

		if x.instruction != nil {
			if x.instruction["query"] != nil {
				yamlFileModule.RulesBlock = module.RulesBlock{
					&module.RuleBlock{
						Name:   "this it gpt",
						Query:  x.instruction["query"].(string),
						Output: "this is gpt,title is:{{.title}}\ndescription is: {{.description}}\nremediation is: {{.remediation}}\nseverity is:{{.severity}}\ntags is: {{.tags}}\n",
					},
				}
			}
		}
	}
	return yamlFileModule, diagnostics
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *YamlFileToModuleParser) parseUintValueWithDiagnosticsAndSetLocation(block module.Block, fieldName string, entry *nodeEntry, blockBasePath string, diagnostics *schema.Diagnostics) *uint64 {
	valueInteger := x.parseUintWithDiagnostics(entry.value, blockBasePath+"."+fieldName, diagnostics)

	if entry.key != nil {
		x.setLocationWithDiagnostics(block, fieldName+module.NodeLocationSelfKey, blockBasePath, entry.key, diagnostics)
	}

	x.setLocationWithDiagnostics(block, fieldName+module.NodeLocationSelfValue, blockBasePath, entry.value, diagnostics)

	return valueInteger
}

func (x *YamlFileToModuleParser) parseStringValueWithDiagnosticsAndSetLocation(block module.Block, fieldName string, entry *nodeEntry, blockBasePath string, diagnostics *schema.Diagnostics) string {
	valueString := x.parseStringWithDiagnostics(entry.value, blockBasePath+"."+fieldName, diagnostics)

	if entry.key != nil {
		x.setLocationWithDiagnostics(block, fieldName+module.NodeLocationSelfKey, blockBasePath, entry.key, diagnostics)
	}

	x.setLocationWithDiagnostics(block, fieldName+module.NodeLocationSelfValue, blockBasePath, entry.value, diagnostics)

	return valueString
}

// Parse node as a string slice
func (x *YamlFileToModuleParser) parseStringSlice(node *yaml.Node, blockPath string) ([]string, *schema.Diagnostics) {
	diagnostics := schema.NewDiagnostics()

	// modules must be an array element
	if node.Kind != yaml.SequenceNode {
		return nil, x.buildNodeErrorMsgForArrayType(node, blockPath)
	}

	elementSlice := make([]string, 0)
	for index, elementNode := range node.Content {
		useNodeValue := x.parseStringWithDiagnostics(elementNode, fmt.Sprintf("%s[%d]", blockPath, index), diagnostics)
		if useNodeValue != "" {
			elementSlice = append(elementSlice, useNodeValue)
		}
	}
	return elementSlice, diagnostics
}

func (x *YamlFileToModuleParser) parseStringSliceAndSetLocation(block module.Block, fieldName string, entry *nodeEntry, blockBasePath string, diagnostics *schema.Diagnostics) []string {

	blockPath := blockBasePath + "." + fieldName

	elementSlice := make([]string, 0)
	switch entry.value.Kind {
	case yaml.SequenceNode:
		for index, elementNode := range entry.value.Content {
			elementFullPath := fmt.Sprintf("%s[%d]", blockPath, index)
			useNodeValue := x.parseStringWithDiagnostics(elementNode, elementFullPath, diagnostics)
			if useNodeValue != "" {

				elementSlice = append(elementSlice, useNodeValue)

				relativePath := fmt.Sprintf("%s[%d]%s", fieldName, index, module.NodeLocationSelfValue)
				err := block.SetNodeLocation(relativePath, module.BuildLocationFromYamlNode(x.yamlFilePath, elementFullPath, elementNode))
				if err != nil {
					diagnostics.AddErrorMsg("file = %s, set location %s error: %s", x.yamlFilePath, elementFullPath, err.Error())
				}
			}
		}
		if len(elementSlice) == 0 {
			return nil
		}
	case yaml.ScalarNode:
		index := 0
		elementNode := entry.value
		elementFullPath := fmt.Sprintf("%s[%d]", blockPath, index)
		useNodeValue := x.parseStringWithDiagnostics(entry.value, elementFullPath, diagnostics)
		if useNodeValue != "" {

			elementSlice = append(elementSlice, useNodeValue)

			relativePath := fmt.Sprintf("%s[%d]%s", fieldName, index, module.NodeLocationSelfValue)
			err := block.SetNodeLocation(relativePath, module.BuildLocationFromYamlNode(x.yamlFilePath, elementFullPath, elementNode))
			if err != nil {
				diagnostics.AddErrorMsg("file = %s, set location %s error: %s", x.yamlFilePath, elementFullPath, err.Error())
			}
		}
	default:
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForArrayType(entry.key, blockPath))
	}

	// set self location
	x.setLocationKVWithDiagnostics(block, fieldName, blockPath, entry, diagnostics)

	return elementSlice
}

func (x *YamlFileToModuleParser) parseStringMapAndSetLocation(block module.Block, fieldName string, entry *nodeEntry, blockBasePath string, diagnostics *schema.Diagnostics) map[string]string {

	blockPath := blockBasePath + "." + fieldName

	// modules must be an array element
	if entry.value.Kind != yaml.MappingNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForMappingType(entry.key, blockPath))
		return nil
	}

	toMap, d := x.toMap(entry.value, blockPath)
	diagnostics.AddDiagnostics(d)
	if utils.HasError(d) {
		return nil
	}

	m := make(map[string]string, 0)
	for key, entry := range toMap {
		if entry.value.Kind != yaml.ScalarNode {
			diagnostics.AddDiagnostics(x.buildNodeErrorMsgForScalarType(entry.key, blockPath, "string"))
			continue
		}

		m[key] = x.parseStringValueWithDiagnosticsAndSetLocation(block, fieldName+"."+key, entry, blockBasePath, diagnostics)
	}

	x.setLocationKVWithDiagnostics(block, fieldName, blockPath, entry, diagnostics)

	return m
}

func (x *YamlFileToModuleParser) parseAny(node *yaml.Node, blockPath string) (any, *schema.Diagnostics) {
	keyName := "any-key"
	handlerNode := yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: keyName,
			},
			node,
		},
	}
	out, err := yaml.Marshal(handlerNode)
	if err != nil {
		// TODO
		return nil, schema.NewDiagnostics().AddErrorMsg(err.Error())
	}
	var r map[string]any
	err = yaml.Unmarshal(out, &r)
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg(err.Error())
	}
	return r[keyName], nil
}

func (x *YamlFileToModuleParser) parseUintWithDiagnostics(node *yaml.Node, blockPath string, diagnostics *schema.Diagnostics) *uint64 {
	if node.Kind != yaml.ScalarNode {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForScalarType(node, blockPath, "int"))
		return nil
	}
	intValue, err := strconv.Atoi(strings.TrimSpace(node.Value))
	if err != nil {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForScalarType(node, blockPath, "int"))
		return nil
	}
	if intValue < 0 {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForScalarType(node, blockPath, "uint"))
		return nil
	}
	return pointer.ToUint64Pointer(uint64(intValue))
}

func (x *YamlFileToModuleParser) parseStringWithDiagnostics(node *yaml.Node, blockPath string, diagnostics *schema.Diagnostics) string {
	if node.Kind == yaml.ScalarNode {
		return node.Value
	} else {
		diagnostics.AddDiagnostics(x.buildNodeErrorMsgForScalarType(node, blockPath, "string"))
		return ""
	}
}

type nodeEntry struct {
	key, value *yaml.Node
}

func newNodeEntry(key, value *yaml.Node) *nodeEntry {
	return &nodeEntry{
		key:   key,
		value: value,
	}
}

func (x *YamlFileToModuleParser) toMap(node *yaml.Node, blockPath string) (map[string]*nodeEntry, *schema.Diagnostics) {

	// check node type, must is mapping type
	if node.Kind != yaml.MappingNode {
		return nil, x.buildNodeErrorMsgForMappingType(node, blockPath)
	}

	// convert to map
	m := make(map[string]*nodeEntry, 0)
	for index := 0; index < len(node.Content); index += 2 {
		key := node.Content[index]

		// key must is string type
		if key.Kind != yaml.ScalarNode {
			return nil, x.buildNodeErrorMsgForScalarType(key, fmt.Sprintf("%s.%s", blockPath, key.Value), "string")
		}

		value := node.Content[index+1]
		m[key.Value] = &nodeEntry{
			key:   key,
			value: value,
		}
	}
	return m, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *YamlFileToModuleParser) setLocationKVWithDiagnostics(block module.Block, relativeYamlSelectorPath, fullYamlSelectorPath string, nodeEntry *nodeEntry, diagnostics *schema.Diagnostics) {

	if nodeEntry.key != nil {
		x.setLocationWithDiagnostics(block, relativeYamlSelectorPath+module.NodeLocationSelfKey, fullYamlSelectorPath, nodeEntry.key, diagnostics)
	}

	if nodeEntry.value != nil {
		x.setLocationWithDiagnostics(block, relativeYamlSelectorPath+module.NodeLocationSelfValue, fullYamlSelectorPath, nodeEntry.value, diagnostics)
	}
}

func (x *YamlFileToModuleParser) setLocationWithDiagnostics(block module.Block, relativeYamlSelectorPath, fullYamlSelectorPath string, node *yaml.Node, diagnostics *schema.Diagnostics) {
	location := module.BuildLocationFromYamlNode(x.yamlFilePath, fullYamlSelectorPath, node)
	err := block.SetNodeLocation(relativeYamlSelectorPath, location)
	if err != nil {
		diagnostics.AddErrorMsg("YamlFileToModuleParser error, build location for file %s %s error: %s", x.yamlFilePath, fullYamlSelectorPath, err.Error())
	}
}

func (x *YamlFileToModuleParser) buildNodeErrorMsgForUnSupport(keyNode, valueNode *yaml.Node, blockPath string) *schema.Diagnostics {
	keyLocation := module.BuildLocationFromYamlNode(x.yamlFilePath, blockPath, keyNode)
	valueLocation := module.BuildLocationFromYamlNode(x.yamlFilePath, blockPath, valueNode)
	location := module.MergeKeyValueLocation(keyLocation, valueLocation)
	location.YamlSelector = keyLocation.YamlSelector
	errorMsg := fmt.Sprintf("syntax error, do not support %s", blockPath)
	report := module.RenderErrorTemplate(errorMsg, location)
	return schema.NewDiagnostics().AddErrorMsg(report)
}

func (x *YamlFileToModuleParser) buildNodeErrorMsgForScalarType(node *yaml.Node, blockPath string, scalarTypeName string) *schema.Diagnostics {
	return x.buildNodeErrorMsg(blockPath, node, fmt.Sprintf("syntax error, %s must is a %s type", blockPath, scalarTypeName))
}

func (x *YamlFileToModuleParser) buildNodeErrorMsgForMappingType(node *yaml.Node, blockPath string) *schema.Diagnostics {
	return x.buildNodeErrorMsg(blockPath, node, fmt.Sprintf("syntax error, %s block must is a mapping type", blockPath))
}

func (x *YamlFileToModuleParser) buildNodeErrorMsgForArrayType(node *yaml.Node, blockPath string) *schema.Diagnostics {
	return x.buildNodeErrorMsg(blockPath, node, fmt.Sprintf("syntax error, %s block must is a array type", blockPath))
}

func (x *YamlFileToModuleParser) buildNodeErrorMsg(blockPath string, node *yaml.Node, errorMessage string) *schema.Diagnostics {
	location := module.BuildLocationFromYamlNode(x.yamlFilePath, blockPath, node)
	report := module.RenderErrorTemplate(errorMessage, location)
	return schema.NewDiagnostics().AddErrorMsg(report)
}

// ------------------------------------------------- --------------------------------------------------------------------
