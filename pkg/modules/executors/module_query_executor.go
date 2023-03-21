package executors

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/selefra/selefra-provider-sdk/grpc/shard"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/modules/planner"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"path/filepath"
	"sync"
)

// ------------------------------------------------- --------------------------------------------------------------------

// RuleQueryResult Indicates the query result of a rule
type RuleQueryResult struct {

	// The index number of the current task
	Index int

	// Which module does this rule belong to
	Module *module.Module

	// What is the render value after query
	RuleBlock *module.RuleBlock

	// What is the query plan used to query the rules, with some context information and so on
	RulePlan *planner.RulePlan

	// Which version of which provider is used
	Provider *registry.Provider

	// Which configuration is used
	ProviderConfiguration *module.ProviderBlock

	// Which database is being queried
	Schema string

	// Find the row of data in issue
	Row *schema.Row
}

// ------------------------------------------------- --------------------------------------------------------------------

// ModuleQueryExecutorOptions Option to perform module queries
type ModuleQueryExecutorOptions struct {

	// Query plan to execute
	Plan *planner.ModulePlan

	// The path to install to
	DownloadWorkspace string

	// Receive real-time message feedback
	MessageChannel *message.Channel[*schema.Diagnostics]

	// The rules detected during query execution are put into this channel
	RuleQueryResultChannel *message.Channel[*RuleQueryResult]

	// Tracking installation progress
	ProgressTracker getter.ProgressTracker

	// Used to communicate with the provider
	ProviderInformationMap map[string]*shard.GetProviderInformationResponse

	// Each Provider may have multiple Fetch tasks. As long as the policy is bound to the Provider, the policy must be executed for all Storage of the Provider
	ProviderExpandMap map[string][]*planner.ProviderContext

	// The number of concurrent queries used
	WorkerNum uint64
}

// ------------------------------------------------- --------------------------------------------------------------------

const ModuleQueryExecutorName = "module-query-executor"

type ModuleQueryExecutor struct {
	options *ModuleQueryExecutorOptions

	//ruleMetricCounter *RuleMetricCounter
	//ruleMetricChannel chan *RuleMetric
}

var _ Executor = &ModuleQueryExecutor{}

func NewModuleQueryExecutor(options *ModuleQueryExecutorOptions) *ModuleQueryExecutor {
	return &ModuleQueryExecutor{
		options: options,
		//ruleMetricCounter: NewRuleMetricCounter(),
		//ruleMetricChannel: make(chan *RuleMetric, 100),
	}
}

func (x *ModuleQueryExecutor) Name() string {
	return ModuleQueryExecutorName
}

// ------------------------------------------------- --------------------------------------------------------------------

//func (x *ModuleQueryExecutor) StartMetricWorker() {
//	go func() {
//		for metric := range x.ruleMetricChannel {
//			x.ruleMetricCounter.Submit(metric)
//		}
//	}()
//}
//
//func (x *ModuleQueryExecutor) SubmitRuleMetric(rule string, hits int) {
//	x.ruleMetricChannel <- &RuleMetric{Rule: rule, HitCount: hits}
//}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *ModuleQueryExecutor) Execute(ctx context.Context) *schema.Diagnostics {

	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
		x.options.RuleQueryResultChannel.SenderWaitAndClose()
	}()

	rulePlanSlice := x.makeRulePlanSlice(ctx, x.options.Plan)
	if len(rulePlanSlice) == 0 {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("module %s no rule need query", x.options.Plan.BuildFullName()))
		return nil
	}
	channel := x.toRulePlanChannel(rulePlanSlice)
	x.RunQueryWorker(ctx, channel)

	//close(x.ruleMetricChannel)

	return nil
}

func (x *ModuleQueryExecutor) RunQueryWorker(ctx context.Context, channel chan *planner.RulePlan) {
	wg := sync.WaitGroup{}
	for i := uint64(0); i < x.options.WorkerNum; i++ {
		wg.Add(1)
		NewModuleQueryExecutorWorker(x, channel, &wg).Run(ctx)
	}
	wg.Wait()
}

func (x *ModuleQueryExecutor) toRulePlanChannel(rulePlanSlice []*planner.RulePlan) chan *planner.RulePlan {
	rulePlanChannel := make(chan *planner.RulePlan, len(rulePlanSlice))
	for _, rulePlan := range rulePlanSlice {
		rulePlanChannel <- rulePlan
	}
	close(rulePlanChannel)
	return rulePlanChannel
}

// All the rule execution plans of the module and submodules are levelled and then placed in a task queue
func (x *ModuleQueryExecutor) makeRulePlanSlice(ctx context.Context, modulePlan *planner.ModulePlan) []*planner.RulePlan {

	rulePlanSlice := make([]*planner.RulePlan, 0)

	// The rule execution plan for the current module
	if len(modulePlan.RulesPlan) != 0 {
		rulePlanSlice = append(rulePlanSlice, modulePlan.RulesPlan...)
	}

	// The execution plan of the submodule
	for _, subModule := range modulePlan.SubModulesPlan {
		rulePlanSlice = append(rulePlanSlice, x.makeRulePlanSlice(ctx, subModule)...)
	}

	return rulePlanSlice
}

// ------------------------------------------------- --------------------------------------------------------------------

type ModuleQueryExecutorWorker struct {
	ruleChannel chan *planner.RulePlan
	wg          *sync.WaitGroup

	moduleQueryExecutor *ModuleQueryExecutor
}

func NewModuleQueryExecutorWorker(moduleQueryExecutor *ModuleQueryExecutor, rulePlanChannel chan *planner.RulePlan, wg *sync.WaitGroup) *ModuleQueryExecutorWorker {
	return &ModuleQueryExecutorWorker{
		ruleChannel:         rulePlanChannel,
		wg:                  wg,
		moduleQueryExecutor: moduleQueryExecutor,
	}
}

func (x *ModuleQueryExecutorWorker) Run(ctx context.Context) {
	go func() {
		defer func() {
			x.wg.Done()
		}()

		for rulePlan := range x.ruleChannel {
			x.execRulePlan(ctx, rulePlan)
		}

	}()
}

func (x *ModuleQueryExecutorWorker) sendMessage(diagnostics *schema.Diagnostics) {
	if utils.IsNotEmpty(diagnostics) {
		x.moduleQueryExecutor.options.MessageChannel.Send(diagnostics)
	}
}

func (x *ModuleQueryExecutorWorker) execRulePlan(ctx context.Context, rulePlan *planner.RulePlan) {

	x.sendMessage(schema.NewDiagnostics().AddInfo("Rule %s begin exec...", rulePlan.String()))

	storages := x.moduleQueryExecutor.options.ProviderExpandMap[rulePlan.BindingProviderName]
	if len(storages) == 0 {
		errorMsg := fmt.Sprintf("Rule %s binding provider %s not found, can not exec query", rulePlan.String(), rulePlan.BindingProviderName)
		x.sendMessage(schema.NewDiagnostics().AddErrorMsg(errorMsg))
		return
	}
	for _, storage := range storages {

		x.execStorageQuery(ctx, rulePlan, storage)
		// TODO Stage log
	}
	// TODO log

	x.sendMessage(schema.NewDiagnostics().AddInfo("Rule %s exec done", rulePlan.String()))
}

func (x *ModuleQueryExecutorWorker) execStorageQuery(ctx context.Context, rulePlan *planner.RulePlan, providerContext *planner.ProviderContext) {

	resultSet, diagnostics := providerContext.Storage.Query(ctx, rulePlan.Query)
	if utils.HasError(diagnostics) {
		x.sendMessage(schema.NewDiagnostics().AddErrorMsg("rule %s exec error: %s", rulePlan.String(), diagnostics.ToString()))
		return
	}

	// TODO Print log prompt
	//x.moduleQueryExecutor.options.MessageChannel <- schema.NewDiagnostics().AddInfo("")
	//cli_ui.Infof("%rootConfig - Rule \"%rootConfig\"\n", rule.Path, rule.Name)
	//cli_ui.Infoln("Schema:")
	//cli_ui.Infoln(schema + "\n")
	//cli_ui.Infoln("Description:")

	for {
		rows, d := resultSet.ReadRows(100)
		if rows != nil {
			for _, row := range rows.SplitRowByRow() {
				x.processRuleRow(ctx, rulePlan, providerContext, row)
			}
		}
		if utils.HasError(d) {
			x.sendMessage(d)
		}
		if rows == nil || rows.RowCount() == 0 {
			break
		}
	}
}

// Process the row queried by the rule
func (x *ModuleQueryExecutorWorker) processRuleRow(ctx context.Context, rulePlan *planner.RulePlan, storage *planner.ProviderContext, row *schema.Row) {
	rowScope := planner.ExtendScope(rulePlan.RuleScope)

	// Inject the queried rows into the scope
	values := row.GetValues()
	for index, columnName := range row.GetColumnNames() {
		rowScope.SetVariable(columnName, values[index])
	}

	// Render the actual values for the query results of the rule
	ruleBlockResult, diagnostics := x.renderRule(ctx, rulePlan, rowScope)
	if utils.HasError(diagnostics) {
		x.moduleQueryExecutor.options.MessageChannel.Send(diagnostics)
		return
	}

	result := &RuleQueryResult{
		Module:                rulePlan.Module,
		RulePlan:              rulePlan,
		RuleBlock:             ruleBlockResult,
		Provider:              registry.NewProvider(storage.ProviderName, storage.ProviderVersion),
		ProviderConfiguration: storage.ProviderConfiguration,
		Schema:                storage.Schema,
		Row:                   row,
	}
	x.moduleQueryExecutor.options.RuleQueryResultChannel.Send(result)

	//x.sendMessage(schema.NewDiagnostics().AddInfo(json_util.ToJsonString(ruleBlockResult)))

}

func (x *ModuleQueryExecutorWorker) renderRule(ctx context.Context, rulePlan *planner.RulePlan, rowScope *planner.Scope) (*module.RuleBlock, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	ruleBlock := rulePlan.RuleBlock.Copy()

	// Start rendering the dependent variables
	// name
	if ruleBlock.Name != "" {
		ruleName, err := rowScope.RenderingTemplate(rulePlan.Name, rulePlan.Name)
		if err != nil {
			// TODO Construct error context
			return nil, diagnostics.AddErrorMsg("render rule name error: %s", err.Error())
		}
		ruleBlock.Name = ruleName
	}

	// labels
	if len(ruleBlock.Labels) > 0 {
		labels := make(map[string]string)
		for key, value := range rulePlan.Labels {
			newValue, err := rowScope.RenderingTemplate(value, value)
			if err != nil {
				// TODO Construct error context
				return nil, diagnostics.AddErrorMsg("render rule labels error: %s", err.Error())
			}
			labels[key] = newValue
		}
		ruleBlock.Labels = labels
	}

	// output
	if ruleBlock.Output != "" {
		output, err := rowScope.RenderingTemplate(rulePlan.Output, rulePlan.Output)
		if err != nil {
			// TODO Construct error context
			return nil, diagnostics.AddErrorMsg("render output labels error: %s", err.Error())
		}
		ruleBlock.Output = output
	}

	// Rendering of metadata blocks
	d := x.renderRuleMetadata(ctx, rulePlan, ruleBlock, rowScope)
	if diagnostics.AddDiagnostics(d).HasError() {
		return nil, diagnostics
	}

	return ruleBlock, diagnostics
}

// A block of render policy metadata
func (x *ModuleQueryExecutorWorker) renderRuleMetadata(ctx context.Context, rulePlan *planner.RulePlan, ruleBlock *module.RuleBlock, rowScope *planner.Scope) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()
	var err error

	if ruleBlock.MetadataBlock == nil {
		return nil
	}
	metadata := ruleBlock.MetadataBlock

	// description
	if metadata.Description != "" {
		metadata.Description, err = rowScope.RenderingTemplate(metadata.Description, metadata.Description)
		if err != nil {
			// TODO
			return diagnostics.AddErrorMsg("rendering rule description error: %s ", err.Error())
		}
	}

	// title
	if metadata.Title != "" {
		metadata.Title, err = rowScope.RenderingTemplate(metadata.Title, metadata.Title)
		if err != nil {
			// TODO
			return diagnostics.AddErrorMsg("rendering rule title error: %s ", err.Error())
		}
	}

	// Read the text of the fix, if necessary
	if metadata.Remediation != "" {
		markdownFileFullPath := filepath.Join(rulePlan.Module.ModuleLocalDirectory, metadata.Remediation)
		file, err := os.ReadFile(markdownFileFullPath)
		if err != nil {
			return diagnostics.AddErrorMsg("read file %s error: %s", markdownFileFullPath, err.Error())
		}
		metadata.Remediation = string(file)
	}

	// tags
	if len(metadata.Tags) != 0 {
		newTags := make([]string, len(metadata.Tags))
		for index, tag := range metadata.Tags {
			newTag, err := rowScope.RenderingTemplate(tag, tag)
			if err != nil {
				// TODO
				return diagnostics.AddErrorMsg("rendering tag error: %s", err.Error())
			}
			newTags[index] = newTag
		}
		metadata.Tags = newTags
	}

	// author
	if metadata.Author != "" {
		author, err := rowScope.RenderingTemplate(metadata.Author, metadata.Author)
		if err != nil {
			// TODO
			return diagnostics.AddErrorMsg("render author error: %s", err.Error())
		}
		metadata.Author = author
	}

	// provider
	if metadata.Provider != "" {
		provider, err := rowScope.RenderingTemplate(metadata.Provider, metadata.Provider)
		if err != nil {
			// TODO
			return diagnostics.AddErrorMsg("render provider error: %s", err.Error())
		}
		metadata.Provider = provider
	}

	// severity
	if metadata.Severity != "" {
		severity, err := rowScope.RenderingTemplate(metadata.Severity, metadata.Severity)
		if err != nil {
			// TODO
			return diagnostics.AddErrorMsg("render severity error: %s", err.Error())
		}
		metadata.Severity = severity
	}

	// id
	if metadata.Id != "" {
		id, err := rowScope.RenderingTemplate(metadata.Id, metadata.Id)
		if err != nil {
			// TODO
			return diagnostics.AddErrorMsg("render id error: %s", err.Error())
		}
		metadata.Id = id
	}

	return diagnostics
}

// ------------------------------------------------- --------------------------------------------------------------------

//type RuleMetricCounter struct {
//	ruleMetricMap map[string]*RuleMetric
//}
//
//func NewRuleMetricCounter() *RuleMetricCounter {
//	return &RuleMetricCounter{
//		ruleMetricMap: make(map[string]*RuleMetric),
//	}
//}
//
//func (x *RuleMetricCounter) Submit(ruleMetric *RuleMetric) {
//	if ruleMetric == nil {
//		return
//	}
//	lastRule, exists := x.ruleMetricMap[ruleMetric.Rule]
//	if !exists {
//		x.ruleMetricMap[ruleMetric.Rule] = ruleMetric
//		return
//	} else {
//		x.ruleMetricMap[ruleMetric.Rule] = ruleMetric.Merge(lastRule)
//	}
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//type RuleMetric struct {
//	Rule     string
//	HitCount int
//}
//
//func (x *RuleMetric) Merge(other *RuleMetric) *RuleMetric {
//	if x == nil {
//		return other
//	} else if other == nil {
//		return x
//	}
//	if x.Rule != other.Rule {
//		return nil
//	}
//	return &RuleMetric{
//		Rule:     x.Rule,
//		HitCount: x.HitCount + other.HitCount,
//	}
//}

// ------------------------------------------------- --------------------------------------------------------------------

//// create table name to provider name mapping
//func (x *ModuleQueryExecutor) buildTableToProviderMap() (map[string]string, *schema.Diagnostics) {
//	diagnostics := schema.NewDiagnostics()
//	tableToProviderMap := make(map[string]string, 0)
//	for providerName, providerPlugin := range x.options.ProviderPluginMap {
//		information, err := providerPlugin.Provider().GetProviderInformation(context.Background(), &shard.GetProviderInformationRequest{})
//		if err != nil {
//			return nil, diagnostics
//		}
//		if diagnostics.AddDiagnostics(information.Diagnostics).HasError() {
//			return nil, diagnostics
//		}
//		for tableName := range information.Tables {
//			tableToProviderMap[tableName] = providerName
//		}
//	}
//	return tableToProviderMap, diagnostics
//}

// ------------------------------------------------- --------------------------------------------------------------------
