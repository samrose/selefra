package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/utils"
)

// IsProviderInstalled Used to query whether the provider is installed locally
func (x *LocalProvidersManager) IsProviderInstalled(ctx context.Context, provider *LocalProvider) (bool, *schema.Diagnostics) {

	// If it is not the latest version, you can directly determine the path
	if !provider.IsLatestVersion() {
		path := x.buildLocalProviderVersionPath(provider.Name, provider.Version)
		return utils.Exists(path), nil
	}

	// If it is the latest version, obtain the version number of the latest version
	metadata, err := x.providerRegistry.GetMetadata(ctx, provider.Provider)
	if err != nil {
		return false, schema.NewDiagnostics().AddErrorMsg("provider %s get metadata error: %s", provider.Name, err.Error())
	}
	version := metadata.LatestVersion
	path := x.buildLocalProviderVersionPath(provider.Name, version)
	return utils.Exists(path), nil

}
