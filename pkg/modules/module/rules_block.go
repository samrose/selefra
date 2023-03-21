package module

import (
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/utils"
	"path/filepath"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

type RulesBlock []*RuleBlock

var _ MergableBlock[RulesBlock] = (*RulesBlock)(nil)
var _ Block = (*RulesBlock)(nil)

func (x RulesBlock) Merge(other RulesBlock) (RulesBlock, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	mergedRules := make(RulesBlock, 0)
	ruleNameSet := make(map[string]struct{}, 0)

	// merge self
	for _, ruleBlock := range x {
		if _, exists := ruleNameSet[ruleBlock.Name]; exists {
			errorTips := fmt.Sprintf("Rule with the same name is not allowed in the same module. The rule name %s is duplication", ruleBlock.Name)
			report := RenderErrorTemplate(errorTips, ruleBlock.GetNodeLocation(""))
			diagnostics.AddErrorMsg(report)
			continue
		}
		ruleNameSet[ruleBlock.Name] = struct{}{}
		mergedRules = append(mergedRules, ruleBlock)
	}

	// merge other
	for _, ruleBlock := range other {
		if _, exists := ruleNameSet[ruleBlock.Name]; exists {
			errorTips := fmt.Sprintf("Rule with the same name is not allowed in the same module. The rule name %s is duplication", ruleBlock.Name)
			report := RenderErrorTemplate(errorTips, ruleBlock.GetNodeLocation(""))
			diagnostics.AddErrorMsg(report)
			continue
		}
		ruleNameSet[ruleBlock.Name] = struct{}{}
		mergedRules = append(mergedRules, ruleBlock)
	}

	return mergedRules, diagnostics
}

func (x RulesBlock) Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {
	diagnostics := schema.NewDiagnostics()

	// Each block should be able to pass inspection
	for _, ruleBlock := range x {
		diagnostics.AddDiagnostics(ruleBlock.Check(module, validatorContext))
	}

	return diagnostics
}

func (x RulesBlock) IsEmpty() bool {
	return len(x) == 0
}

func (x RulesBlock) GetNodeLocation(selector string) *NodeLocation {
	panic(ErrNotSupport)
}

func (x RulesBlock) SetNodeLocation(selector string, nodeLocation *NodeLocation) error {
	panic(ErrNotSupport)
}

// ------------------------------------------------- --------------------------------------------------------------------

// RuleBlock Represents a rule block
type RuleBlock struct {

	// Name of policy
	Name string `yaml:"name" json:"name"`

	// Query statement corresponding to the policy
	Query string `yaml:"query" json:"query"`

	// Some custom tags
	Labels map[string]string `yaml:"labels" json:"labels"`

	// Metadata for the policy
	MetadataBlock *RuleMetadataBlock `json:"metadata" yaml:"metadata"`

	// Policy output
	Output string `yaml:"output" json:"output"`

	*LocatableImpl `yaml:"-"`
}

var _ Block = &RuleBlock{}
var _ Validator = &RuleBlock{}

func NewRuleBlock() *RuleBlock {
	return &RuleBlock{
		LocatableImpl: NewLocatableImpl(),
	}
}

func (x *RuleBlock) Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	// name
	if x.Name == "" {
		errorTips := fmt.Sprintf("Rule name must not be empty")
		report := RenderErrorTemplate(errorTips, x.GetNodeLocation("name"))
		diagnostics.AddErrorMsg(report)
	}

	// query
	if x.Query == "" {
		errorTips := fmt.Sprintf("Rule query must not be empty")
		report := RenderErrorTemplate(errorTips, x.GetNodeLocation("query"))
		diagnostics.AddErrorMsg(report)
	}

	// output
	if x.Output == "" {
		errorTips := fmt.Sprintf("Rule output must not be empty")
		report := RenderErrorTemplate(errorTips, x.GetNodeLocation("output"))
		diagnostics.AddErrorMsg(report)
	}

	if x.MetadataBlock != nil {
		diagnostics.AddDiagnostics(x.MetadataBlock.Check(module, validatorContext))
	}

	return diagnostics
}

func (x *RuleBlock) IsEmpty() bool {
	return x.Name == "" &&
		len(x.Labels) == 0 &&
		x.Query == "" &&
		(x.MetadataBlock == nil || x.MetadataBlock.IsEmpty()) &&
		x.Output == ""
}

func (x *RuleBlock) Copy() *RuleBlock {
	ruleBlock := &RuleBlock{
		Name:          x.Name,
		Query:         x.Query,
		Labels:        x.Labels,
		Output:        x.Output,
		LocatableImpl: x.LocatableImpl,
	}
	if x.MetadataBlock != nil {
		ruleBlock.MetadataBlock = x.MetadataBlock.Copy(NewRuleMetadataBlockRuntime(ruleBlock))
	}
	return ruleBlock
}

// ------------------------------------------------- --------------------------------------------------------------------

// RuleMetadataBlock Represents metadata information for a block
type RuleMetadataBlock struct {

	// A globally unique policy ID
	Id string `yaml:"id" json:"id"`

	// The severity of the problem
	Severity string `yaml:"severity" json:"severity"`

	// The Provider to which it is associated
	Provider string `yaml:"provider" json:"provider"`

	// Some custom tags
	Tags []string `yaml:"tags" json:"tags"`

	// Who is the author of the strategy
	Author string `yaml:"author" json:"author"`

	// The fix must be a local file relative path that points to a Markdown file
	Remediation string `yaml:"remediation" json:"remediation"`

	// Bug title
	Title string `yaml:"title" json:"title"`

	// Some description of the Bug
	Description string `yaml:"description" json:"description"`

	*LocatableImpl `yaml:"-"`
	runtime        *RuleMetadataBlockRuntime
}

var _ Block = &RuleMetadataBlock{}
var _ Validator = &RuleMetadataBlock{}

func NewRuleMetadataBlock(rule *RuleBlock) *RuleMetadataBlock {
	x := &RuleMetadataBlock{
		LocatableImpl: NewLocatableImpl(),
	}
	x.runtime = NewRuleMetadataBlockRuntime(rule)
	return x
}

func (x *RuleMetadataBlock) Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	// The rule id must be globally unique if it specifies
	if x.Id != "" {
		if _, exists := validatorContext.GetRuleBlockById(x.Id); exists {
			errorTips := fmt.Sprintf("Rule metadata id must not duplication: %s", x.Id)
			report := RenderErrorTemplate(errorTips, x.GetNodeLocation("id"+NodeLocationSelfValue))
			diagnostics.AddErrorMsg(report)
		} else {
			validatorContext.AddRuleBlock(x.runtime.rule)
		}
	}

	if x.Remediation != "" {
		if strings.Contains(x.Remediation, "..") {
			errorTips := fmt.Sprintf("Rule %s metadata remediation file path can not contains ..", x.runtime.rule.Name)
			report := RenderErrorTemplate(errorTips, x.GetNodeLocation("remediation"+NodeLocationSelfValue))
			diagnostics.AddErrorMsg(report)
		} else {
			remediationFileExists := filepath.Join(utils.AbsPath(module.ModuleLocalDirectory), x.Remediation)
			if !utils.Exists(remediationFileExists) {
				errorTips := fmt.Sprintf("Rule %s metadata remediation file do not exists or it is is not file", x.runtime.rule.Name)
				report := RenderErrorTemplate(errorTips, x.GetNodeLocation("remediation"+NodeLocationSelfValue))
				diagnostics.AddErrorMsg(report)
			}
		}
	}

	return diagnostics
}

func (x *RuleMetadataBlock) IsEmpty() bool {
	return x.Id == "" &&
		x.Severity == "" &&
		x.Provider == "" &&
		len(x.Tags) == 0 &&
		x.Author == "" &&
		x.Remediation == "" &&
		x.Title == "" &&
		x.Description == ""
}

func (x *RuleMetadataBlock) Copy(runtime *RuleMetadataBlockRuntime) *RuleMetadataBlock {
	return &RuleMetadataBlock{
		Id:            x.Id,
		Severity:      x.Severity,
		Provider:      x.Provider,
		Tags:          x.Tags,
		Author:        x.Author,
		Title:         x.Title,
		Description:   x.Description,
		Remediation:   x.Remediation,
		LocatableImpl: x.LocatableImpl,
		runtime:       runtime,
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

type RuleMetadataBlockRuntime struct {
	rule *RuleBlock
}

func NewRuleMetadataBlockRuntime(rule *RuleBlock) *RuleMetadataBlockRuntime {
	return &RuleMetadataBlockRuntime{
		rule: rule,
	}
}

// ------------------------------------------------- --------------------------------------------------------------------
