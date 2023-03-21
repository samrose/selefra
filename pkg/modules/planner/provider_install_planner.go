package planner

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/selefra/selefra/pkg/version"
	"strings"
)

// ------------------------------------------------ ---------------------------------------------------------------------

// MakeProviderInstallPlan Plan the provider installation for the module
func MakeProviderInstallPlan(ctx context.Context, module *module.Module) (ProvidersInstallPlan, *schema.Diagnostics) {
	return NewProviderInstallPlanner(module).MakePlan(ctx)
}

// ------------------------------------------------ ---------------------------------------------------------------------

type ProvidersInstallPlan []*ProviderInstallPlan

func (x ProvidersInstallPlan) ToMap() map[string]string {
	m := make(map[string]string)
	for _, p := range x {
		m[p.Name] = p.Version
	}
	return m
}

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderInstallPlan Indicates the installation plan of a provider
type ProviderInstallPlan struct {
	// Which version of which provider is to be used to pull data
	*registry.Provider
}

// NewProviderInstallPlan Create an installation plan based on the provider name and version number
func NewProviderInstallPlan(providerName, providerVersion string) *ProviderInstallPlan {
	return &ProviderInstallPlan{
		Provider: registry.NewProvider(providerName, providerVersion),
	}
}

// ------------------------------------------------ ---------------------------------------------------------------------

// ProviderInstallPlanner This command is used to plan the provider installation for Module
type ProviderInstallPlanner struct {
	module *module.Module
}

var _ Planner[ProvidersInstallPlan] = &ProviderInstallPlanner{}

func NewProviderInstallPlanner(module *module.Module) *ProviderInstallPlanner {
	return &ProviderInstallPlanner{
		module: module,
	}
}

func (x *ProviderInstallPlanner) Name() string {
	return "provider-install-planner"
}

func (x *ProviderInstallPlanner) MakePlan(ctx context.Context) (ProvidersInstallPlan, *schema.Diagnostics) {
	diagnostics := schema.NewDiagnostics()
	providerVersionVoteWinnerMap, d := x.providerVersionVote(ctx)
	if diagnostics.AddDiagnostics(d).HasError() {
		return nil, diagnostics
	}
	providerInstallPlanSlice := make([]*ProviderInstallPlan, 0)
	for providerName, providerVersion := range providerVersionVoteWinnerMap {
		providerInstallPlanSlice = append(providerInstallPlanSlice, NewProviderInstallPlan(providerName, providerVersion))
	}
	return providerInstallPlanSlice, diagnostics
}

// provider version election to determine which provider version to use
func (x *ProviderInstallPlanner) providerVersionVote(ctx context.Context) (map[string]string, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// Start with the root module and let all modules vote
	service := NewProviderVersionVoteService()
	x.module.Traversal(ctx, func(ctx context.Context, traversalContext *module.TraversalContext) bool {
		diagnostics.AddDiagnostics(service.Vote(ctx, traversalContext.Module))
		return true
	})
	if utils.HasError(diagnostics) {
		return nil, diagnostics
	}

	// Determine the final version used for each provider
	providerVersionVoteWinnerMap := make(map[string]string, 0)
	errorReportSlice := make([]string, 0)
	for providerName, voteInfo := range service.providerVersionVoteMap {

		winnersVersions := voteInfo.GetWinnersVersionVoteSummary()

		// The election was defeated, and no version received unanimous votes
		if len(winnersVersions) < 1 {
			errorReportSlice = append(errorReportSlice, x.buildVersionVoteFailedReport(voteInfo))
		} else {
			// Select the latest version of the provider that supports all Modules
			winnerVersionSlice := version.Sort(voteInfo.GetWinnersVersionSlice())
			winnerVersion := winnerVersionSlice[len(winnerVersionSlice)-1]
			// TODO debug log
			providerVersionVoteWinnerMap[providerName] = winnerVersion
		}

	}

	if len(errorReportSlice) > 0 {
		for index, report := range errorReportSlice {
			if index != len(errorReportSlice)-1 {
				report += "\n\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n"
			}
			diagnostics.AddErrorMsg(report)
		}
	}
	return providerVersionVoteWinnerMap, diagnostics
}

// When a vote fails, construct a general report so the user knows what went wrong
func (x *ProviderInstallPlanner) buildVersionVoteFailedReport(providerVote *ProviderVote) string {
	report := strings.Builder{}
	report.WriteString(fmt.Sprintf("Failed to vote version for provider %s: \n", providerVote.ProviderName))
	for module, versionSlice := range providerVote.ToModuleAllowProviderVersionMap() {
		version.Sort(versionSlice)
		report.WriteString(fmt.Sprintf("Module %s suport version: %s \n", module.BuildFullName(), strings.Join(versionSlice, ", ")))
	}
	report.WriteString(fmt.Sprintf("Cannot find a %s provider version that supports all of the above modules\n", providerVote.ProviderName))

	return report.String()
}

// ------------------------------------------------- --------------------------------------------------------------------
