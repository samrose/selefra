package module

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
)

// ------------------------------------------------- --------------------------------------------------------------------

type ModulesBlock []*ModuleBlock

var _ Block = (*ModulesBlock)(nil)
var _ MergableBlock[ModulesBlock] = (*ModulesBlock)(nil)

func (x ModulesBlock) Merge(other ModulesBlock) (ModulesBlock, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	moduleNameSet := make(map[string]struct{})
	mergedModules := make(ModulesBlock, 0)

	// merge myself
	for _, moduleBlock := range x {
		if _, exists := moduleNameSet[moduleBlock.Name]; exists {
			errorTips := fmt.Sprintf("Module with the same name is not allowed in the same module. The module name %s is the duplication", moduleBlock.Name)
			report := RenderErrorTemplate(errorTips, moduleBlock.GetNodeLocation(""))
			diagnostics.AddErrorMsg(report)
			continue
		}
		mergedModules = append(mergedModules, moduleBlock)
		moduleNameSet[moduleBlock.Name] = struct{}{}
	}

	// merge other
	for _, moduleBlock := range other {
		if _, exists := moduleNameSet[moduleBlock.Name]; exists {
			errorTips := fmt.Sprintf("Module with the same name is not allowed in the same module. The module name %s is the duplication", moduleBlock.Name)
			report := RenderErrorTemplate(errorTips, moduleBlock.GetNodeLocation(""))
			diagnostics.AddErrorMsg(report)
			continue
		}
		mergedModules = append(mergedModules, moduleBlock)
		moduleNameSet[moduleBlock.Name] = struct{}{}
	}

	return mergedModules, diagnostics
}

func (x ModulesBlock) Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {
	diagnostics := schema.NewDiagnostics()
	for _, moduleBlock := range x {
		diagnostics.AddDiagnostics(moduleBlock.Check(module, validatorContext))
	}
	return diagnostics
}

func (x ModulesBlock) IsEmpty() bool {
	return len(x) == 0
}

func (x ModulesBlock) GetNodeLocation(selector string) *NodeLocation {
	panic("not supported")
}

func (x ModulesBlock) SetNodeLocation(selector string, nodeLocation *NodeLocation) error {
	panic("not supported")
}

func (x ModulesBlock) ModulesInputMap() map[string]*ModuleBlock {
	modulesInputMap := make(map[string]*ModuleBlock)
	for _, subModuleBlock := range x {
		for _, uses := range subModuleBlock.Uses {
			modulesInputMap[uses] = subModuleBlock
		}
	}
	return modulesInputMap
}

// ------------------------------------------------- --------------------------------------------------------------------

type Filter struct {
	Name     string `yaml:"name" json:"name"`
	Severity string `yaml:"severity" json:"severity"`
	Provider string `yaml:"provider" json:"provider"`
}

// ModuleBlock Used to represent a common element in the modules array
type ModuleBlock struct {

	// Module name
	Name string `yaml:"name" json:"name"`

	// What other modules are referenced by this module
	Uses []string `yaml:"uses" json:"uses"`

	// The module supports specifying some filters
	Filter []Filter `yaml:"filter" json:"filter"`

	// The module supports specifying some variables
	Input map[string]any `yaml:"input" json:"input"`

	*LocatableImpl `yaml:"-"`
}

var _ Block = &ModuleBlock{}

func NewModuleBlock() *ModuleBlock {
	return &ModuleBlock{
		LocatableImpl: NewLocatableImpl(),
	}
}

func (x *ModuleBlock) Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	if x.Name == "" {
		errorTips := fmt.Sprintf("Module name must not be empty")
		report := RenderErrorTemplate(errorTips, x.GetNodeLocation("name"))
		diagnostics.AddErrorMsg(report)
	}

	if len(x.Uses) == 0 {
		errorTips := fmt.Sprintf("Module uses must not be empty")
		report := RenderErrorTemplate(errorTips, x.GetNodeLocation("uses"))
		diagnostics.AddErrorMsg(report)
	}

	if len(x.Input) != 0 {
		diagnostics.AddDiagnostics(x.checkInput(module, validatorContext))
	}

	return diagnostics
}

func (x *ModuleBlock) checkInput(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {
	// nothing to do now
	return nil
}

func (x *ModuleBlock) IsEmpty() bool {
	if x == nil {
		return true
	}
	return x.Name == "" && len(x.Uses) == 0 && len(x.Input) == 0
}

// ------------------------------------------------- --------------------------------------------------------------------
