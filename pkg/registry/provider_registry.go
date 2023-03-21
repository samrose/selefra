package registry

import (
	"context"
	"errors"
	"github.com/hashicorp/go-getter"
	"github.com/selefra/selefra/pkg/version"
	"runtime"
)

// ------------------------------------------------- --------------------------------------------------------------------

const (
	ProvidersListDirectoryName = "provider"
	MetaDataFileName           = "metadata.yaml"
	SupplementFileName         = "supplement.yaml"
)

// ------------------------------------------------- --------------------------------------------------------------------

type Provider struct {
	*version.NameAndVersion
}

func NewProvider(providerName, providerVersion string) *Provider {
	return &Provider{
		NameAndVersion: version.NewNameAndVersion(providerName, providerVersion),
	}
}

// ParseProvider example: aws@v0.0.1
func ParseProvider(providerNameAndVersion string) *Provider {
	nameAndVersion := version.ParseNameAndVersion(providerNameAndVersion)
	return &Provider{
		NameAndVersion: nameAndVersion,
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

type ProviderMetadata struct {
	Name          string   `json:"name" yaml:"name"`
	LatestVersion string   `json:"latest-version" yaml:"latest-version"`
	LatestUpdate  string   `json:"latest-updated" yaml:"latest-updated"`
	Introduction  string   `json:"introduction" yaml:"introduction"`
	Versions      []string `json:"versions" yaml:"versions"`
}

func (x *ProviderMetadata) HasVersion(version string) bool {
	for _, v := range x.Versions {
		if v == version {
			return true
		}
	}
	return false
}

type ProviderSupplement struct {
	PackageName string    `json:"package-name" yaml:"package-name"`
	Source      string    `json:"source" yaml:"source"`
	Checksums   Checksums `json:"checksums" yaml:"checksums"`
}

type Checksums struct {
	LinuxArm64   string `json:"linux_arm64" yaml:"linux_arm64"`
	LinuxAmd64   string `json:"linux_amd64" yaml:"linux_amd64"`
	WindowsArm64 string `json:"windows_arm64" yaml:"windows_arm64"`
	WindowsAmd64 string `json:"windows_amd64" yaml:"windows_amd64"`
	DarwinArm64  string `json:"darwin_arm64" yaml:"darwin_arm64"`
	DarwinAmd64  string `json:"darwin_amd64" yaml:"darwin_amd64"`
}

func (x *Checksums) selectChecksums() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return x.DarwinAmd64, nil
		case "arm64":
			return x.DarwinArm64, nil
		default:
			return "", errors.New("unsupported arch")
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return x.WindowsAmd64, nil
		case "arm64":
			return x.WindowsArm64, nil
		default:
			return "", errors.New("unsupported arch")
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return x.LinuxAmd64, nil
		case "arm64":
			return x.LinuxArm64, nil
		default:
			return "", errors.New("unsupported arch")
		}
	default:
		return "", errors.New("unsupported os")
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderRegistryDownloadOptions Some options when downloading the provider
type ProviderRegistryDownloadOptions struct {

	// Which directory to save the downloaded provider to
	ProviderDownloadDirectoryPath string

	// Whether to skip authentication. If the authentication is skipped, the checksum of the downloaded file is not verified
	SkipVerify *bool

	// Downloading can be time-consuming, so you can set up a monitor to track progress
	ProgressTracker getter.ProgressTracker
}

// ProviderRegistry Used to represent the registry of a provider
type ProviderRegistry interface {

	// CheckUpdate Check whether the given provider has a newer version
	CheckUpdate(ctx context.Context, provider *Provider) (*Provider, error)

	// GetLatestVersion Gets the latest version of the specified provider
	GetLatestVersion(ctx context.Context, provider *Provider) (*Provider, error)

	// GetAllVersion Gets all versions of the given provider
	GetAllVersion(ctx context.Context, provider *Provider) ([]*Provider, error)

	GetMetadata(ctx context.Context, provider *Provider) (*ProviderMetadata, error)

	GetSupplement(ctx context.Context, provider *Provider) (*ProviderSupplement, error)

	// Download the provider of the given version
	Download(ctx context.Context, provider *Provider, options *ProviderRegistryDownloadOptions) (string, error)

	// Search for providers in registry by keyword
	Search(ctx context.Context, keyword string) ([]*Provider, error)

	// List Lists all providers on registry
	List(ctx context.Context) ([]*Provider, error)
}

// ------------------------------------------------- --------------------------------------------------------------------
