package test

import (
	"context"
	"errors"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/executors"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"sync/atomic"
)

// TODO 2023-2-20 15:32:56 Returns a non-zero value if the test fails
func NewTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "test",
		Short:            "Check whether the configuration is valid",
		Long:             "Check whether the configuration is valid",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, args []string) error {

			//projectWorkspace := "./test_data/test_query_module"
			//downloadWorkspace := "./test_download"

			projectWorkspace := "./"
			downloadWorkspace, _ := config.GetDefaultDownloadCacheDirectory()

			return Test(cmd.Context(), projectWorkspace, downloadWorkspace)
		},
	}

	cmd.SetHelpFunc(cmd.HelpFunc())

	return cmd
}

func Test(ctx context.Context, projectWorkspace, downloadWorkspace string) error {
	cli_ui.Infof("\nTesting Selefra operation environment...\n")
	hasError := atomic.Bool{}
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			_ = cli_ui.PrintDiagnostics(message)
		}
		if utils.HasError(message) {
			hasError.Store(true)
		}
	})
	dsn, err := getDsn(ctx, projectWorkspace, downloadWorkspace)
	if err != nil {
		return err
	}
	d := executors.NewProjectLocalLifeCycleExecutor(&executors.ProjectLocalLifeCycleExecutorOptions{
		ProjectWorkspace:                     projectWorkspace,
		DownloadWorkspace:                    downloadWorkspace,
		MessageChannel:                       messageChannel,
		ProjectLifeCycleStep:                 executors.ProjectLifeCycleStepFetch,
		FetchStep:                            executors.FetchStepGetInformation,
		ProjectCloudLifeCycleExecutorOptions: nil,
		DSN:                                  dsn,
		FetchWorkerNum:                       1,
		QueryWorkerNum:                       1,
	}).Execute(context.Background())
	messageChannel.ReceiverWait()

	cli_ui.Infoln("\t- Client verification completed")
	cli_ui.Infoln("\t- Providers verification completed")
	cli_ui.Infoln("\t- Profile verification completed")
	cli_ui.Infoln("\nComplete the Selefra runtime environment test!")

	if utils.IsNotEmpty(d) {
		_ = cli_ui.PrintDiagnostics(d)
		cli_ui.Errorln("Apply failed")
	} else {
		cli_ui.Infoln("Apply done")
	}

	if hasError.Load() {
		return errors.New("Need help? Known on Slack or open a Github Issue: https://github.com/selefra/selefra#community")
	}
	return nil
}

func getDsn(ctx context.Context, projectWorkspace, downloadWorkspace string) (string, error) {

	// 1. load from project workspace
	dsn, _ := loadDSNFromProjectWorkspace(ctx, projectWorkspace, downloadWorkspace)
	if dsn != "" {
		//cli_ui.Infof("Find database connection in workspace. %s \n", projectWorkspace)
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
			//cli_ui.Infof("Find database connection in you selefra cloud. \n")
			return orgDSN, nil
		}
	}

	// 3. get dsn from env
	dsn = os.Getenv(env.DatabaseDsn)
	if dsn != "" {
		//cli_ui.Infof("Find database connection in your env. \n")
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
