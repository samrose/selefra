package planner

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	moduleBlock "github.com/selefra/selefra/pkg/modules/module"
)

// ------------------------------------------------- --------------------------------------------------------------------

// MakeModuleQueryPlan Generate an execution plan for the module
func MakeModuleQueryPlan(ctx context.Context, options *ModulePlannerOptions) (*ModulePlan, *schema.Diagnostics) {
	return NewModulePlanner(options).MakePlan(ctx)
}

// ------------------------------------------------- --------------------------------------------------------------------

// ModulePlan Represents the execution plan of a module
type ModulePlan struct {
	Instruction map[string]interface{}

	// Which module is this execution plan generated for
	*module.Module

	// Scope at the module level
	ModuleScope *Scope

	// The execution plan of the submodule
	SubModulesPlan []*ModulePlan

	// The execution plan of the rule under this module
	RulesPlan []*RulePlan
}

//// ------------------------------------------------- --------------------------------------------------------------------
//
//// RootModulePlan The execution plan of the root module
//type RootModulePlan struct {
//
//	// The root module's execution plan is also a module execution plan
//	*ModulePlan
//
//	// The provider pull plan for all the following modules is extracted to the root module level
//	ProviderFetchPlanSlice []*ProviderFetchPlan
//}
//

// ------------------------------------------------- --------------------------------------------------------------------

// ModulePlannerOptions Options when creating the Module Planner
type ModulePlannerOptions struct {
	Instruction map[string]interface{}
	// make plan for which module
	Module *module.Module

	// Table to Provider mapping
	TableToProviderMap map[string]string
}

// ------------------------------------------------- --------------------------------------------------------------------

// ModulePlanner Used to generate an execution plan for a module
type ModulePlanner struct {
	options *ModulePlannerOptions
}

var _ Planner[*ModulePlan] = &ModulePlanner{}

func NewModulePlanner(options *ModulePlannerOptions) *ModulePlanner {
	return &ModulePlanner{
		options: options,
	}
}

func (x *ModulePlanner) Name() string {
	return "module-planner"
}

func (x *ModulePlanner) MakePlan(ctx context.Context) (*ModulePlan, *schema.Diagnostics) {
	return x.buildModulePlanner(ctx, x.options.Module, NewScope())
}

// Specify execution plans for modules and submodules
func (x *ModulePlanner) buildModulePlanner(ctx context.Context, module *module.Module, moduleScope *Scope) (*ModulePlan, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	modulePlan := &ModulePlan{
		Instruction: x.options.Instruction,
		Module:      module,
		// Inherits the scope of the parent module
		ModuleScope:    moduleScope,
		SubModulesPlan: nil,
		RulesPlan:      nil,
	}

	// Generate an execution plan for the rules in the module
	for _, ruleBlock := range module.RulesBlock {
		rulePlan, d := NewRulePlanner(&RulePlannerOptions{
			ModulePlan:         modulePlan,
			Module:             module,
			ModuleScope:        modulePlan.ModuleScope,
			RuleBlock:          ruleBlock,
			TableToProviderMap: x.options.TableToProviderMap,
		}).MakePlan(ctx)
		if diagnostics.Add(d).HasError() {
			return nil, diagnostics
		}
		modulePlan.RulesPlan = append(modulePlan.RulesPlan, rulePlan)
	}

	var subModuleInputMap map[string]*moduleBlock.ModuleBlock
	if len(module.ModulesBlock) != 0 {
		subModuleInputMap = module.ModulesBlock.ModulesInputMap()
	}
	// Generate an execution plan for the submodules
	for _, subModule := range module.SubModules {

		subModuleScope := ExtendScope(modulePlan.ModuleScope)

		// Also, the module may have some initialized variables
		if subModuleInputMap != nil {
			if subModuleBlock := subModuleInputMap[subModule.Source]; subModuleBlock != nil && len(subModuleBlock.Input) != 0 {
				subModuleScope.SetVariables(subModuleBlock.Input)
			}
		}

		subModulePlan, d := x.buildModulePlanner(ctx, subModule, subModuleScope)
		if diagnostics.AddDiagnostics(d).HasError() {
			return nil, diagnostics
		}
		modulePlan.SubModulesPlan = append(modulePlan.SubModulesPlan, subModulePlan)
	}

	return modulePlan, diagnostics
}

// ------------------------------------------------- --------------------------------------------------------------------
