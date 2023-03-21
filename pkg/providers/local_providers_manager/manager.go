package local_providers_manager

import (
	"github.com/selefra/selefra/pkg/registry"
	"path/filepath"
)

const (

	// LocalProvidersDirectoryName The locally installed providers are stored in the download directory
	LocalProvidersDirectoryName = "providers"

	// LocalProvidersVersionMetaFileName The local provider version will have a metadata file, and this field indicates the name of that metadata file
	LocalProvidersVersionMetaFileName = ".version-meta.json"

	// LocalProvidersProviderMetaFileName The local provider will have a metadata file, and this field represents the name of that metadata file
	LocalProvidersProviderMetaFileName = ".provider-meta.json"
)

// ------------------------------------------------- --------------------------------------------------------------------

// LocalProvidersManager TODO Add file locks to avoid concurrency problems during multi-process operations
type LocalProvidersManager struct {

	// selefra Specifies the storage path of the downloaded file
	downloadWorkspace string

	// The provider registry is used to update the provider from the remote end
	providerRegistry registry.ProviderRegistry
}

func NewLocalProvidersManager(downloadWorkspace string) (*LocalProvidersManager, error) {

	// init provider registry
	options := registry.NewProviderGithubRegistryOptions(downloadWorkspace)
	providerRegistry, err := registry.NewProviderGithubRegistry(options)
	if err != nil {
		return nil, err
	}

	return &LocalProvidersManager{
		downloadWorkspace: downloadWorkspace,
		providerRegistry:  providerRegistry,
	}, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *LocalProvidersManager) buildLocalProvidersPath() string {
	return filepath.Join(x.downloadWorkspace, LocalProvidersDirectoryName)
}

func (x *LocalProvidersManager) buildLocalProviderPath(providerName string) string {
	return filepath.Join(x.downloadWorkspace, LocalProvidersDirectoryName, providerName)
}

// provider metadata file path
func (x *LocalProvidersManager) buildLocalProviderMetaFilePath(providerName string) string {
	return filepath.Join(x.downloadWorkspace, LocalProvidersDirectoryName, providerName, LocalProvidersProviderMetaFileName)
}

// Folder in which the provider version is stored
func (x *LocalProvidersManager) buildLocalProviderVersionPath(providerName, providerVersion string) string {
	return filepath.Join(x.downloadWorkspace, LocalProvidersDirectoryName, providerName, providerVersion)
}

// Location for storing the metadata of the provider version
func (x *LocalProvidersManager) buildLocalProviderVersionMetaFilePath(providerName, providerVersion string) string {
	return filepath.Join(x.downloadWorkspace, LocalProvidersDirectoryName, providerName, providerVersion, LocalProvidersVersionMetaFileName)
}

// ------------------------------------------------- --------------------------------------------------------------------
