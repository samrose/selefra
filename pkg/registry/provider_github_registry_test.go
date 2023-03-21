package registry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestProviderGithubRegistry() *ProviderGithubRegistry {
	registry, err := NewProviderGithubRegistry(&ProviderGithubRegistryOptions{
		DownloadWorkspace: "./test_download",
	})
	if err != nil {
		panic(err)
	}
	return registry
}

func TestProviderGithubRegistry_Download(t *testing.T) {
	// TODO
	download, err := newTestProviderGithubRegistry().Download(context.Background(), NewProvider("aws", "v0.0.1"), &ProviderRegistryDownloadOptions{
		ProviderDownloadDirectoryPath: "./test_download/providers/aws/v0.0.1",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, download)
}

func TestProviderGithubRegistry_List(t *testing.T) {
	providerSlice, err := newTestProviderGithubRegistry().List(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, providerSlice)
}

func TestProviderGithubRegistry_Search(t *testing.T) {
	providers, err := newTestProviderGithubRegistry().Search(context.Background(), "a")
	assert.Nil(t, err)
	assert.NotNil(t, providers)
}

func TestProviderGithubRegistry_GetAllVersion(t *testing.T) {
	providers, err := newTestProviderGithubRegistry().GetAllVersion(context.Background(), NewProvider("aws", ""))
	assert.Nil(t, err)
	assert.NotNil(t, providers)
}

func TestProviderGithubRegistry_GetMetadata(t *testing.T) {
	providerMetadata, err := newTestProviderGithubRegistry().GetMetadata(context.Background(), NewProvider("aws", ""))
	assert.Nil(t, err)
	assert.NotNil(t, providerMetadata)
}

func TestProviderGithubRegistry_GetSupplement(t *testing.T) {
	providerSupplement, err := newTestProviderGithubRegistry().GetSupplement(context.Background(), NewProvider("aws", "v0.0.10"))
	assert.Nil(t, err)
	assert.NotNil(t, providerSupplement)
}

func TestProviderGithubRegistry_GetLatestVersion(t *testing.T) {
	providerSupplement, err := newTestProviderGithubRegistry().GetLatestVersion(context.Background(), NewProvider("aws", ""))
	assert.Nil(t, err)
	assert.NotNil(t, providerSupplement)
}

func TestProviderGithubRegistry_CheckUpdate(t *testing.T) {
	providerSupplement, err := newTestProviderGithubRegistry().CheckUpdate(context.Background(), NewProvider("aws", "v0.0.9"))
	assert.Nil(t, err)
	assert.NotNil(t, providerSupplement)
}
