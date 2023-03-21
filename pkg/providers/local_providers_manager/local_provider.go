package local_providers_manager

import (
	"github.com/selefra/selefra/pkg/registry"
	"time"
)

// ------------------------------------------------- --------------------------------------------------------------------

type LocalProviderSource int

const (
	LocalProviderSourceUnknown LocalProviderSource = iota
	LocalProviderSourceGitHubRegistry
	LocalProviderSourceLocalRegistry
)

// ------------------------------------------------- --------------------------------------------------------------------

type LocalProvider struct {
	*registry.Provider

	// provider executable file path
	ExecutableFilePath string `json:"executable-file-path"`

	// only support sha256 current
	Checksum string `json:"checksum"`

	// The installation time of the provider
	InstallTime time.Time `json:"install-time"`

	// Where is this provider obtained from
	Source LocalProviderSource `json:"source"`

	// Source dependent context
	SourceContext string `json:"source-context"`
}

func NewLocalProvider(providerName, providerVersion string) *LocalProvider {
	return &LocalProvider{
		Provider: registry.NewProvider(providerName, providerVersion),
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// LocalProviderVersions Indicates all versions and related information of a provider
type LocalProviderVersions struct {

	// provider name
	ProviderName string `json:"provider-name"`

	// All versions of the provider
	ProviderVersionMap map[string]*LocalProvider `json:"provider-version-map"`
}

func NewLocalProviderVersions(providerName string) *LocalProviderVersions {
	return &LocalProviderVersions{
		ProviderName:       providerName,
		ProviderVersionMap: make(map[string]*LocalProvider, 0),
	}
}

func (x *LocalProviderVersions) AddLocalProvider(localProvider *LocalProvider) {
	x.ProviderVersionMap[localProvider.Version] = localProvider
}

// ------------------------------------------------- --------------------------------------------------------------------
