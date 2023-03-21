package planner

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"sort"
)

// ------------------------------------------------- --------------------------------------------------------------------

type RulePlan struct {

	// The execution plan of the module to which it is associated
	ModulePlan *ModulePlan

	// The module to which it is associated
	Module *module.Module

	// Is the execution plan for which block
	*module.RuleBlock

	// Which provider is the rule bound to? Currently, a rule can be bound to only one provider
	BindingProviderName string

	// Render a good rule - bound Query
	Query string

	// Which tables are used in this Query
	BindingTables []string

	RuleScope *Scope
}

func (x *RulePlan) String() string {
	if x.MetadataBlock != nil {
		return x.Name + ":" + x.MetadataBlock.Id
	} else {
		return x.Name
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// MakeRulePlan Plan the execution of the rule
func MakeRulePlan(ctx context.Context, options *RulePlannerOptions) (*RulePlan, *schema.Diagnostics) {
	return NewRulePlanner(options).MakePlan(ctx)
}

// ------------------------------------------------- --------------------------------------------------------------------

// RulePlannerOptions Parameters required when creating a module execution plan
type RulePlannerOptions struct {

	// The execution plan of the module to which it is associated
	ModulePlan *ModulePlan

	// The module to which it is associated
	Module *module.Module

	// The scope of the owning module
	ModuleScope *Scope

	// Is the execution plan for which block
	RuleBlock *module.RuleBlock

	// Mapping between the table and the provider
	TableToProviderMap map[string]string
}

// ------------------------------------------------- --------------------------------------------------------------------

// RulePlanner An enforcement plan for this rule
type RulePlanner struct {
	options *RulePlannerOptions
}

var _ Planner[*RulePlan] = &RulePlanner{}

func (x *RulePlanner) Name() string {
	return "rule-planner"
}

func NewRulePlanner(options *RulePlannerOptions) *RulePlanner {
	return &RulePlanner{
		options: options,
	}
}

// MakePlan Develop an implementation plan for rule
func (x *RulePlanner) MakePlan(ctx context.Context) (*RulePlan, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// Render the query statement for the Rule
	ruleScope := ExtendScope(x.options.ModuleScope)
	query, err := ruleScope.RenderingTemplate(x.options.RuleBlock.Query, x.options.RuleBlock.Query)
	if err != nil {
		location := x.options.RuleBlock.GetNodeLocation("query" + module.NodeLocationSelfValue)
		// TODO 2023-2-24 15:10:15 bug: Can't correct marks used in yaml | a line
		report := module.RenderErrorTemplate(fmt.Sprintf("rendering query template error: %s", err.Error()), location)
		return nil, diagnostics.AddErrorMsg(report)
	}

	// Resolve the binding of the Rule to the Provider and table
	bindingProviders, bindingTables := x.extractBinding(query, x.options.TableToProviderMap)
	if len(bindingProviders) != 1 {
		var errorTips string
		if len(bindingProviders) == 0 {
			errorTips = fmt.Sprintf("Your rule query should use at least one of the provider tables. Check that your sql is written correctly: %s", x.options.RuleBlock.Query)
		} else {
			errorTips = fmt.Sprintf("The tables used in your rule query span multiple providers; the current version of the rule query only allows several tables from one provider to be used: %s", x.options.RuleBlock.Query)
		}
		location := x.options.RuleBlock.GetNodeLocation("query" + module.NodeLocationSelfValue)
		// TODO 2023-2-24 15:10:15 bug: Can't correct marks used in yaml | a line
		report := module.RenderErrorTemplate(errorTips, location)
		return nil, diagnostics.AddErrorMsg(report)
	}

	// Create a Rule execution plan
	return &RulePlan{

		ModulePlan: x.options.ModulePlan,
		Module:     x.options.Module,

		RuleBlock: x.options.RuleBlock,

		BindingProviderName: bindingProviders[0],
		BindingTables:       bindingTables,

		Query: query,

		RuleScope: ruleScope,
	}, diagnostics
}

// Extract the names of the tables it uses from the rendered rule Query
func (x *RulePlanner) extractBinding(query string, tableToProviderMap map[string]string) (bindingProviders []string, bindingTables []string) {
	bindingProviderSet := make(map[string]struct{})
	bindingTableSet := make(map[string]struct{})
	inWord := false
	lastIndex := 0
	for index, c := range query {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c >= '0' && c <= '9' {
			if !inWord {
				inWord = true
				lastIndex = index
			}
		} else {
			if inWord {
				word := query[lastIndex:index]
				if providerName, exists := tableToProviderMap[word]; exists {
					bindingTableSet[word] = struct{}{}
					bindingProviderSet[providerName] = struct{}{}
				}
				inWord = false
			}
		}
	}

	for providerName := range bindingProviderSet {
		bindingProviders = append(bindingProviders, providerName)
	}
	for tableName := range bindingTableSet {
		bindingTables = append(bindingTables, tableName)
	}

	// keep dictionary order, show it to console need keep same
	sort.Strings(bindingProviders)
	sort.Strings(bindingTables)

	return
}

// ------------------------------------------------- --------------------------------------------------------------------

// The old scheme does implicit provider association, while the new scheme does whitelist association
//// Extracting the provider name from the table name used by the policy is an implicit association
//func (x *RulePlanner) extractImplicitProvider(tablesName []string) ([]string, *schema.Diagnostics) {
//	diagnostics := schema.NewDiagnostics()
//	providerNameSet := make(map[string]struct{}, 0)
//	for _, tableName := range tablesName {
//		split := strings.SplitN(tableName, "_", 2)
//		if len(split) != 2 {
//			diagnostics.AddErrorMsg("can not found implicit provider name from table name %s", tableName)
//		} else {
//			providerNameSet[split[0]] = struct{}{}
//		}
//	}
//	providerNameSlice := make([]string, 0)
//	for providerName := range providerNameSet {
//		providerNameSlice = append(providerNameSlice, providerName)
//	}
//	return providerNameSlice, diagnostics
//}
//
//// Extract the names of the tables it uses from the rendered rule Query
//func (x *RulePlanner) extractTableNameSliceFromRuleQuery(s string, whitelistWordSet map[string]string) []string {
//	var matchResultSet []string
//	inWord := false
//	lastIndex := 0
//	for index, c := range s {
//		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c >= '0' && c <= '9' {
//			if !inWord {
//				inWord = true
//				lastIndex = index
//			}
//		} else {
//			if inWord {
//				word := s[lastIndex:index]
//				if _, exists := whitelistWordSet[word]; exists {
//					matchResultSet = append(matchResultSet, word)
//				}
//				inWord = false
//			}
//		}
//	}
//	return matchResultSet
//}

// ------------------------------------------------- --------------------------------------------------------------------
