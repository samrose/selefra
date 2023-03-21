package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/registry"
	"strings"
)

// SearchLocal Search for the provider installed on the local device
func (x *LocalProvidersManager) SearchLocal(ctx context.Context, keyword string) ([]*LocalProviderVersions, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	providers, d := x.ListProviders()
	if diagnostics.AddDiagnostics(d).HasError() {
		return nil, diagnostics
	}

	keyword = strings.ToLower(keyword)
	hitProviderSlice := make([]*LocalProviderVersions, 0)
	for _, provider := range providers {
		if strings.Contains(strings.ToLower(provider.ProviderName), keyword) {
			hitProviderSlice = append(hitProviderSlice, provider)
		}
	}

	return hitProviderSlice, diagnostics
}

// SearchRegistry Search the provider by keyword on the configured registry
func (x *LocalProvidersManager) SearchRegistry(ctx context.Context, keyword string) ([]*registry.Provider, *schema.Diagnostics) {
	providerSlice, err := x.providerRegistry.Search(ctx, keyword)
	if err != nil {
		return nil, schema.NewDiagnostics().AddError(err)
	}
	return providerSlice, nil
}
