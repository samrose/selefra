package registry

import (
	"context"
	"github.com/selefra/selefra-utils/pkg/pointer"
	"github.com/selefra/selefra/pkg/http_client"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
)

func newTestProviderLocalRegistry() *ProviderLocalRegistry {
	downloadDirectory := "./test_download/registry/github/selefra/registry/"
	err := http_client.NewGitHubRepoDownloader().Download(context.Background(), &http_client.GitHubRepoDownloaderOptions{
		Owner:             "selefra",
		Repo:              "registry",
		DownloadDirectory: downloadDirectory,
		CacheTime:         pointer.ToDurationPointer(time.Hour),
		// TODO no ProgressListener, is ok?
	})
	if err != nil {
		panic(err)
	}
	registryDirectory := filepath.Join(downloadDirectory + "/registry-main")
	registry, err := NewProviderLocalRegistry(registryDirectory)
	if err != nil {
		panic(err)
	}

	registry, err = NewProviderLocalRegistry(registryDirectory)
	if err != nil {
		panic(err)
	}
	return registry
}

func TestProviderLocalRegistry_Download(t *testing.T) {
	download, err := newTestProviderLocalRegistry().Download(context.Background(), NewProvider("aws", "v0.0.1"), &ProviderRegistryDownloadOptions{
		ProviderDownloadDirectoryPath: "./test_download/providers/aws/v0.0.1",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, download)
}

func TestProviderLocalRegistry_List(t *testing.T) {
	providerSlice, err := newTestProviderLocalRegistry().List(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, providerSlice)
}

func TestProviderLocalRegistry_Search(t *testing.T) {
	providers, err := newTestProviderLocalRegistry().Search(context.Background(), "a")
	assert.Nil(t, err)
	assert.NotNil(t, providers)
}

func TestProviderLocalRegistry_GetAllVersion(t *testing.T) {
	providers, err := newTestProviderLocalRegistry().GetAllVersion(context.Background(), NewProvider("aws", ""))
	assert.Nil(t, err)
	assert.NotNil(t, providers)
}

func TestProviderLocalRegistry_GetMetadata(t *testing.T) {
	providerMetadata, err := newTestProviderLocalRegistry().GetMetadata(context.Background(), NewProvider("aws", ""))
	assert.Nil(t, err)
	assert.NotNil(t, providerMetadata)
}

func TestProviderLocalRegistry_GetSupplement(t *testing.T) {
	providerSupplement, err := newTestProviderLocalRegistry().GetSupplement(context.Background(), NewProvider("aws", "v0.0.10"))
	assert.Nil(t, err)
	assert.NotNil(t, providerSupplement)
}

func TestProviderLocalRegistry_GetLatestVersion(t *testing.T) {
	providerSupplement, err := newTestProviderLocalRegistry().GetLatestVersion(context.Background(), NewProvider("aws", ""))
	assert.Nil(t, err)
	assert.NotNil(t, providerSupplement)
}

func TestProviderLocalRegistry_CheckUpdate(t *testing.T) {
	providerSupplement, err := newTestProviderLocalRegistry().CheckUpdate(context.Background(), NewProvider("aws", "v0.0.9"))
	assert.Nil(t, err)
	assert.NotNil(t, providerSupplement)
}
