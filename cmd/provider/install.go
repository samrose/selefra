package provider

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/selefra/selefra/pkg/version"
	"github.com/spf13/cobra"
)

func newCmdProviderInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "install",
		Short:            "Install one or more providers, for example: selefra provider install aws",
		Long:             "Install one or more providers, for example: selefra provider install aws",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			downloadDirectory, err := config.GetDefaultDownloadCacheDirectory()
			if err != nil {
				return err
			}
			return Install(ctx, downloadDirectory, args...)
		},
	}

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

func Install(ctx context.Context, downloadWorkspace string, requiredProviders ...string) (err error) {

	if len(requiredProviders) == 0 {
		cli_ui.Errorf("Please specify one or more providers to install, for example: selefra provider install aws \n")
		return nil
	}

	manager, err := local_providers_manager.NewLocalProvidersManager(downloadWorkspace)
	if err != nil {
		return err
	}
	for _, nameAndVersionString := range requiredProviders {
		nameAndVersion := version.ParseNameAndVersion(nameAndVersionString)
		messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
			e := cli_ui.PrintDiagnostics(message)
			if err == nil {
				err = e
			}
		})
		manager.InstallProvider(ctx, &local_providers_manager.InstallProvidersOptions{
			RequiredProvider: local_providers_manager.NewLocalProvider(nameAndVersion.Name, nameAndVersion.Version),
			MessageChannel:   messageChannel,
		})
		messageChannel.ReceiverWait()
	}
	return err
}

//func install(ctx context.Context, args []string) error {
//	configYaml, err := config.GetConfig()
//	if err != nil {
//		cli_ui.Errorln(err.Error())
//		return err
//	}
//
//	namespace, _, err := utils.Home()
//	if err != nil {
//		cli_ui.Errorln(err.Error())
//		return nil
//	}
//
//	provider := registry.NewProviderRegistry(namespace)
//	for _, s := range args {
//		splitArr := strings.Split(s, "@")
//		var name string
//		var version string
//		if len(splitArr) > 1 {
//			name = splitArr[0]
//			version = splitArr[1]
//		} else {
//			name = splitArr[0]
//			version = "latest"
//		}
//		pr := registry.Provider{
//			Name:    name,
//			Version: version,
//			Source:  "",
//		}
//		p, err := provider.Download(ctx, pr, true)
//		continueFlag := false
//		for _, provider := range configYaml.Selefra.ProviderDecls {
//			providerName := *provider.Source
//			if strings.ToLower(providerName) == strings.ToLower(p.Name) && strings.ToLower(provider.Version) == strings.ToLower(p.Version) {
//				continueFlag = true
//				break
//			}
//		}
//		if continueFlag {
//			cli_ui.Warningln(fmt.Sprintf("ProviderBlock %s@%s already installed", p.Name, p.Version))
//			continue
//		}
//		if err != nil {
//			cli_ui.Errorf("Installed %s@%s failed：%s", p.Name, p.Version, err.Error())
//			return nil
//		} else {
//			cli_ui.Infof("Installed %s@%s verified", p.Name, p.Version)
//		}
//		cli_ui.Infof("Synchronization %s@%s's config...", p.Name, p.Version)
//		plug, err := plugin.NewManagedPlugin(p.Filepath, p.Name, p.Version, "", nil)
//		if err != nil {
//			cli_ui.Errorf("Synchronization %s@%s's config failed：%s", p.Name, p.Version, err.Error())
//			return nil
//		}
//
//		plugProvider := plug.Provider()
//		storageOpt := pgstorage.DefaultPgStorageOpts()
//		opt, err := json.Marshal(storageOpt)
//		initRes, err := plugProvider.Init(ctx, &shard.ProviderInitRequest{
//			ModuleLocalDirectory: pointer.ToStringPointer(global.WorkSpace()),
//			Storage: &shard.Storage{
//				Type:           0,
//				StorageOptions: opt,
//			},
//			IsInstallInit:  pointer.TruePointer(),
//			ProviderConfig: pointer.ToStringPointer(""),
//		})
//
//		if err != nil {
//			cli_ui.Errorln(err.Error())
//			return nil
//		}
//
//		if initRes != nil && initRes.Diagnostics != nil {
//			err := cli_ui.PrintDiagnostic(initRes.Diagnostics.GetDiagnosticSlice())
//			if err != nil {
//				return nil
//			}
//		}
//
//		res, err := plugProvider.GetProviderInformation(ctx, &shard.GetProviderInformationRequest{})
//		if err != nil {
//			cli_ui.Errorf("Synchronization %s@%s's config failed：%s", p.Name, p.Version, err.Error())
//			return nil
//		}
//		cli_ui.Infof("Synchronization %s@%s's config successful", p.Name, p.Version)
//		err = tools.AppendProviderDecl(p, configYaml, version)
//		if err != nil {
//			cli_ui.Errorln(err.Error())
//			return nil
//		}
//		hasProvider := false
//		for _, Node := range configYaml.Providers.Content {
//			if Node.Kind == yaml.ScalarNode && Node.Value == p.Name {
//				hasProvider = true
//				break
//			}
//		}
//		if !hasProvider {
//			err = tools.SetProviderTmpl(res.DefaultConfigTemplate, p, configYaml)
//		}
//		if err != nil {
//			cli_ui.Errorf("set %s@%s's config failed：%s", p.Name, p.Version, err.Error())
//			return nil
//		}
//	}
//
//	str, err := yaml.Marshal(configYaml)
//	if err != nil {
//		cli_ui.Errorln(err.Error())
//		return nil
//	}
//	path, err := config.GetConfigPath()
//	if err != nil {
//		cli_ui.Errorln(err.Error())
//		return nil
//	}
//	err = os.WriteFile(path, str, 0644)
//	return nil
//}
