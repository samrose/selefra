package provider

import (
	"context"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/spf13/cobra"
)

func newCmdProviderRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "remove",
		Short:            "Remove providers one or more from the download cache, for example: selefra provider remove aws@v0.0.1",
		Long:             "Remove providers one or more from the download cache, for example: selefra provider remove aws@v0.0.1",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, names []string) error {
			downloadDirectory, err := config.GetDefaultDownloadCacheDirectory()
			if err != nil {
				return err
			}
			return Remove(cmd.Context(), downloadDirectory, names...)
		},
	}

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

func Remove(ctx context.Context, downloadWorkspace string, names ...string) error {
	manager, err := local_providers_manager.NewLocalProvidersManager(downloadWorkspace)
	if err != nil {
		return err
	}
	d := manager.RemoveProviders(ctx, names...)
	return cli_ui.PrintDiagnostics(d)
}

//func Remove(names []string) error {
//	argsMap := make(map[string]bool)
//	for i := range names {
//		argsMap[names[i]] = true
//	}
//	deletedMap := make(map[string]bool)
//	cof, err := config.GetConfig()
//	if err != nil {
//		return err
//	}
//	namespace, _, err := utils.Home()
//	if err != nil {
//		return err
//	}
//	provider := registry.NewProviderRegistry(namespace)
//
//	for _, p := range cof.Selefra.ProviderDecls {
//		name := *p.Source
//		path := utils.GetPathBySource(*p.Source, p.Version)
//		prov := registry.ProviderBinary{
//			Provider: registry.Provider{
//				Name:    name,
//				Version: p.Version,
//				Source:  "",
//			},
//			Filepath: path,
//		}
//		if !argsMap[p.Name] || deletedMap[p.Path] {
//			break
//		}
//
//		err := provider.DeleteProvider(prov)
//		if err != nil {
//			if !errors.Is(err, os.ErrNotExist) {
//				cli_ui.Warningf("Failed to remove  %s: %s", p.Name, err.Error())
//			}
//		}
//		_, jsonPath, err := utils.Home()
//		if err != nil {
//			return err
//		}
//		c, err := os.ReadFile(jsonPath)
//		if err == nil {
//			var configMap = make(map[string]string)
//			err = json.Unmarshal(c, &configMap)
//			if err != nil {
//				return err
//			}
//			delete(configMap, *p.Source+"@"+p.Version)
//			c, err = json.Marshal(configMap)
//			if err != nil {
//				return err
//			}
//			err = os.Remove(jsonPath)
//			if err != nil {
//				return err
//			}
//			err = os.WriteFile(jsonPath, c, 0644)
//			if err != nil {
//				return err
//			}
//			deletedMap[path] = true
//		}
//		cli_ui.Infof("Removed %s success", *p.Source)
//	}
//	return nil
//}
