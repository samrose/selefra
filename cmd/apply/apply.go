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
			output, _ := cmd.PersistentFlags().GetString("output")
			openaiApiKey, _ := cmd.PersistentFlags().GetString("openai_api_key")
			openaiMode, _ := cmd.PersistentFlags().GetString("openai_mode")
			openaiLimit, _ := cmd.PersistentFlags().GetUint64("openai_limit")
			//projectWorkspace := "./test_data/test_query_module"
			//downloadWorkspace := "./test_download"
			instructions := make(map[string]interface{})
			instructions["output"] = output
			instructions["openai_api_key"] = openaiApiKey
			instructions["openai_mode"] = openaiMode
			instructions["openai_limit"] = openaiLimit
			projectWorkspace := "./"
			downloadWorkspace, _ := config.GetDefaultDownloadCacheDirectory()

			return Apply(cmd.Context(), instructions, projectWorkspace, downloadWorkspace)
		},
	}
	cmd.PersistentFlags().StringP("output", "p", "", "display content format")
	cmd.PersistentFlags().StringP("openai_api_key", "k", "", "your openai_api_key")
	cmd.PersistentFlags().StringP("openai_mode", "m", "", "what mode to use for analysis\n")
	cmd.PersistentFlags().Uint64P("openai_limit", "i", 10, "how many pieces were analyzed in total")

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

// ------------------------------------------------- --------------------------------------------------------------------

// Apply a project
func Apply(ctx context.Context, instructions map[string]interface{}, projectWorkspace, downloadWorkspace string) error {

	hasError := atomic.Bool{}
	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			if err := cli_ui.PrintDiagnostics(message); err != nil {
				hasError.Store(true)
			}
		}
	})
	d := executors.NewProjectLocalLifeCycleExecutor(&executors.ProjectLocalLifeCycleExecutorOptions{
		Instruction:          instructions,
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
		QueryWorkerNum: 1,
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
