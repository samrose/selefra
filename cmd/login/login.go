package login

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/spf13/cobra"
)

func NewLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "login [token]",
		Short:            "Login to selefra cloud using token",
		Long:             "Login to selefra cloud using token",
		PersistentPreRun: global.DefaultWrappedInit(),
		RunE:             RunFunc,
	}

	cmd.SetHelpFunc(cmd.HelpFunc())
	return cmd
}

func RunFunc(cmd *cobra.Command, args []string) error {

	diagnostics := schema.NewDiagnostics()

	cloudServerHost := cli_env.GetServerHost()
	logger.InfoF("Use server address: %s", cloudServerHost)

	client, d := cloud_sdk.NewCloudClient(cloudServerHost)
	if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
		return err
	}
	logger.InfoF("Create cloud client success \n")

	var token string
	if len(args) != 0 {
		token = args[0]
		cli_ui.Warningf("Security warning: Entering a token directly on the command line will be recorded in the command line history and may cause your token to leak! \n")
	}

	// If you are already logged in, repeat login is not allowed and you must log out first
	getCredentials, _ := client.GetCredentials()
	if getCredentials != nil {
		cli_ui.Errorf("You already logged in as %s, please logout first. \n", getCredentials.UserName)
		return nil
	}

	// Read the token from standard input
	if token == "" {
		token, d = cli_ui.InputCloudToken(cloudServerHost)
		if err := cli_ui.PrintDiagnostics(d); err != nil {
			return err
		}
	}
	if token == "" {
		cli_ui.Errorf("Token can not be empty! \n")
		return nil
	}

	credentials, d := client.Login(token)
	if err := cli_ui.PrintDiagnostics(d); err != nil {
		cli_ui.ShowLoginFailed(token)
		return nil
	}

	cli_ui.ShowLoginSuccess(cloudServerHost, credentials)

	return nil
}
