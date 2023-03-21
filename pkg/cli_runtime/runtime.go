package cli_runtime
//
//import (
//	"context"
//	"github.com/selefra/selefra-provider-sdk/provider/schema"
//	"github.com/selefra/selefra/cli_ui"
//	"github.com/selefra/selefra/pkg/cli_env"
//	"github.com/selefra/selefra/pkg/cloud_sdk"
//	"github.com/selefra/selefra/pkg/message"
//	"github.com/selefra/selefra/pkg/modules/module"
//	"github.com/selefra/selefra/pkg/modules/module_loader"
//	"github.com/selefra/selefra/pkg/utils"
//)
//
//// Runtime Command line runtime
//var Runtime *CLIRuntime
//
//type CLIRuntime struct {
//
//	// Which is the working directory
//	Workspace string
//
//	// Which directory to download it to
//	DownloadWorkspace string
//
//	// Errors that may occur during operation
//	Diagnostics *schema.Diagnostics
//
//	// The root module in the working directory
//	RootModule *module.Module
//
//	CloudClient *cloud_sdk.CloudClient
//}
//
//func Init(workspace string) {
//	Runtime = NewCLIRuntime(workspace)
//	Runtime.LoadWorkspaceModule()
//}
//
//func NewCLIRuntime(workspace string) *CLIRuntime {
//	x := &CLIRuntime{
//		Workspace: workspace,
//	}
//	return x
//}
//
//func (x *CLIRuntime) InitCloudClient() {
//	host, diagnostics := FindServerHost()
//	x.Diagnostics.AddDiagnostics(diagnostics)
//	if utils.HasError(diagnostics) {
//		return
//	}
//	client, d := cloud_sdk.NewCloudClient(host)
//	x.Diagnostics.AddDiagnostics(d)
//	if utils.HasError(d) {
//		return
//	}
//	x.CloudClient = client
//
//	// Log in automatically if you have local credentials
//	credentials, _ := client.GetCredentials()
//	if credentials != nil {
//		login, d := client.Login(credentials.Token)
//		if utils.HasError(d) {
//			cli_ui.ShowLoginFailed(credentials.Token)
//			return
//		}
//		cli_ui.ShowLoginSuccess(host, login)
//	}
//
//}
//
//func (x *CLIRuntime) LoadWorkspaceModule() *CLIRuntime {
//
//	if utils.HasError(x.Diagnostics) {
//		return x
//	}
//
//	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
//		// TODO log
//	})
//	loader, err := module_loader.NewLocalDirectoryModuleLoader(&module_loader.LocalDirectoryModuleLoaderOptions{
//		ModuleDirectory: x.Workspace,
//		ModuleLoaderOptions: &module_loader.ModuleLoaderOptions{
//			MessageChannel: messageChannel,
//		},
//	})
//	if err != nil {
//		messageChannel.SenderWaitAndClose()
//		x.Diagnostics.AddErrorMsg("create module load from directory %s error: %s", x.Workspace, err.Error())
//		return x
//	}
//	workspaceModule, _ := loader.Load(context.Background())
//	messageChannel.ReceiverWait()
//	if workspaceModule != nil {
//		x.RootModule = workspaceModule
//	}
//
//	return x
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//func FindServerHost() (string, *schema.Diagnostics) {
//
//	// Try to get it from the configuration file
//	if Runtime != nil &&
//		Runtime.RootModule != nil &&
//		Runtime.RootModule.SelefraBlock != nil &&
//		Runtime.RootModule.SelefraBlock.CloudBlock != nil &&
//		Runtime.RootModule.SelefraBlock.CloudBlock.HostName != "" {
//		return Runtime.RootModule.SelefraBlock.CloudBlock.HostName, nil
//	}
//
//	// Try to get it from an environment variable
//	if cli_env.GetServerHost() != "" {
//		return cli_env.GetServerHost(), nil
//	}
//
//	// You can't get either, so use the default
//	return DefaultCloudHost, nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
//
//func GetDSN() (string, *schema.Diagnostics) {
//
//	// Use the configuration of the current module first if it is configured in the current module
//	if Runtime != nil && Runtime.RootModule != nil && Runtime.RootModule.SelefraBlock != nil && Runtime.RootModule.SelefraBlock.ConnectionBlock != nil {
//		return Runtime.RootModule.SelefraBlock.ConnectionBlock.BuildDSN(), nil
//	}
//
//	// Otherwise, check whether to log in
//	if Runtime.CloudClient != nil && Runtime.CloudClient.IsLoggedIn() {
//		return Runtime.CloudClient.FetchOrgDSN()
//	}
//
//	//// Environment variable
//	//if env.GetDatabaseDsn() != "" {
//	//	return env.GetDatabaseDsn(), nil
//	//}
//
//	// TODO Built-in PG database
//	return "", nil
//}
//
//// ------------------------------------------------- --------------------------------------------------------------------
