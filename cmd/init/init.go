package init

import (
	"context"
	"errors"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/spf13/cobra"
	"os"
	"sync/atomic"
)

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [project name]",
		Short: "Prepare your working directory for other commands",
		Long:  "Prepare your working directory for other commands",
		RunE: func(cmd *cobra.Command, args []string) error {

			relevance, _ := cmd.PersistentFlags().GetString("relevance")
			force, _ := cmd.PersistentFlags().GetBool("force")

			downloadDirectory, err := config.GetDefaultDownloadCacheDirectory()
			if err != nil {
				return err
			}

			projectWorkspace := "./"

			dsn, err := getDsn(cmd.Context(), projectWorkspace, downloadDirectory)
			if err != nil {
				cli_ui.Errorf("Get dsn error: %s \n", err.Error())
				return err
			}

			return NewInitCommandExecutor(&InitCommandExecutorOptions{
				IsForceInit:       force,
				RelevanceProject:  relevance,
				ProjectWorkspace:  projectWorkspace,
				DownloadWorkspace: downloadDirectory,
				DSN:               dsn,
			}).Run(cmd.Context())
		},
	}
	cmd.PersistentFlags().BoolP("force", "f", false, "force overwriting the directory if it is not empty")
	cmd.PersistentFlags().StringP("relevance", "r", "", "associate to selefra cloud project, use only after login")

	cmd.SetHelpFunc(cmd.HelpFunc())
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
	hasError := atomic.Bool{}
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if err := cli_ui.PrintDiagnostics(message); err != nil {
			hasError.Store(true)
		}
	})
	dsn = pgstorage.DefaultPostgreSQL(downloadWorkspace, messageChannel)
	messageChannel.ReceiverWait()
	if dsn != "" {
		cli_ui.Infof("Start default postgresql. \n")
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
		messageChannel.SenderWaitAndClose()
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
