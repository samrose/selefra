package provider

//import (
//	"context"
//	"github.com/selefra/selefra/cli_ui"
//	"github.com/selefra/selefra/cmd/fetch"
//	"github.com/selefra/selefra/cmd/tools"
//	"github.com/selefra/selefra/config"
//	"github.com/selefra/selefra/global"
//	"github.com/selefra/selefra/pkg/registry"
//	"github.com/selefra/selefra/pkg/utils"
//	"github.com/spf13/cobra"
//)
//
//func newCmdProviderUpdate() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:              "update",
//		Short:            "Upgrade one or more plugins",
//		Long:             "Upgrade one or more plugins",
//		PersistentPreRun: global.DefaultWrappedInit(),
//		RunE: func(cmd *cobra.Command, args []string) error {
//			return update(cmd.Context(), args)
//		},
//	}
//
//	cmd.SetHelpFunc(cmd.HelpFunc())
//	return cmd
//}
//
//func update(ctx context.Context, args []string) error {
//	err := config.IsSelefra()
//	if err != nil {
//		cli_ui.Errorln(err.Error())
//		return err
//	}
//	argsMap := make(map[string]bool)
//	for i := range args {
//		argsMap[args[i]] = true
//	}
//	rootConfig, err := config.GetConfig()
//	if err != nil {
//		return err
//	}
//	namespace, _, err := utils.Home()
//	if err != nil {
//		return err
//	}
//	provider := registry.NewProviderRegistry(namespace)
//	for _, decl := range rootConfig.Selefra.ProviderDecls {
//		prov := registry.ProviderBinary{
//			Provider: registry.Provider{
//				Name:    decl.Name,
//				Version: decl.Version,
//				Source:  "",
//			},
//			Filepath: decl.Path,
//		}
//		if len(args) != 0 && !argsMap[decl.Name] {
//			break
//		}
//
//		pp, err := provider.CheckUpdate(ctx, prov)
//		if err != nil {
//			return err
//		}
//		decl.Path = pp.Filepath
//		decl.Version = pp.Version
//
//		for _, prvd := range tools.ProvidersByID(rootConfig, decl.Name) {
//			err = fetch.Fetch(ctx, decl, prvd)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
