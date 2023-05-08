package gpt

import (
	"context"
	"errors"
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

func NewGPTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "gpt [prompt]",
		Short:            "Use ChatGPT for analysis",
		Long:             "Use ChatGPT for analysis",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("your need to input a prompt")
			}
			query := args[0]
			openaiApiKey, _ := cmd.PersistentFlags().GetString("openai_api_key")
			dir, _ := cmd.PersistentFlags().GetString("dir")
			openaiMode, _ := cmd.PersistentFlags().GetString("openai_mode")
			openaiLimit, _ := cmd.PersistentFlags().GetUint64("openai_limit")
			output, _ := cmd.PersistentFlags().GetString("output")

			//projectWorkspace := "./test_data/test_query_module"
			//downloadWorkspace := "./test_download"

			projectWorkspace := "./"
			downloadWorkspace, _ := config.GetDefaultDownloadCacheDirectory()

			instructions := make(map[string]interface{})
			instructions["query"] = query
			instructions["dir"] = dir
			instructions["openai_api_key"] = openaiApiKey
			instructions["openai_mode"] = openaiMode
			instructions["openai_limit"] = openaiLimit
			instructions["output"] = output

			if instructions["query"] == nil || instructions["query"] == "" {
				return errors.New("query is required")
			}

			return Gpt(cmd.Context(), instructions, projectWorkspace, downloadWorkspace)
		},
	}

	cmd.PersistentFlags().StringP("output", "p", "", "display content format")
	cmd.PersistentFlags().StringP("dir", "d", "", "define the output directory")
	cmd.PersistentFlags().StringP("openai_api_key", "k", "", "your openai_api_key")
	cmd.PersistentFlags().StringP("openai_mode", "m", "", "what mode to use for analysis")
	cmd.PersistentFlags().Uint64P("openai_limit", "i", 10, "how many pieces were analyzed in total")

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

// ------------------------------------------------- --------------------------------------------------------------------

// Apply a project
func Gpt(ctx context.Context, instructions map[string]interface{}, projectWorkspace, downloadWorkspace string) error {

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
		cli_ui.Errorln("Gpt failed")
		return err
	} else {
		cli_ui.Infoln("Selefra Exit")
		return nil
	}
}

// ------------------------------------------------- --------------------------------------------------------------------
