package module_loader

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/md5_util"
	"github.com/selefra/selefra/pkg/http_client"
	"github.com/selefra/selefra/pkg/modules/module"
	"path/filepath"
)

// ------------------------------------------------- --------------------------------------------------------------------

// URLModuleLoaderOptions Parameter options when creating the URL module loader
type URLModuleLoaderOptions struct {
	*ModuleLoaderOptions

	// Module URL, It's a zip package
	ModuleURL string `json:"module-url" yaml:"module-url"`
}

//func (x *URLModuleLoaderOptions) Copy() *URLModuleLoaderOptions {
//	return &URLModuleLoaderOptions{
//		// TODO
//		//ModuleLoaderOptions: x.ModuleLoaderOptions.Copy(),
//		ModuleURL: x.ModuleURL,
//	}
//}

//func (x *URLModuleLoaderOptions) CopyForURL(moduleURL string) *URLModuleLoaderOptions {
//	options := x.Copy()
//	options.ModuleURL = moduleURL
//	return options
//}

// ------------------------------------------------- --------------------------------------------------------------------

// URLModuleLoader Load the module from a URL, which should be a zipped package that happens to be the module's directory
type URLModuleLoader struct {
	options *URLModuleLoaderOptions

	// Which path to download to
	moduleDownloadDirectoryPath string
}

var _ ModuleLoader[*URLModuleLoaderOptions] = &URLModuleLoader{}

func NewURLModuleLoader(options *URLModuleLoaderOptions) (*URLModuleLoader, error) {

	directoryName, err := md5_util.Md5String(options.ModuleURL)
	if err != nil {
		return nil, err
	}
	moduleDownloadDirectoryPath := filepath.Join(options.DownloadDirectory, DownloadModulesDirectoryName, string(ModuleLoaderTypeURL), directoryName)

	return &URLModuleLoader{
		options:                     options,
		moduleDownloadDirectoryPath: moduleDownloadDirectoryPath,
	}, nil
}

func (x *URLModuleLoader) Name() ModuleLoaderType {
	return ModuleLoaderTypeURL
}

func (x *URLModuleLoader) Load(ctx context.Context) (*module.Module, bool) {

	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
	}()

	// step 01. Download and decompress
	err := http_client.DownloadToDirectory(ctx, x.moduleDownloadDirectoryPath, x.options.ModuleURL, x.options.ProgressTracker)
	if err != nil {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("module load from %s failed, error = %s", x.options.ModuleURL, err.Error()))
		return nil, false
	}

	// send tips
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("download url module %s to local directory %s", x.options.ModuleURL, x.moduleDownloadDirectoryPath))

	// step 02. The download is decompressed and converted to loading from the local path
	localDirectoryModuleLoaderOptions := &LocalDirectoryModuleLoaderOptions{
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source:  x.options.Source,
			Version: x.options.Version,
			// TODO
			//ProgressTracker:   x.ProgressTracker,
			DownloadDirectory: x.options.DownloadDirectory,
			MessageChannel:    x.options.MessageChannel.MakeChildChannel(),
			// The dependency level does not increase
			DependenciesTree: x.options.DependenciesTree,
		},
		ModuleDirectory: x.moduleDownloadDirectoryPath,
	}
	loader, err := NewLocalDirectoryModuleLoader(localDirectoryModuleLoaderOptions)
	if err != nil {
		localDirectoryModuleLoaderOptions.MessageChannel.SenderWaitAndClose()
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("create local directory %s module loader error: %s", x.moduleDownloadDirectoryPath, err.Error()))
		return nil, false
	}

	return loader.Load(ctx)
}

func (x *URLModuleLoader) Options() *URLModuleLoaderOptions {
	return x.options
}
