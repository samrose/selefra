package fetch

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/executors"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/spf13/cobra"
	"sync/atomic"
)

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "fetch",
		Short:            "Fetch resources from configured providers",
		Long:             "Fetch resources from configured providers",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, args []string) error {

			projectWorkspace := "./"
			downloadWorkspace, _ := config.GetDefaultDownloadCacheDirectory()

			//cli_runtime.Init(projectWorkspace)

			Fetch(projectWorkspace, downloadWorkspace)

			return nil
		},
	}

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

func Fetch(projectWorkspace, downloadWorkspace string) *schema.Diagnostics {

	hasError := atomic.Bool{}
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			if err := cli_ui.PrintDiagnostics(message); err != nil {
				hasError.Store(true)
			}
		}
	})
	d := executors.NewProjectLocalLifeCycleExecutor(&executors.ProjectLocalLifeCycleExecutorOptions{
		ProjectWorkspace:                     projectWorkspace,
		DownloadWorkspace:                    downloadWorkspace,
		MessageChannel:                       messageChannel,
		ProjectLifeCycleStep:                 executors.ProjectLifeCycleStepFetch,
		FetchStep:                            executors.FetchStepFetch,
		ProjectCloudLifeCycleExecutorOptions: nil,
		//DSN:                                  env.GetDatabaseDsn(),
		FetchWorkerNum: 1,
		QueryWorkerNum: 20,
	}).Execute(context.Background())
	messageChannel.ReceiverWait()
	_ = cli_ui.PrintDiagnostics(d)
	if utils.HasError(d) || hasError.Load() {
		cli_ui.Errorln("fetch failed!")
	} else {
		cli_ui.Infoln("fetch done!")
	}

	return nil
}
