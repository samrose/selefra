package planner

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/registry"
	selefraVersion "github.com/selefra/selefra/pkg/version"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderVersionVoteService When multiple versions of the same provider are available for a module, which version should be used? So take a vote!
type ProviderVersionVoteService struct {

	// <providerName, ProviderVote>
	providerVersionVoteMap map[string]*ProviderVote
}

func NewProviderVersionVoteService() *ProviderVersionVoteService {
	return &ProviderVersionVoteService{
		providerVersionVoteMap: make(map[string]*ProviderVote),
	}
}

// Vote Vote on the module to see which version should be used
func (x *ProviderVersionVoteService) Vote(ctx context.Context, module *module.Module) *schema.Diagnostics {
	diagnostics := schema.NewDiagnostics()
	for _, requiredProviderBlock := range module.SelefraBlock.RequireProvidersBlock {
		if _, exists := x.providerVersionVoteMap[requiredProviderBlock.Source]; !exists {
			providerVote, d := NewProviderVote(ctx, requiredProviderBlock)
			if diagnostics.AddDiagnostics(d).HasError() {
				return diagnostics
			}
			x.providerVersionVoteMap[requiredProviderBlock.Source] = providerVote
		}
		d := x.providerVersionVoteMap[requiredProviderBlock.Source].Vote(module, requiredProviderBlock)
		if diagnostics.AddDiagnostics(d).HasError() {
			return diagnostics
		}
	}
	return diagnostics
}

// TODO

// TODO
//// GiveMeResult Query voting result
//func (x *ProviderVersionVoteService) GiveMeResult(providerName string) (string, *schema.Diagnostics) {
//	vote, exists := x.providerVersionVoteMap[providerName]
//	if !exists {
//		return "", schema.NewDiagnostics().AddErrorMsg("")
//	}
//	return "",
//}

// ------------------------------------------------- --------------------------------------------------------------------

type ProviderVote struct {
	TotalVoteTimes      int
	ProviderName        string
	VersionVoteCountMap map[string]*VersionVoteSummary
	providerMetadata    *registry.ProviderMetadata
}

func NewProviderVote(ctx context.Context, requiredProviderBlock *module.RequireProviderBlock) (*ProviderVote, *schema.Diagnostics) {
	x := &ProviderVote{}
	x.ProviderName = requiredProviderBlock.Source
	d := x.InitProviderVersionVoteCountMap(ctx, requiredProviderBlock)
	return x, d
}

// Vote Each module can participate in voting
func (x *ProviderVote) Vote(voteModule *module.Module, requiredProviderBlock *module.RequireProviderBlock) *schema.Diagnostics {

	x.TotalVoteTimes++

	// If it is the latest version, replace it with the latest version
	versionString := requiredProviderBlock.Version
	if selefraVersion.IsLatestVersion(versionString) {
		versionString = x.providerMetadata.LatestVersion
	}
	constraint, err := version.NewConstraint(versionString)
	if err != nil {
		location := requiredProviderBlock.GetNodeLocation("version" + module.NodeLocationSelfValue)
		report := module.RenderErrorTemplate(fmt.Sprintf("required provider version constraint parse failed: %s", versionString), location)
		return schema.NewDiagnostics().AddErrorMsg(report)
	}

	voteSuccessCount := 0
	for _, voteSummary := range x.VersionVoteCountMap {
		if selefraVersion.IsConstraintsAllow(constraint, voteSummary.ProviderVersion) {
			voteSummary.VoteSet[voteModule] = struct{}{}
			voteSuccessCount++
		}
	}

	if voteSuccessCount == 0 {
		canUseVersions := selefraVersion.Sort(x.GetVoteVersions())
		location := requiredProviderBlock.GetNodeLocation("version" + module.NodeLocationSelfValue)
		errorTips := fmt.Sprintf("required provider version constraint %s , no version was found that met the requirements, can use versions: %s", versionString, strings.Join(canUseVersions, ", "))
		report := module.RenderErrorTemplate(errorTips, location)
		return schema.NewDiagnostics().AddErrorMsg(report)
	}

	return nil
}

// InitProviderVersionVoteCountMap Obtain the Provider versions from Registry and vote for these products later
func (x *ProviderVote) InitProviderVersionVoteCountMap(ctx context.Context, block *module.RequireProviderBlock) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	x.VersionVoteCountMap = make(map[string]*VersionVoteSummary)

	// It's not actually going to download, so it doesn't matter what the path is here
	options := registry.NewProviderGithubRegistryOptions("./")
	provider, err := registry.NewProviderGithubRegistry(options)
	if err != nil {
		return diagnostics.AddErrorMsg("create provider github registry failed: %s", err.Error())
	}
	metadata, err := provider.GetMetadata(ctx, registry.NewProvider(x.ProviderName, selefraVersion.VersionLatest))
	if err != nil {
		location := block.GetNodeLocation("source" + module.NodeLocationSelfValue)
		report := module.RenderErrorTemplate(fmt.Sprintf("get provider %s meta information from registry error: %s", x.ProviderName, err.Error()), location)
		return diagnostics.AddErrorMsg(report)
	}
	if len(metadata.Versions) == 0 {
		return diagnostics.AddErrorMsg("provider %s registry metadata not found any version", x.ProviderName)
	}
	for _, providerVersion := range metadata.Versions {
		summary, d := NewVoteSummary(x.ProviderName, providerVersion)
		if diagnostics.AddDiagnostics(d).HasError() {
			return diagnostics
		}
		x.VersionVoteCountMap[providerVersion] = summary
	}
	x.providerMetadata = metadata
	return diagnostics
}

// GetWinnersVersionVoteSummary Get the version that wins the vote. There may be multiple versions that win at the same time
func (x *ProviderVote) GetWinnersVersionVoteSummary() map[string]*VersionVoteSummary {
	m := make(map[string]*VersionVoteSummary)
	for versionString, voteSummary := range x.VersionVoteCountMap {
		if len(voteSummary.VoteSet) == x.TotalVoteTimes {
			m[versionString] = voteSummary
		}
	}
	return m
}

// GetWinnersVersionSlice Gets the version numbers of all versions that won the vote
func (x *ProviderVote) GetWinnersVersionSlice() []string {
	versionSlice := make([]string, 0)
	for versionString := range x.GetWinnersVersionVoteSummary() {
		versionSlice = append(versionSlice, versionString)
	}
	return versionSlice
}

// GetVoteVersions Get the versions that are voted on
func (x *ProviderVote) GetVoteVersions() []string {
	versionStringSlice := make([]string, 0)
	for versionString := range x.VersionVoteCountMap {
		versionStringSlice = append(versionStringSlice, versionString)
	}
	return versionStringSlice
}

// ToModuleAllowProviderVersionMap Convert to which versions of this Provider are supported by the module
func (x *ProviderVote) ToModuleAllowProviderVersionMap() map[*module.Module][]string {
	moduleUseProviderVersionMap := make(map[*module.Module][]string, 0)
	for providerVersion, voteSummary := range x.VersionVoteCountMap {
		for module := range voteSummary.VoteSet {
			versionSlice := moduleUseProviderVersionMap[module]
			moduleUseProviderVersionMap[module] = append(versionSlice, providerVersion)
		}
	}
	return moduleUseProviderVersionMap
}

// ------------------------------------------------- --------------------------------------------------------------------

type VersionVoteSummary struct {

	// Which version
	ProviderVersion *version.Version

	// How many votes did you get
	VoteSet map[*module.Module]struct{}
}

func NewVoteSummary(providerName, providerVersion string) (*VersionVoteSummary, *schema.Diagnostics) {
	newVersion, err := version.NewVersion(providerVersion)
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg("parse provider %s version %s error: %s", providerName, providerVersion, err.Error())
	}
	return &VersionVoteSummary{
		ProviderVersion: newVersion,
		VoteSet:         make(map[*module.Module]struct{}),
	}, nil
}

// ------------------------------------------------- --------------------------------------------------------------------
