package registry

import (
	"context"
	"fmt"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

// ProviderLocalRegistry The local path implementation of the provider repository
type ProviderLocalRegistry struct {
	registryDirectory          string
	registryGitHubRepoFullName string
}

var _ ProviderRegistry = (*ProviderLocalRegistry)(nil)

func NewProviderLocalRegistry(registryDirectory string, registryGitHubRepoFullName ...string) (*ProviderLocalRegistry, error) {
	stat, err := os.Stat(registryDirectory)
	if err != nil {
		return nil, fmt.Errorf("visit registryDirectory %s error: %s", registryDirectory, err.Error())
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not registryDirectory", registryDirectory)
	}

	if len(registryGitHubRepoFullName) == 0 {
		registryGitHubRepoFullName = append(registryGitHubRepoFullName, ProviderGithubRegistryDefaultRepoFullName)
	}

	return &ProviderLocalRegistry{
		registryDirectory:          registryDirectory,
		registryGitHubRepoFullName: registryGitHubRepoFullName[0],
	}, nil
}

func (x *ProviderLocalRegistry) CheckUpdate(ctx context.Context, provider *Provider) (*Provider, error) {

	if provider.IsLatestVersion() {
		return nil, nil
	}

	metaPath := filepath.Join(x.registryDirectory, ProvidersListDirectoryName, provider.Name, MetaDataFileName)
	meta, err := utils.ReadYamlFile[*ProviderMetadata](metaPath)
	if err != nil {
		return nil, err
	}

	if meta.LatestVersion == provider.Version {
		return nil, nil
	}

	return NewProvider(provider.Name, meta.LatestVersion), nil
}

func (x *ProviderLocalRegistry) GetMetadata(ctx context.Context, provider *Provider) (*ProviderMetadata, error) {
	metaPath := filepath.Join(x.registryDirectory, ProvidersListDirectoryName, provider.Name, MetaDataFileName)
	meta, err := utils.ReadYamlFile[*ProviderMetadata](metaPath)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (x *ProviderLocalRegistry) GetSupplement(ctx context.Context, provider *Provider) (*ProviderSupplement, error) {
	supplementPath := filepath.Join(x.registryDirectory, ProvidersListDirectoryName, provider.Name, provider.Version, SupplementFileName)
	supplement, err := utils.ReadYamlFile[*ProviderSupplement](supplementPath)
	if err != nil {
		return nil, err
	}
	return supplement, nil
}

func (x *ProviderLocalRegistry) GetLatestVersion(ctx context.Context, provider *Provider) (*Provider, error) {
	metaPath := filepath.Join(x.registryDirectory, ProvidersListDirectoryName, provider.Name, MetaDataFileName)
	meta, err := utils.ReadYamlFile[*ProviderMetadata](metaPath)
	if err != nil {
		return nil, err
	}
	return NewProvider(provider.Name, meta.LatestVersion), nil
}

func (x *ProviderLocalRegistry) GetAllVersion(ctx context.Context, provider *Provider) ([]*Provider, error) {
	meta, err := x.GetMetadata(ctx, provider)
	if err != nil {
		return nil, err
	}

	providerSlice := make([]*Provider, 0, len(meta.Versions))
	for _, v := range meta.Versions {
		providerSlice = append(providerSlice, NewProvider(provider.Name, v))
	}
	return providerSlice, nil
}

func (x *ProviderLocalRegistry) Download(ctx context.Context, provider *Provider, options *ProviderRegistryDownloadOptions) (string, error) {
	registry, err := NewProviderGithubRegistry(NewProviderGithubRegistryOptions(x.registryDirectory, x.registryGitHubRepoFullName))
	if err != nil {
		return "", err
	}
	return downloadProvider(ctx, registry.buildRegistryUrl(), provider, options)
}

func (x *ProviderLocalRegistry) Search(ctx context.Context, keyword string) ([]*Provider, error) {
	allProviderSlice, err := x.List(ctx)
	if err != nil {
		return nil, err
	}
	keyword = strings.ToLower(keyword)
	hitProviderSlice := make([]*Provider, 0)
	for _, provider := range allProviderSlice {
		if strings.Contains(strings.ToLower(provider.Name), keyword) {
			hitProviderSlice = append(hitProviderSlice, provider)
		}
	}
	return hitProviderSlice, nil
}

func (x *ProviderLocalRegistry) List(ctx context.Context) ([]*Provider, error) {
	providersListDirectoryPath := filepath.Join(x.registryDirectory, ProvidersListDirectoryName)
	entrySlice, err := os.ReadDir(providersListDirectoryPath)
	if err != nil {
		return nil, err
	}
	providerSlice := make([]*Provider, 0)
	for _, entry := range entrySlice {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "template" {
			continue
		}
		metaFilePath := filepath.Join(providersListDirectoryPath, entry.Name(), MetaDataFileName)
		meta, err := utils.ReadYamlFile[*ProviderMetadata](metaFilePath)
		if err != nil {
			return nil, err
		}
		providerSlice = append(providerSlice, NewProvider(meta.Name, meta.LatestVersion))
	}
	return providerSlice, nil
}
