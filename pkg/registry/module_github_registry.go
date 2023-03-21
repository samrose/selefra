package registry

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/pointer"
	"github.com/selefra/selefra/pkg/http_client"
	"github.com/selefra/selefra/pkg/utils"
	"path/filepath"
	"strings"
	"time"
)

// TODO Consider a clone-based way to execute private positions support

// ------------------------------------------------- --------------------------------------------------------------------

const (
	ModulesListDirectoryName = "module"
)

// ModuleGithubRegistryDefaultRepoFullName The default official module registry
const ModuleGithubRegistryDefaultRepoFullName = "selefra/registry"

// ------------------------------------------------- --------------------------------------------------------------------

type ModuleGithubRegistryOptions struct {
	DownloadWorkspace    string
	RegistryRepoFullName *string
}

func NewModuleGithubRegistryOptions(downloadWorkspace string, registryRepoFullName ...string) *ModuleGithubRegistryOptions {

	if len(registryRepoFullName) == 0 {
		registryRepoFullName = append(registryRepoFullName, ModuleGithubRegistryDefaultRepoFullName)
	}

	return &ModuleGithubRegistryOptions{
		DownloadWorkspace:    downloadWorkspace,
		RegistryRepoFullName: pointer.ToStringPointer(registryRepoFullName[0]),
	}
}

func (x *ModuleGithubRegistryOptions) Check() *schema.Diagnostics {
	// TODO check params
	return nil
}

// ------------------------------------------------- --------------------------------------------------------------------

type ModuleGitHubRegistry struct {
	// The owner of registry's repo
	owner string
	// The name of registry's repo
	repoName string

	options *ModuleGithubRegistryOptions
}

var _ ModuleRegistry = &ModuleGitHubRegistry{}

func NewModuleGitHubRegistry(options *ModuleGithubRegistryOptions) (*ModuleGitHubRegistry, error) {

	// set default registry url
	if options.RegistryRepoFullName == nil {
		options.RegistryRepoFullName = pointer.ToStringPointer(ModuleGithubRegistryDefaultRepoFullName)
	}

	// Parse the full name of the github repository
	owner, repo, err := utils.ParseGitHubRepoFullName(pointer.FromStringPointer(options.RegistryRepoFullName))
	if err != nil {
		return nil, err
	}

	return &ModuleGitHubRegistry{
		owner:    owner,
		repoName: repo,
		options:  options,
	}, nil
}

func (x *ModuleGitHubRegistry) Download(ctx context.Context, module *Module, options *ModuleRegistryDownloadOptions) (string, error) {
	return downloadModule(ctx, x.buildRegistryUrl(), module, options)
}

func (x *ModuleGitHubRegistry) GetLatestVersion(ctx context.Context, module *Module) (*Module, error) {
	metadata, err := x.GetMetadata(ctx, module)
	if err != nil {
		return nil, err
	}
	return NewModule(metadata.Name, metadata.LatestVersion), nil
}

func (x *ModuleGitHubRegistry) GetAllVersion(ctx context.Context, module *Module) ([]*Module, error) {
	metadata, err := x.GetMetadata(ctx, module)
	if err != nil {
		return nil, err
	}
	moduleSlice := make([]*Module, 0)
	for _, version := range metadata.Versions {
		moduleSlice = append(moduleSlice, NewModule(module.Name, version))
	}
	return moduleSlice, nil
}

func (x *ModuleGitHubRegistry) GetMetadata(ctx context.Context, module *Module) (*ModuleMetadata, error) {
	return getModuleMeta(ctx, x.buildRegistryUrl(), module)
}

func (x *ModuleGitHubRegistry) GetSupplement(ctx context.Context, module *Module) (*ModuleSupplement, error) {
	return getModuleSupplement(ctx, x.buildRegistryUrl(), module)
}

func (x *ModuleGitHubRegistry) buildModuleRegistryDownloadDirectory() string {
	return filepath.Join(x.options.DownloadWorkspace, "registry/github", x.owner, x.repoName)
}

func (x *ModuleGitHubRegistry) List(ctx context.Context) ([]*Module, error) {
	localRegistryDirectoryPath := x.buildModuleRegistryDownloadDirectory()
	err := http_client.NewGitHubRepoDownloader().Download(ctx, &http_client.GitHubRepoDownloaderOptions{
		Owner:             x.owner,
		Repo:              x.repoName,
		DownloadDirectory: localRegistryDirectoryPath,
		CacheTime:         pointer.ToDurationPointer(time.Hour),
		// TODO no ProgressListener, is ok?
	})
	if err != nil {
		return nil, err
	}
	// TODO create local module registry
	registry, err := NewModuleLocalRegistry(localRegistryDirectoryPath, x.buildRegistryRepoFullName())
	if err != nil {
		return nil, err
	}
	return registry.List(ctx)
}

func (x *ModuleGitHubRegistry) Search(ctx context.Context, keyword string) ([]*Module, error) {
	allModuleSlice, err := x.List(ctx)
	if err != nil {
		return nil, err
	}
	keyword = strings.ToLower(keyword)
	searchResultSlice := make([]*Module, 0)
	for _, module := range allModuleSlice {
		if strings.Contains(strings.ToLower(module.Name), keyword) {
			searchResultSlice = append(searchResultSlice, module)
		}
	}
	return searchResultSlice, nil
}

func (x *ModuleGitHubRegistry) buildRegistryRepoFullName() string {
	return fmt.Sprintf("%s/%s", x.owner, x.repoName)
}

func (x *ModuleGitHubRegistry) buildRegistryUrl() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/", x.owner, x.repoName)
}

func (x *ModuleGitHubRegistry) CheckUpdate(ctx context.Context, module *Module) (*Module, error) {
	if module.IsLatestVersion() {
		return nil, nil
	}

	metadata, err := x.GetMetadata(ctx, module)
	if err != nil {
		return nil, err
	}
	// already is latest version
	if module.Version == metadata.LatestVersion {
		return nil, nil
	}

	return NewModule(module.Name, metadata.LatestVersion), nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func downloadModule(ctx context.Context, registryUrl string, module *Module, options *ModuleRegistryDownloadOptions) (string, error) {

	if err := formatModuleVersion(ctx, registryUrl, module); err != nil {
		return "", err
	}

	supplement, err := getModuleSupplement(ctx, registryUrl, module)
	if err != nil {
		return "", err
	}

	if err := utils.EnsureDirectoryNotExists(options.ModuleDownloadDirectoryPath); err != nil {
		return "", err
	}

	githubReleaseAssertName := supplement.PackageName

	// TODO optimization, Improve compatibility
	// example: https://github.com/selefra/rules-aws-misconfiguration-s3/releases/download/v0.0.1/rules-aws-misconfigure-s3.zip
	// example: https://github.com/selefra/rules-aws-misconfiguration-s3/archive/refs/tags/v0.0.2.zip
	githubReleaseAssertURL := supplement.Source + "/releases/download/" + module.Version + "/" + githubReleaseAssertName + ".zip"

	// TODO checksum
	//if !pointer.FromBoolPointerOrDefault(options.SkipVerify, true) {
	//	checksum, err := supplement.Checksums.selectChecksums()
	//	if err != nil {
	//		return "", err
	//	}
	//	githubReleaseAssertURL += "?checksum=sha256:" + checksum
	//}

	// Example URL:
	// https://github.com/selefra/rules-aws-misconfiguration-s3/archive/refs/tags/v0.0.2.zip
	// https://github.com/selefra/rules-aws-misconfigure-s3/releases/download/v0.0.4/rules-aws-misconfigure-s3.zip
	err = http_client.DownloadToDirectory(ctx, options.ModuleDownloadDirectoryPath, githubReleaseAssertURL, options.ProgressTracker)
	if err != nil {
		return "", err
	}

	// search download file
	if utils.Exists(options.ModuleDownloadDirectoryPath) {
		return filepath.Join(options.ModuleDownloadDirectoryPath), nil
	}

	return "", fmt.Errorf("module %s download failed", supplement.PackageName)
}

func formatModuleVersion(ctx context.Context, registryUrl string, module *Module) error {
	if !module.IsLatestVersion() {
		return nil
	}
	meta, err := getModuleMeta(ctx, registryUrl, module)
	if err != nil {
		return err
	}
	module.Version = meta.LatestVersion
	return nil
}

func buildModuleDownloadPath(downloadWorkspace string, module *Module) string {
	return fmt.Sprintf("%s/%s/%s/%s", downloadWorkspace, ModulesListDirectoryName, module.Name, module.Version)
}

func getModuleMeta(ctx context.Context, registryUrl string, module *Module) (*ModuleMetadata, error) {
	metadataUrl := fmt.Sprintf("%s/main/%s/%s/%s", registryUrl, ModulesListDirectoryName, module.Name, MetaDataFileName)
	getYaml, err := http_client.GetYaml[*ModuleMetadata](ctx, metadataUrl)
	if err != nil {
		return nil, err
	}
	return getYaml, nil
}

func getModuleSupplement(ctx context.Context, registryUrl string, module *Module) (*ModuleSupplement, error) {
	supplementUrl := fmt.Sprintf("%s/main/%s/%s/%s/%s", registryUrl, ModulesListDirectoryName, module.Name, module.Version, SupplementFileName)
	supplement, err := http_client.GetYaml[*ModuleSupplement](ctx, supplementUrl)
	if err != nil {
		return nil, err
	}
	return supplement, err
}

// ------------------------------------------------- --------------------------------------------------------------------

//func GetHomeModulesPath(modules string, org string) (string, error) {
//	path, _, err := Home()
//	if err != nil {
//		return "", err
//	}
//	modulesPath := filepath.Join(path, "download/modules")
//	err = ModulesUpdate(modules, modulesPath, org)
//	if err != nil {
//		return "", err
//	}
//	_, err = os.Stat(modulesPath)
//	if err != nil {
//		return "", err
//	}
//	if errors.Is(err, os.ErrNotExist) {
//		err = os.MkdirAll(modulesPath, 0755)
//		if err != nil {
//			return "", err
//		}
//	}
//	return modulesPath, nil
//}

//func ModulesUpdate(modulesName string, modulesPath string, org string) error {
//	_, config, err := Home()
//	if err != nil {
//		return err
//	}
//	c, err := os.ReadFile(config)
//	if err != nil {
//		return err
//	}
//	var configMap = make(map[string]string)
//	err = json.Unmarshal(c, &configMap)
//	if err != nil {
//		return err
//	}
//
//	if org != "" {
//		url := "https://" + global.SERVER + "/cli/download/" + org + "/" + global.Token() + "/" + modulesName + ".zip"
//		_, err := os.Stat(filepath.Join(modulesPath, modulesName))
//		if err == nil {
//			err = os.RemoveAll(filepath.Join(modulesPath, modulesName))
//			if err != nil {
//				return err
//			}
//		}
//		err = modules.DownloadModule(url, filepath.Join(modulesPath, modulesName))
//		if err != nil {
//			return err
//		}
//		return nil
//	} else {
//		if LatestVersion == "" {
//			metadata, err := getModulesMetadata(context.Background(), modulesName)
//			if err != nil {
//				return err
//			}
//			LatestVersion = metadata.LatestVersion
//		}
//		if err != nil {
//			return err
//		}
//		_, e := os.Stat(filepath.Join(modulesPath, modulesName))
//		if configMap["modules"+"/"+modulesName] == LatestVersion && e == nil {
//			return nil
//		} else {
//			supplement, err := getModulesModulesSupplement(context.Background(), modulesName, LatestVersion)
//			if err != nil {
//				return err
//			}
//			url := supplement.Source + "/releases/download/" + LatestVersion + "/" + modulesName + ".zip"
//			err = os.RemoveAll(filepath.Join(modulesPath, modulesName))
//			if err != nil {
//				return err
//			}
//			err = modules.DownloadModule(url, modulesPath)
//			if err != nil {
//				return err
//			}
//			configMap["modules"+"/"+modulesName] = LatestVersion
//			c, err := json.Marshal(configMap)
//			if err != nil {
//				return err
//			}
//			err = os.Remove(config)
//			if err != nil {
//				return err
//			}
//			err = os.WriteFile(config, c, 0644)
//		}
//	}
//	return nil
//}

//func GetPathBySource(source, version string) string {
//	_, config, err := Home()
//	if err != nil {
//		return ""
//	}
//	c, err := os.ReadFile(config)
//	if err != nil {
//		return ""
//	}
//	var configMap = make(map[string]string)
//	err = json.Unmarshal(c, &configMap)
//	if err != nil {
//		return ""
//	}
//
//	ss := strings.SplitN(source, "@", 2)
//
//	return configMap[ss[0]+"@"+version]
//}

//type ModuleMetadata struct {
//	Name          string   `json:"name" yaml:"name"`
//	LatestVersion string   `json:"latest-version" yaml:"latest-version"`
//	LatestUpdate  string   `json:"latest-updated" yaml:"latest-updated"`
//	Introduction  string   `json:"introduction" yaml:"introduction"`
//	Versions      []string `json:"versions" yaml:"versions"`
//}
//
//type ModulesSupplement struct {
//	PackageName string `json:"package-name" yaml:"package-name"`
//	Source      string `json:"source" yaml:"source"`
//	Checksums   string `json:"checksums" yaml:"checksums"`
//}

// ------------------------------------------------- --------------------------------------------------------------------
