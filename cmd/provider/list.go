package provider

import (
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/selefra/selefra/pkg/version"
	"github.com/spf13/cobra"
)

func newCmdProviderList() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "list",
		Short:            "List currently installed providers",
		Long:             "List currently installed providers",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, args []string) error {

			downloadWorkspace, err := config.GetDefaultDownloadCacheDirectory()
			if err != nil {
				return err
			}

			return List(downloadWorkspace)
		},
	}

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

func List(downloadWorkspace string) error {

	manager, err := local_providers_manager.NewLocalProvidersManager(downloadWorkspace)
	if err != nil {
		return err
	}
	providers, diagnostics := manager.ListProviders()
	if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
		return err
	}
	if len(providers) == 0 {
		return nil
	}

	table := make([][]string, 0)
	for _, provider := range providers {
		versions := make([]string, 0)
		for versionString := range provider.ProviderVersionMap {
			versions = append(versions, versionString)
		}
		version.Sort(versions)
		for _, versionString := range versions {
			table = append(table, []string{
				provider.ProviderName, versionString, provider.ProviderVersionMap[versionString].ExecutableFilePath,
			})
		}
	}
	cli_ui.ShowTable([]string{"Name", "Version", "Source"}, table, nil, true)

	return nil
}
