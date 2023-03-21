package query

import (
	"context"
	"errors"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra-provider-sdk/storage_factory"
	"github.com/selefra/selefra-utils/pkg/dsn_util"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "query",
		Short:            "Query infrastructure data from pgstorage",
		Long:             "Query infrastructure data from pgstorage",
		PersistentPreRun: global.DefaultWrappedInit(),
		Run: func(cmd *cobra.Command, args []string) {

			ctx := cmd.Context()

			downloadDirectory, err := config.GetDefaultDownloadCacheDirectory()
			if err != nil {
				cli_ui.Errorln(err.Error() + " \n")
				return
			}

			projectWorkspace := "./"

			dsn, err := getDsn(ctx, projectWorkspace, downloadDirectory)
			if err != nil {
				cli_ui.Errorln(err.Error() + "\n")
				return
			}

			// show tips
			c, err := dsn_util.NewConfigByDSN(dsn)
			if err != nil {
				cli_ui.Errorln("Parse dsn %s error: %s", dsn, err.Error())
				return
			}
			cli_ui.Infof("Connection to you database `%s` ... \n", c.ToDSN(true))

			options := postgresql_storage.NewPostgresqlStorageOptions(dsn)
			databaseStorage, diagnostics := storage_factory.NewStorage(cmd.Context(), storage_factory.StorageTypePostgresql, options)
			if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
				return
			}

			defer func() {
				databaseStorage.Close()
			}()

			queryClient, _ := NewQueryClient(ctx, storage_factory.StorageTypePostgresql, databaseStorage)
			queryClient.Run(ctx)

		},
	}
	return cmd
}

func getDsn(ctx context.Context, projectWorkspace, downloadWorkspace string) (string, error) {

	// 1. load from project workspace
	dsn, _ := loadDSNFromProjectWorkspace(ctx, projectWorkspace, downloadWorkspace)
	if dsn != "" {
		cli_ui.Infof("Find database connection in workspace. %s \n", projectWorkspace)
		return dsn, nil
	}

	// 2. load from selefra cloud
	client, diagnostics := cloud_sdk.NewCloudClient(cli_env.GetServerHost())
	if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
		return "", err
	}
	if c, _ := client.GetCredentials(); c != nil {
		c, d := client.Login(c.Token)
		if err := cli_ui.PrintDiagnostics(d); err != nil {
			return "", err
		}
		d = client.SaveCredentials(c)
		if err := cli_ui.PrintDiagnostics(d); err != nil {
			return "", err
		}
		orgDSN, d := client.FetchOrgDSN()
		if err := cli_ui.PrintDiagnostics(d); err != nil {
			return "", err
		}
		if orgDSN != "" {
			cli_ui.Infof("Find database connection in you selefra cloud. \n")
			return orgDSN, nil
		}
	}

	// 3. get dsn from env
	dsn = os.Getenv(env.DatabaseDsn)
	if dsn != "" {
		cli_ui.Infof("Find database connection in your env. \n")
		return dsn, nil
	}

	// 4. start default postgresql instance
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			_ = cli_ui.PrintDiagnostics(message)
		}
	})
	dsn = pgstorage.DefaultPostgreSQL(downloadWorkspace, messageChannel)
	messageChannel.ReceiverWait()
	if dsn != "" {
		return dsn, nil
	}

	return "", errors.New("Can not find database connection")
}

// Look for DSN in the configuration of the project's working directory
func loadDSNFromProjectWorkspace(ctx context.Context, projectWorkspace, downloadWorkspace string) (string, error) {
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		// Any error while loading will not print
		//if utils.IsNotEmpty(message) {
		//	_ = cli_ui.PrintDiagnostics(message)
		//}
	})
	loader, err := module_loader.NewLocalDirectoryModuleLoader(&module_loader.LocalDirectoryModuleLoaderOptions{
		ModuleLoaderOptions: &module_loader.ModuleLoaderOptions{
			Source:            projectWorkspace,
			Version:           "",
			DownloadDirectory: downloadWorkspace,
			ProgressTracker:   nil,
			MessageChannel:    messageChannel,
			DependenciesTree:  []string{projectWorkspace},
		},
	})
	if err != nil {
		return "", err
	}
	rootModule, b := loader.Load(ctx)
	if !b {
		return "", nil
	}
	if rootModule.SelefraBlock != nil && rootModule.SelefraBlock.ConnectionBlock != nil && rootModule.SelefraBlock.ConnectionBlock.BuildDSN() != "" {
		return rootModule.SelefraBlock.ConnectionBlock.BuildDSN(), nil
	}
	return "", nil
}
