package registry

import (
	"context"
	"fmt"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

type ModuleLocalRegistry struct {
	registryDirectory          string
	registryGitHubRepoFullName string
}

var _ ModuleRegistry = &ModuleLocalRegistry{}

func NewModuleLocalRegistry(registryDirectory string, registryGitHubRepoFullName ...string) (*ModuleLocalRegistry, error) {
	stat, err := os.Stat(registryDirectory)
	if err != nil {
		return nil, fmt.Errorf("visit registryDirectory %s error: %s", registryDirectory, err.Error())
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not registryDirectory", registryDirectory)
	}

	if len(registryGitHubRepoFullName) == 0 {
		registryGitHubRepoFullName = append(registryGitHubRepoFullName, ModuleGithubRegistryDefaultRepoFullName)
	}

	return &ModuleLocalRegistry{
		registryDirectory:          registryDirectory,
		registryGitHubRepoFullName: registryGitHubRepoFullName[0],
	}, nil
}

func (x *ModuleLocalRegistry) CheckUpdate(ctx context.Context, module *Module) (*Module, error) {
	if module.IsLatestVersion() {
		return nil, nil
	}

	metaPath := filepath.Join(x.registryDirectory, ModulesListDirectoryName, module.Name, MetaDataFileName)
	meta, err := utils.ReadYamlFile[*ModuleMetadata](metaPath)
	if err != nil {
		return nil, err
	}

	if meta.LatestVersion == module.Version {
		return nil, nil
	}

	return NewModule(module.Name, meta.LatestUpdate), nil
}

func (x *ModuleLocalRegistry) Download(ctx context.Context, module *Module, options *ModuleRegistryDownloadOptions) (string, error) {
	registry, err := NewModuleGitHubRegistry(NewModuleGithubRegistryOptions(x.registryDirectory, x.registryGitHubRepoFullName))
	if err != nil {
		return "", err
	}
	return downloadModule(ctx, registry.buildRegistryUrl(), module, options)
}

func (x *ModuleLocalRegistry) GetLatestVersion(ctx context.Context, module *Module) (*Module, error) {
	metaPath := filepath.Join(x.registryDirectory, ModulesListDirectoryName, module.Name, MetaDataFileName)
	meta, err := utils.ReadYamlFile[*ProviderMetadata](metaPath)
	if err != nil {
		return nil, err
	}
	return NewModule(module.Name, meta.LatestVersion), nil
}

func (x *ModuleLocalRegistry) GetAllVersion(ctx context.Context, module *Module) ([]*Module, error) {
	meta, err := x.GetMetadata(ctx, module)
	if err != nil {
		return nil, err
	}

	providerSlice := make([]*Module, 0, len(meta.Versions))
	for _, v := range meta.Versions {
		providerSlice = append(providerSlice, NewModule(module.Name, v))
	}
	return providerSlice, nil
}

func (x *ModuleLocalRegistry) GetMetadata(ctx context.Context, module *Module) (*ModuleMetadata, error) {
	metaPath := filepath.Join(x.registryDirectory, ModulesListDirectoryName, module.Name, MetaDataFileName)
	meta, err := utils.ReadYamlFile[*ModuleMetadata](metaPath)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (x *ModuleLocalRegistry) GetSupplement(ctx context.Context, module *Module) (*ModuleSupplement, error) {
	supplementPath := filepath.Join(x.registryDirectory, ModulesListDirectoryName, module.Name, module.Version, SupplementFileName)
	supplement, err := utils.ReadYamlFile[*ModuleSupplement](supplementPath)
	if err != nil {
		return nil, err
	}
	return supplement, nil
}

func (x *ModuleLocalRegistry) List(ctx context.Context) ([]*Module, error) {
	modulesListDirectoryPath := filepath.Join(x.registryDirectory, ModulesListDirectoryName)
	entrySlice, err := os.ReadDir(modulesListDirectoryPath)
	if err != nil {
		return nil, err
	}
	providerSlice := make([]*Module, 0)
	for _, entry := range entrySlice {
		if !entry.IsDir() {
			continue
		}
		metaFilePath := filepath.Join(modulesListDirectoryPath, entry.Name(), MetaDataFileName)
		meta, err := utils.ReadYamlFile[*ModuleMetadata](metaFilePath)
		if err != nil {
			return nil, err
		}
		providerSlice = append(providerSlice, NewModule(meta.Name, meta.LatestVersion))
	}
	return providerSlice, nil
}

func (x *ModuleLocalRegistry) Search(ctx context.Context, keyword string) ([]*Module, error) {
	allModuleSlice, err := x.List(ctx)
	if err != nil {
		return nil, err
	}
	keyword = strings.ToLower(keyword)
	hitModuleSlice := make([]*Module, 0)
	for _, module := range allModuleSlice {
		if strings.Contains(strings.ToLower(module.Name), keyword) {
			hitModuleSlice = append(hitModuleSlice, module)
		}
	}
	return hitModuleSlice, nil
}
