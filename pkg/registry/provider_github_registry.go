package registry

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/pointer"
	"github.com/selefra/selefra/pkg/http_client"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/telemetry"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/songzhibin97/gkit/ternary"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ProviderGithubRegistryDefaultRepoFullName The official registry repository
const ProviderGithubRegistryDefaultRepoFullName = "selefra/registry"

var (
	providerBinarySuffix = ternary.ReturnString(runtime.GOOS == "windows", ".exe", "")
)

//func request(ctx context.Context, method string, _url string, body []byte, headers ...Header) ([]byte, error) {
//	var lastErr error
//	for tryTimes := 0; tryTimes < 5; tryTimes++ {
//		client := &http.Client{}
//		sBody := strings.NewReader(string(body))
//		request, err := http.NewRequestWithContext(ctx, method, _url, sBody)
//		if err != nil {
//			return nil, err
//		}
//		request.Header.Add("Content-Type", "application/json")
//		for _, header := range headers {
//			request.Header.Add(header.Key, header.Value)
//		}
//
//		resp, err := client.Do(request)
//		if err != nil {
//			//return nil, err
//			lastErr = err
//			continue
//		}
//		// just ok
//		defer resp.Body.Close()
//		if resp.StatusCode != http.StatusOK {
//			lastErr = fmt.Errorf("request url %s response code %d not equal 200", _url, resp.StatusCode)
//			continue
//		}
//		rByte, err := ioutil.ReadAll(resp.Body)
//		if err != nil {
//			lastErr = fmt.Errorf("request url %s, read body err : %s", _url, err.Error())
//			continue
//		}
//		return rByte, err
//	}
//	return nil, lastErr
//}

// ------------------------------------------------- --------------------------------------------------------------------

type ProviderGithubRegistryOptions struct {
	DownloadWorkspace    string
	RegistryRepoFullName *string
}

func NewProviderGithubRegistryOptions(downloadWorkspace string, registryRepoFullName ...string) *ProviderGithubRegistryOptions {

	if len(registryRepoFullName) == 0 {
		registryRepoFullName = append(registryRepoFullName, ProviderGithubRegistryDefaultRepoFullName)
	}

	return &ProviderGithubRegistryOptions{
		DownloadWorkspace:    downloadWorkspace,
		RegistryRepoFullName: pointer.ToStringPointer(registryRepoFullName[0]),
	}
}

func (x *ProviderGithubRegistryOptions) Check() *schema.Diagnostics {
	// TODO check params
	return nil
}

// ProviderGithubRegistry provider registry github implementation
type ProviderGithubRegistry struct {
	// The owner of registry's repo
	owner string
	// The name of registry's repo
	repoName string

	options *ProviderGithubRegistryOptions
}

var _ ProviderRegistry = &ProviderGithubRegistry{}

func NewProviderGithubRegistry(options *ProviderGithubRegistryOptions) (*ProviderGithubRegistry, error) {

	// set default registry url
	if options.RegistryRepoFullName == nil {
		options.RegistryRepoFullName = pointer.ToStringPointer(ProviderGithubRegistryDefaultRepoFullName)
	}

	// Parse the full name of the github repository
	owner, repo, err := utils.ParseGitHubRepoFullName(pointer.FromStringPointer(options.RegistryRepoFullName))
	if err != nil {
		return nil, err
	}

	return &ProviderGithubRegistry{
		owner:    owner,
		repoName: repo,
		options:  options,
	}, nil
}

func (x *ProviderGithubRegistry) buildRegistryRepoFullName() string {
	return fmt.Sprintf("%s/%s", x.owner, x.repoName)
}

func (x *ProviderGithubRegistry) buildRegistryUrl() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/", x.owner, x.repoName)
}

//func (x *ProviderGithubRegistry) getProviderDownloadDirectory(providerName, providerVersion string) string {
//	return filepath.Join(x.downloadWorkspace, ProvidersListDirectoryName, providerName, providerVersion)
//}

func (x *ProviderGithubRegistry) buildProviderRegistryDownloadDirectory() string {
	return filepath.Join(x.options.DownloadWorkspace, "registry/github", x.owner, x.repoName)
}

func (x *ProviderGithubRegistry) List(ctx context.Context) ([]*Provider, error) {
	localRegistryDirectoryPath := x.buildProviderRegistryDownloadDirectory()
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
	registryDirectory := filepath.Join(localRegistryDirectoryPath, x.repoName+"-main")
	registry, err := NewProviderLocalRegistry(registryDirectory, x.buildRegistryRepoFullName())
	if err != nil {
		return nil, err
	}
	return registry.List(ctx)
}

func (x *ProviderGithubRegistry) Search(ctx context.Context, keyword string) ([]*Provider, error) {
	allProviderSlice, err := x.List(ctx)
	if err != nil {
		return nil, err
	}
	keyword = strings.ToLower(keyword)
	searchResultSlice := make([]*Provider, 0)
	for _, provider := range allProviderSlice {
		if strings.Contains(strings.ToLower(provider.Name), keyword) {
			searchResultSlice = append(searchResultSlice, provider)
		}
	}
	return searchResultSlice, nil
}

func (x *ProviderGithubRegistry) GetLatestVersion(ctx context.Context, provider *Provider) (*Provider, error) {
	metadata, err := x.getProviderMetadata(ctx, provider)
	if err != nil {
		return nil, err
	}
	return NewProvider(metadata.Name, metadata.LatestVersion), nil
}

func (x *ProviderGithubRegistry) GetAllVersion(ctx context.Context, provider *Provider) ([]*Provider, error) {
	metadata, err := x.getProviderMetadata(ctx, provider)
	if err != nil {
		return nil, err
	}
	providerSlice := make([]*Provider, 0)
	for _, version := range metadata.Versions {
		providerSlice = append(providerSlice, NewProvider(provider.Name, version))
	}
	return providerSlice, nil
}

func (x *ProviderGithubRegistry) CheckUpdate(ctx context.Context, provider *Provider) (*Provider, error) {

	if provider.IsLatestVersion() {
		return nil, nil
	}

	metadata, err := x.getProviderMetadata(ctx, provider)
	if err != nil {
		return nil, err
	}
	// already is latest version
	if provider.Version == metadata.LatestVersion {
		return nil, nil
	}

	return NewProvider(provider.Name, metadata.LatestVersion), nil
}

func (x *ProviderGithubRegistry) getSupplement(ctx context.Context, provider *Provider) (*ProviderSupplement, error) {
	supplementUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s/%s/%s/%s", x.owner, x.repoName, ProvidersListDirectoryName, provider.Name, provider.Version, SupplementFileName)
	supplement, err := http_client.GetYaml[*ProviderSupplement](ctx, supplementUrl)
	if err != nil {
		return nil, err
	}
	return supplement, err
	//downloadUrl := supplement.Supplement.Source + "/releases/download/" + ProviderGithubRegistry.Version + "/" + ProviderGithubRegistry.Name + "_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
}

//func (x *ProviderGithubRegistry) fillVersion(ctx context.Context, provider *Provider, skipVerify bool) error {
//	if provider.Version != "" && provider.Version != "latest" && skipVerify {
//		return nil
//	}
//
//	metadata, err := x.getProviderMetadata(ctx, provider)
//	if err != nil {
//		return err
//	}
//	if provider.Version != "" && provider.Version != "latest" {
//		// check version number exists
//		for _, version := range metadata.Versions {
//			if provider.Version != version {
//				continue
//			}
//			return nil
//		}
//		return errors.New("version not found")
//	}
//	provider.Version = metadata.LatestVersion
//	return nil
//}

func (x *ProviderGithubRegistry) getProviderMetadata(ctx context.Context, provider *Provider) (*ProviderMetadata, error) {
	return getProviderMeta(ctx, x.buildRegistryUrl(), provider)
}

func (x *ProviderGithubRegistry) Download(ctx context.Context, provider *Provider, options *ProviderRegistryDownloadOptions) (string, error) {
	return downloadProvider(ctx, x.buildRegistryUrl(), provider, options)
}

func (x *ProviderGithubRegistry) GetMetadata(ctx context.Context, provider *Provider) (*ProviderMetadata, error) {
	return getProviderMeta(ctx, x.buildRegistryUrl(), provider)
}

func (x *ProviderGithubRegistry) GetSupplement(ctx context.Context, provider *Provider) (*ProviderSupplement, error) {
	return getProviderSupplement(ctx, x.buildRegistryUrl(), provider)
}

// ------------------------------------------------- --------------------------------------------------------------------

func downloadProvider(ctx context.Context, registryUrl string, provider *Provider, options *ProviderRegistryDownloadOptions) (string, error) {

	if err := formatProviderVersion(ctx, registryUrl, provider); err != nil {
		return "", err
	}

	supplement, err := getProviderSupplement(ctx, registryUrl, provider)
	if err != nil {
		return "", err
	}

	if err := utils.EnsureDirectoryNotExists(options.ProviderDownloadDirectoryPath); err != nil {
		return "", err
	}

	githubReleaseAssertName := supplement.PackageName + "_" + strings.Replace(provider.Version, "v", "", 1) + "_" + runtime.GOOS + "_" + runtime.GOARCH

	// TODO optimization, Improve compatibility
	// The providerBinarySuffix depends on the provider repository's CI. If that CI changes the providerBinarySuffix, it must be changed accordingly
	githubReleaseAssertURL := supplement.Source + "/releases/download/" + provider.Version + "/" + githubReleaseAssertName + ".tar.gz"

	if !pointer.FromBoolPointerOrDefault(options.SkipVerify, true) {
		checksum, err := supplement.Checksums.selectChecksums()
		if err != nil {
			return "", err
		}
		githubReleaseAssertURL += "?checksum=sha256:" + checksum
	}

	event := telemetry.NewEvent("provider-install").
		Add("url", githubReleaseAssertURL).
		Add("provider_name", provider.Name).
		Add("provider_version", provider.Version)
	d := telemetry.Submit(ctx, event)
	if utils.IsNotEmpty(d) {
		logger.ErrorF("telemetry provider install, msg = %s", d.String())
	}

	//targetUrl := cli_env.GetSelefraCloudHttpHost() + "/diagnosis.tar.gz?url=" + base64.StdEncoding.EncodeToString([]byte(githubReleaseAssertURL))
	err = http_client.DownloadToDirectory(ctx, options.ProviderDownloadDirectoryPath, githubReleaseAssertURL, options.ProgressTracker)
	//err = http_client.DownloadToDirectory(ctx, options.ProviderDownloadDirectoryPath, targetUrl, options.ProgressTracker)
	if err != nil {
		return "", err
	}

	// search download file
	providerExecuteFilePath := filepath.Join(options.ProviderDownloadDirectoryPath, supplement.PackageName+providerBinarySuffix)
	if utils.Exists(providerExecuteFilePath) {
		return providerExecuteFilePath, nil
	}

	return "", fmt.Errorf("provider %s download failed", supplement.PackageName)
}

func formatProviderVersion(ctx context.Context, registryUrl string, provider *Provider) error {
	if !provider.IsLatestVersion() {
		return nil
	}
	meta, err := getProviderMeta(ctx, registryUrl, provider)
	if err != nil {
		return err
	}
	provider.Version = meta.LatestVersion
	return nil
}

func buildProviderDownloadPath(downloadWorkspace string, provider *Provider) string {
	return fmt.Sprintf("%s/%s/%s/%s", downloadWorkspace, ProvidersListDirectoryName, provider.Name, provider.Version)
}

func getProviderMeta(ctx context.Context, registryUrl string, provider *Provider) (*ProviderMetadata, error) {
	metadataUrl := fmt.Sprintf("%smain/%s/%s/%s", registryUrl, ProvidersListDirectoryName, provider.Name, MetaDataFileName)
	getYaml, err := http_client.GetYaml[*ProviderMetadata](ctx, metadataUrl)
	if err != nil {
		return nil, err
	}
	return getYaml, nil
}

func getProviderSupplement(ctx context.Context, registryUrl string, provider *Provider) (*ProviderSupplement, error) {
	supplementUrl := fmt.Sprintf("%s/main/%s/%s/%s/%s", registryUrl, ProvidersListDirectoryName, provider.Name, provider.Version, SupplementFileName)
	supplement, err := http_client.GetYaml[*ProviderSupplement](ctx, supplementUrl)
	if err != nil {
		return nil, err
	}
	return supplement, err
}

// ------------------------------------------------- --------------------------------------------------------------------
