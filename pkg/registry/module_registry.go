package registry

import (
	"context"
	"github.com/hashicorp/go-getter"
	"github.com/selefra/selefra/pkg/version"
)

// ------------------------------------------------- --------------------------------------------------------------------

type Module struct {
	*version.NameAndVersion
}

func ParseModule(moduleNameAndVersion string) *Module {
	return &Module{
		NameAndVersion: version.ParseNameAndVersion(moduleNameAndVersion),
	}
}

func NewModule(moduleName, moduleVersion string) *Module {
	return &Module{
		NameAndVersion: version.NewNameAndVersion(moduleName, moduleVersion),
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

type ModuleMetadata struct {
	Name          string   `json:"name" yaml:"name"`
	LatestVersion string   `json:"latest-version" yaml:"latest-version"`
	LatestUpdate  string   `json:"latest-updated" yaml:"latest-updated"`
	Introduction  string   `json:"introduction" yaml:"introduction"`
	Versions      []string `json:"versions" yaml:"versions"`
}

func (x *ModuleMetadata) HasVersion(version string) bool {
	for _, v := range x.Versions {
		if v == version {
			return true
		}
	}
	return false
}

type ModuleSupplement struct {
	PackageName string `json:"package-name" yaml:"package-name"`
	Source      string `json:"source" yaml:"source"`
	Checksums   string `json:"checksum" yaml:"checksum"`
}

// ------------------------------------------------- --------------------------------------------------------------------

type ModuleRegistryDownloadOptions struct {

	// Which directory to save the downloaded module to
	ModuleDownloadDirectoryPath string

	// Whether to skip authentication. If the authentication is skipped, the checksum of the downloaded file is not verified
	SkipVerify *bool

	// Downloading can be time-consuming, so you can set up a monitor to track progress
	ProgressTracker getter.ProgressTracker
}

// ModuleRegistry Used to represent a registry for a Module repository
type ModuleRegistry interface {

	// CheckUpdate Used to check whether a module has a newer version than the current one
	CheckUpdate(ctx context.Context, module *Module) (*Module, error)

	// Download the module to a local registryDirectory
	Download(ctx context.Context, module *Module, options *ModuleRegistryDownloadOptions) (string, error)

	//// DeleteModule Delete the module downloaded locally
	//DeleteModule(localModuleInfo *LocalModule) error

	// GetLatestVersion Gets the latest version of a given module
	GetLatestVersion(ctx context.Context, module *Module) (*Module, error)

	// GetAllVersion Gets all versions of a given module
	GetAllVersion(ctx context.Context, module *Module) ([]*Module, error)

	GetMetadata(ctx context.Context, module *Module) (*ModuleMetadata, error)

	GetSupplement(ctx context.Context, module *Module) (*ModuleSupplement, error)

	// List Lists all modules installed locally
	List(ctx context.Context) ([]*Module, error)

	// Search Searches the remote registry for modules containing the given keyword
	Search(ctx context.Context, keyword string) ([]*Module, error)
}

// ------------------------------------------------- --------------------------------------------------------------------
