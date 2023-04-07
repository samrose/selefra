package cmd

import (
	"fmt"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/cmd/apply"
	"github.com/selefra/selefra/cmd/fetch"
	"github.com/selefra/selefra/cmd/gpt"
	initCmd "github.com/selefra/selefra/cmd/init"
	"github.com/selefra/selefra/cmd/login"
	"github.com/selefra/selefra/cmd/logout"
	"github.com/selefra/selefra/cmd/provider"
	"github.com/selefra/selefra/cmd/query"
	"github.com/selefra/selefra/cmd/test"
	"github.com/selefra/selefra/cmd/version"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/telemetry"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var group = make(map[string][]*cobra.Command)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "selefra",
	Short: "Selefra - Infrastructure data as Code",
	Long: `
Selefra - Infrastructure data as Code

For details see the selefra document https://selefra.io/docs
If you like selefra, give us a star https://github.com/selefra/selefra
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		level, _ := cmd.Flags().GetString("loglevel")
		global.SetLogLevel(level)

		// get telemetry from command params
		telemetryEnable, err := cmd.Flags().GetBool("telemetry")
		if err == nil {
			// user give telemetry param
			telemetry.TelemetryEnable = telemetryEnable
		} else {
			// try find it in env variables
			telemetryEnableString := cli_env.GetSelefraTelemetryEnable()
			if telemetryEnableString != "" {
				if telemetryEnableString == "true" || telemetryEnableString == "t" {
					telemetry.TelemetryEnable = true
				} else if telemetryEnableString == "false" || telemetryEnableString == "f" {
					telemetry.TelemetryEnable = false
				}
			} else {
				// keep default value
			}
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// need close telemetry on exit
		diagnostics := telemetry.Close(cmd.Context())
		return cli_ui.PrintDiagnostics(diagnostics)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		utils.Close()
	}()

	if err := rootCmd.Execute(); err != nil {
		log.Printf("Error occurred in Execute: %+v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "log level")
	rootCmd.PersistentFlags().BoolP("telemetry ", "t", true, "Whether to enable telemetry. This parameter is enabled by default")
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
	group["main"] = []*cobra.Command{
		initCmd.NewInitCmd(),
		test.NewTestCmd(),
		apply.NewApplyCmd(),
		login.NewLoginCmd(),
		logout.NewLogoutCmd(),
		gpt.NewGPTCmd(),
	}

	group["other"] = []*cobra.Command{
		fetch.NewFetchCmd(),
		provider.NewProviderCmd(),
		query.NewQueryCmd(),
		version.NewVersionCmd(),
	}

	rootCmd.AddCommand(group["main"]...)
	rootCmd.AddCommand(group["other"]...)

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(strings.TrimSpace(cmd.Long))

		fmt.Println("\nUsage:")
		fmt.Printf("  %-13s", "selefra [command]\n\n")

		fmt.Println("Main commands:")
		for _, c := range group["main"] {
			fmt.Printf("  %-13s%s\n", c.Name(), c.Short)
		}
		fmt.Println()
		fmt.Println("All other commands:")
		for _, c := range group["other"] {
			fmt.Printf("  %-13s%s\n", c.Name(), c.Short)
		}
		fmt.Println()

		fmt.Println("Flags")
		fmt.Println(cmd.Flags().FlagUsages())

		fmt.Println(`Use "selefra [command] --help" for more information about a command.`)
	})

}
