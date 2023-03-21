package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/utils"
)

// Get Obtain information about the installed provider on the local device
func (x *LocalProvidersManager) Get(ctx context.Context, localProvider *LocalProvider) (*LocalProvider, *schema.Diagnostics) {
	diagnostics := schema.NewDiagnostics()
	providerVersionMetaFilePath := x.buildLocalProviderVersionMetaFilePath(localProvider.Name, localProvider.Version)
	localProviderMeta, err := utils.ReadJsonFile[*LocalProvider](providerVersionMetaFilePath)
	if err != nil {
		return nil, diagnostics.AddErrorMsg("read local provider version %s meta file failed: %s", localProvider.String(), err)
	}
	return localProviderMeta, diagnostics
}
