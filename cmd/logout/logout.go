package logout

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/spf13/cobra"
)

func NewLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout to selefra cloud",
		Long:  "Logout to selefra cloud",
		RunE:  RunFunc,
	}

	return cmd
}

func RunFunc(cmd *cobra.Command, args []string) error {

	diagnostics := schema.NewDiagnostics()

	// Server address
	cloudServerHost := cli_env.GetServerHost()
	logger.InfoF("Use server address: %s", cloudServerHost)

	client, d := cloud_sdk.NewCloudClient(cloudServerHost)
	if diagnostics.AddDiagnostics(d).HasError() {
		return cli_ui.PrintDiagnostics(diagnostics)
	}
	logger.InfoF("Create cloud client success \n")

	// If you are not logged in, you are not allowed to log out
	credentials, _ := client.GetCredentials()
	if credentials == nil {
		cli_ui.Errorln("You are not login, please login first! \n")
		return nil
	}
	logger.InfoF("Get credentials success \n")

	// Destroy the local token
	client.SetToken(credentials.Token)
	if err := cli_ui.PrintDiagnostics(client.Logout()); err != nil {
		return err
	}
	cli_ui.ShowLogout(credentials)
	return nil
}
