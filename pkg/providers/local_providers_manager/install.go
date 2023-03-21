package local_providers_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/pointer"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"time"
)

type InstallProvidersOptions struct {

	// What are the providers required to be installed
	RequiredProvider *LocalProvider

	// Used to receive messages in real time
	MessageChannel *message.Channel[*schema.Diagnostics]

	ProgressTracker getter.ProgressTracker
}

func (x *LocalProvidersManager) InstallProvider(ctx context.Context, options *InstallProvidersOptions) {

	defer func() {
		options.MessageChannel.SenderWaitAndClose()
	}()

	path := x.buildLocalProviderVersionPath(options.RequiredProvider.Name, options.RequiredProvider.Version)
	if utils.Exists(path) {
		options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("Provider %s in directory %s already installed, remove it first", options.RequiredProvider.String(), path))
		return
	}

	// check require provider & version exist
	metadata, err := x.providerRegistry.GetMetadata(ctx, registry.NewProvider(options.RequiredProvider.Name, options.RequiredProvider.Version))
	if err != nil {
		options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("Get provider %s metadata error: %s", options.RequiredProvider.String(), err.Error()))
		return
	}

	// parse install version
	var version string
	if !options.RequiredProvider.IsLatestVersion() {
		if !metadata.HasVersion(options.RequiredProvider.Version) {
			report := fmt.Sprintf("Provider %s does not exist, can not install it, I'm very sorry.", options.RequiredProvider.String())
			options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(report))
			return
		}
		version = options.RequiredProvider.Version
	} else {
		version = metadata.LatestVersion
		path := x.buildLocalProviderVersionPath(options.RequiredProvider.Name, version)
		if utils.Exists(path) {
			options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Provider %s latest version has been installed on %s", options.RequiredProvider.Name, path))
			return
		}
	}

	// Download the provider executable file
	providerVersionPath := x.buildLocalProviderVersionPath(options.RequiredProvider.Name, version)
	downloadOptions := &registry.ProviderRegistryDownloadOptions{
		ProviderDownloadDirectoryPath: providerVersionPath,
		SkipVerify:                    pointer.TruePointer(),
		ProgressTracker:               options.ProgressTracker,
	}
	providerExecuteFilePath, err := x.providerRegistry.Download(ctx, registry.NewProvider(options.RequiredProvider.Name, version), downloadOptions)
	if err != nil {
		options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("Install provider %s in directory %s failed: %s", options.RequiredProvider.String(), utils.AbsPath(providerVersionPath), err.Error()))
		return
	}
	options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Install provider %s in directory %s success", options.RequiredProvider.String(), utils.AbsPath(providerVersionPath)))

	// Construct metadata
	localProvider := LocalProvider{
		Provider:           registry.NewProvider(options.RequiredProvider.Name, version),
		ExecutableFilePath: providerExecuteFilePath,
		Checksum:           "",
		InstallTime:        time.Now(),
		Source:             LocalProviderSourceGitHubRegistry,
	}
	marshal, err := json.Marshal(localProvider)
	if err != nil {
		options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("On install provider %s, json marshal local provider error: %s", options.RequiredProvider.String(), err.Error()))
		return
	}
	metaFilePath := x.buildLocalProviderVersionMetaFilePath(options.RequiredProvider.Name, version)
	err = os.WriteFile(metaFilePath, marshal, os.ModePerm)
	if err != nil {
		options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("On install provider %s, save provider version meta in file %s error: %s", options.RequiredProvider.String(), metaFilePath, err.Error()))
		return
	}
	return
}
