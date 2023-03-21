package local_providers_manager

import (
	"errors"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/utils"
	"os"
)

// ListProviders all providers installed locally
func (x *LocalProvidersManager) ListProviders() ([]*LocalProviderVersions, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	path := x.buildLocalProvidersPath()
	entrySlice, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, diagnostics.AddInfo("You haven't installed any providers yet.")
		} else {
			return nil, diagnostics.AddErrorMsg("Can not exec list command, open directory %s error: %s", path, err.Error())
		}
	}

	versionsSlice := make([]*LocalProviderVersions, 0)
	for _, entry := range entrySlice {
		if !entry.IsDir() {
			continue
		}
		providerName := entry.Name()
		versions, d := x.ListProviderVersions(providerName)

		if !diagnostics.AddDiagnostics(d).HasError() && len(versions.ProviderVersionMap) > 0 {
			versionsSlice = append(versionsSlice, versions)
		}
	}

	return versionsSlice, diagnostics
}

// ListProviderVersions Lists all the installed versions of this Provider
func (x *LocalProvidersManager) ListProviderVersions(providerName string) (*LocalProviderVersions, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	providerDirectory := x.buildLocalProviderPath(providerName)
	providerVersionEntrySlice, err := os.ReadDir(providerDirectory)
	if err != nil {
		return nil, diagnostics.AddErrorMsg("List provider versions read directory %s error: %s", utils.AbsPath(providerDirectory), err.Error())
	}

	versions := NewLocalProviderVersions(providerName)
	for _, providerVersionEntry := range providerVersionEntrySlice {
		if !providerVersionEntry.IsDir() {
			continue
		}
		providerVersion := providerVersionEntry.Name()
		path := x.buildLocalProviderVersionMetaFilePath(providerName, providerVersion)
		localProvider, err := utils.ReadJsonFile[*LocalProvider](path)
		if err != nil {
			diagnostics.AddError(err)
		} else {
			versions.AddLocalProvider(localProvider)
		}
	}

	return versions, diagnostics
}
