package module_loader

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/md5_util"
	"github.com/selefra/selefra/pkg/http_client"
	"github.com/selefra/selefra/pkg/modules/module"
	"path/filepath"
)

// TODO Need to test

// ------------------------------------------------- --------------------------------------------------------------------

type S3BucketModuleLoaderOptions struct {
	*ModuleLoaderOptions

	S3BucketURL string
}

// ------------------------------------------------- --------------------------------------------------------------------

type S3BucketModuleLoader struct {
	options                     *S3BucketModuleLoaderOptions
	moduleDownloadDirectoryPath string
}

var _ ModuleLoader[*S3BucketModuleLoaderOptions] = &S3BucketModuleLoader{}

func NewS3BucketModuleLoader(options *S3BucketModuleLoaderOptions) (*S3BucketModuleLoader, error) {

	directoryName, err := md5_util.Md5String(options.S3BucketURL)
	if err != nil {
		// TODO
		return nil, err
	}
	moduleDownloadDirectoryPath := filepath.Join(options.DownloadDirectory, DownloadModulesDirectoryName, string(ModuleLoaderTypeS3Bucket), directoryName)

	return &S3BucketModuleLoader{
		options:                     options,
		moduleDownloadDirectoryPath: moduleDownloadDirectoryPath,
	}, nil
}

func (x *S3BucketModuleLoader) Name() ModuleLoaderType {
	return ModuleLoaderTypeS3Bucket
}

func (x *S3BucketModuleLoader) Options() *S3BucketModuleLoaderOptions {
	return x.options
}

func (x *S3BucketModuleLoader) Load(ctx context.Context) (*module.Module, bool) {

	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
	}()

	// step 01. Download and decompress
	err := http_client.DownloadToDirectory(ctx, x.options.S3BucketURL, x.moduleDownloadDirectoryPath, x.options.ProgressTracker)
	if err != nil {
		// TODO
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(""))
		return nil, false
	}

	// step 02. The download is decompressed and converted to loading from the local path
	localDirectoryModuleLoaderOptions := &LocalDirectoryModuleLoaderOptions{
		// TODO
		//ModuleLoaderOptions: x.options.ModuleLoaderOptions.Copy(),
		ModuleDirectory: x.moduleDownloadDirectoryPath,
	}
	loader, err := NewLocalDirectoryModuleLoader(localDirectoryModuleLoaderOptions)
	if err != nil {
		// TODO
		localDirectoryModuleLoaderOptions.MessageChannel.SenderWaitAndClose()
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(""))
		return nil, false
	}

	return loader.Load(ctx)
}
