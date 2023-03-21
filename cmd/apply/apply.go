package apply

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

func NewApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "apply",
		Short:            "Analyze infrastructure",
		Long:             "Analyze infrastructure",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, args []string) error {

			//projectWorkspace := "./test_data/test_query_module"
			//downloadWorkspace := "./test_download"

			projectWorkspace := "./"
			downloadWorkspace, _ := config.GetDefaultDownloadCacheDirectory()

			return Apply(cmd.Context(), projectWorkspace, downloadWorkspace)
		},
	}

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

// ------------------------------------------------- --------------------------------------------------------------------

// Apply a project
func Apply(ctx context.Context, projectWorkspace, downloadWorkspace string) error {

	hasError := atomic.Bool{}
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			if err := cli_ui.PrintDiagnostics(message); err != nil {
				hasError.Store(true)
			}
		}
	})
	d := executors.NewProjectLocalLifeCycleExecutor(&executors.ProjectLocalLifeCycleExecutorOptions{
		ProjectWorkspace:     projectWorkspace,
		DownloadWorkspace:    downloadWorkspace,
		MessageChannel:       messageChannel,
		ProjectLifeCycleStep: executors.ProjectLifeCycleStepQuery,
		FetchStep:            executors.FetchStepFetch,
		ProjectCloudLifeCycleExecutorOptions: &executors.ProjectCloudLifeCycleExecutorOptions{
			EnableConsoleTips: true,
			IsNeedLogin:       true,
		},
		//DSN:                                  env.GetDatabaseDsn(),
		FetchWorkerNum: 1,
		QueryWorkerNum: 10,
	}).Execute(ctx)
	messageChannel.ReceiverWait()
	if err := cli_ui.PrintDiagnostics(d); err != nil {
		cli_ui.Errorln("Apply failed")
		return err
		//} else if hasError.Load() {
		//	cli_ui.Errorln("Apply failed")
		//	return fmt.Errorf("Apply Failed")
	} else {
		cli_ui.Infoln("Apply done")
		return nil
	}
}

// ------------------------------------------------- --------------------------------------------------------------------
