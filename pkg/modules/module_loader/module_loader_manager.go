package module_loader

import (
	"regexp"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

type ModuleLoaderType string

const (

	// ModuleLoaderTypeInvalid Default value if not set
	ModuleLoaderTypeInvalid ModuleLoaderType = ""

	// ModuleLoaderTypeS3Bucket Load modules from S3 bucket.s
	ModuleLoaderTypeS3Bucket ModuleLoaderType = "s3-bucket-module-loader"

	// ModuleLoaderTypeGitHubRegistry Load the module from GitHub's Registry
	ModuleLoaderTypeGitHubRegistry ModuleLoaderType = "github-registry-module-loader"

	// ModuleLoaderTypeLocalDirectory Load the module from the local directory
	ModuleLoaderTypeLocalDirectory ModuleLoaderType = "local-directory-module-loader"

	ModuleLoaderTypeURL ModuleLoaderType = "url-module-loader"
)

// ------------------------------------------------- --------------------------------------------------------------------

var Pattern = regexp.MustCompile("^[A-Za-z_-]?[\\w\\-_@.]+$")

// NewModuleLoaderBySource Distributed to different module loaders based on load options
func NewModuleLoaderBySource(source string) ModuleLoaderType {
	formatSource := strings.ToLower(source)
	switch {
	case strings.HasPrefix(formatSource, "s3://"):
		return ModuleLoaderTypeS3Bucket
	case strings.HasPrefix(formatSource, "http://") || strings.HasPrefix(formatSource, "https://"):
		return ModuleLoaderTypeURL
	case strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../"):
		return ModuleLoaderTypeLocalDirectory
	case Pattern.MatchString(source):
		return ModuleLoaderTypeGitHubRegistry
	default:
		return ModuleLoaderTypeInvalid
	}
}

// ------------------------------------------------- --------------------------------------------------------------------
