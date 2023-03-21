package init

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/grpc/shard"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra-utils/pkg/pointer"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/cmd/init/rule_example"
	"github.com/selefra/selefra/cmd/version"
	"github.com/selefra/selefra/pkg/cloud_sdk"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/executors"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/modules/parser"
	"github.com/selefra/selefra/pkg/modules/planner"
	"github.com/selefra/selefra/pkg/plugin"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/selefra/selefra/pkg/utils"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// ------------------------------------------------- --------------------------------------------------------------------

// InitCommandExecutorOptions The execution options of the executor that executes the initialization command
type InitCommandExecutorOptions struct {

	// Where to put the downloaded file
	DownloadWorkspace string

	// Which path to initialize as the working directory for your project
	ProjectWorkspace string

	// Whether to force the initialization of the working directory
	IsForceInit bool

	// Which project in the cloud you want to associate with
	RelevanceProject string

	// The database link to use
	DSN string
}

// ------------------------------------------------- --------------------------------------------------------------------

type InitCommandExecutor struct {

	// Used to connect to the selefra cloud
	cloudClient *cloud_sdk.CloudClient

	// Some options when executing the command
	options *InitCommandExecutorOptions
}

func NewInitCommandExecutor(options *InitCommandExecutorOptions) *InitCommandExecutor {
	return &InitCommandExecutor{
		options: options,
	}
}

func (x *InitCommandExecutor) Run(ctx context.Context) error {

	// 1. Check and verify that the working directory can be initialized
	if !x.checkWorkspace() {
		return nil
	}

	// 2. choose provider
	providerSlice, err := x.chooseProvidersList(ctx)
	if err != nil {
		return err
	}
	if len(providerSlice) == 0 {
		cli_ui.Infof("You not select provider\n")
	}

	// init files
	selefraBlock := x.initSelefraYaml(ctx, providerSlice)
	if selefraBlock != nil {
		x.initProvidersYaml(ctx, selefraBlock.RequireProvidersBlock)
	}

	x.initRulesYaml(providerSlice)

	//x.initModulesYaml()

	cli_ui.Infof("Initializing workspace done.\n")

	return nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *InitCommandExecutor) initHeaderOutput(providers []string) {
	//for i := range providers {
	//	cli_ui.Successln(" [✔]" + providers[i] + "\n")
	//}
	cli_ui.Infof(`Running with selefra-cli %s

This command will walk you through creating a new Selefra project

Enter a value or leave blank to accept the (default), and press <ENTER>.
Press ^C at any time to quit.`, version.Version)
	cli_ui.Infof("\n\n")
}

func (x *InitCommandExecutor) chooseProvidersList(ctx context.Context) ([]*registry.Provider, error) {
	providerSlice, err := x.requestProvidersList(ctx)
	if err != nil {
		return nil, err
	}

	if len(providerSlice) == 0 {
		return nil, fmt.Errorf("can not get provider list from registry")
	}

	providerNameSlice := make([]string, 0)
	for _, provider := range providerSlice {
		providerNameSlice = append(providerNameSlice, provider.Name)
	}

	x.initHeaderOutput(providerNameSlice)

	providersSet := cli_ui.SelectProviders(providerNameSlice)
	chooseProviderSlice := make([]*registry.Provider, 0)
	for _, provider := range providerSlice {
		if _, exists := providersSet[provider.Name]; exists {
			chooseProviderSlice = append(chooseProviderSlice, provider)
		}
	}
	return chooseProviderSlice, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *InitCommandExecutor) checkWorkspace() bool {

	// 1. check if workspace dir exist, create it
	_, err := os.Stat(x.options.ProjectWorkspace)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(x.options.ProjectWorkspace, 0755)
		if err != nil {
			cli_ui.Errorf("Create workspace directory: %s failed: %s\n", x.options.ProjectWorkspace, err.Error())
			return false
		}
		cli_ui.Infof("Create workspace directory: %s success\n", x.options.ProjectWorkspace)
	}

	if x.isNeedForceInit() {
		if !x.options.IsForceInit {
			cli_ui.Errorf("Directory %s is not empty, rerun in an empty directory, or use -- force/-f to force overwriting in the current directory\n", x.options.ProjectWorkspace)
			return false
		} else if !x.reInitConfirm() {
			return false
		}
	}

	return true
}

// ------------------------------------------------- --------------------------------------------------------------------

// Determine whether mandatory initialization is required
func (x *InitCommandExecutor) isNeedForceInit() bool {
	dir, _ := os.ReadDir(x.options.ProjectWorkspace)
	files := make([]string, 0)
	for _, v := range dir {
		// Ignore files that are used internally
		if v.Name() == "logs" || v.Name() == "selefra" || v.Name() == "selefra.exe" {
			continue
		}
		files = append(files, v.Name())
	}
	return len(files) != 0
}

// ------------------------------------------------- --------------------------------------------------------------------

const (
	SelefraInputInitForceConfirm     = "SELEFRA_INPUT_INIT_FORCE_CONFIRM"
	SelefraInputInitRelevanceProject = "SELEFRA_INPUT_INIT_RELEVANCE_PROJECT"
)

// reInitConfirm check if current workspace is selefra workspace, then tell user to choose if rewrite selefra workspace
func (x *InitCommandExecutor) reInitConfirm() bool {

	reader := bufio.NewReader(os.Stdin)
	cli_ui.Warningf("Warning: %s is already init. Continue and overwrite it?[Y/N]", x.options.ProjectWorkspace)
	text, err := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	if err != nil && !errors.Is(err, io.EOF) {
		cli_ui.Errorf("Read you input error: %s\n", err.Error())
		return false
	}

	// for test
	if text == "" {
		text = os.Getenv(SelefraInputInitForceConfirm)
	}

	if text != "y" && text != "Y" {
		cli_ui.Errorf("Config file already exists\n")
		return false
	}

	return true
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *InitCommandExecutor) initSelefraYaml(ctx context.Context, providerSlice []*registry.Provider) *module.SelefraBlock {

	selefraBlock := module.NewSelefraBlock()
	projectName, b := x.getProjectName()
	if !b {
		return nil
	}
	selefraBlock.Name = projectName

	// cloud block
	selefraBlock.CloudBlock = x.getCloudBlock(projectName)

	// cli version
	selefraBlock.CliVersion = version.Version
	selefraBlock.LogLevel = "info"

	if len(providerSlice) > 0 {
		requiredProviderSlice := make([]*module.RequireProviderBlock, len(providerSlice))
		for index, provider := range providerSlice {
			requiredProviderBlock := module.NewRequireProviderBlock()
			requiredProviderBlock.Name = provider.Name
			requiredProviderBlock.Source = provider.Name
			requiredProviderBlock.Version = provider.Version
			requiredProviderSlice[index] = requiredProviderBlock
		}
		selefraBlock.RequireProvidersBlock = requiredProviderSlice
	}

	selefraBlock.ConnectionBlock = x.GetConnectionBlock()

	out, err := yaml.Marshal(selefraBlock)
	if err != nil {
		cli_ui.Errorf("Selefra block yaml.Marshal error: %s \n", err.Error())
		return nil
	}
	var selefraNode yaml.Node
	err = yaml.Unmarshal(out, &selefraNode)
	if err != nil {
		cli_ui.Errorf("Selefra yaml.Unmarshal error: %s \n", err.Error())
		return nil
	}
	documentRoot := yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			&yaml.Node{Kind: yaml.ScalarNode, Value: parser.SelefraBlockFieldName},
			&yaml.Node{Kind: yaml.MappingNode, Content: selefraNode.Content[0].Content},
		},
	}
	marshal, err := yaml.Marshal(&documentRoot)
	if err != nil {
		cli_ui.Errorf("Selefra yaml.Marshal error: %s \n", err.Error())
		return nil
	}
	selefraFullPath := filepath.Join(utils.AbsPath(x.options.ProjectWorkspace), "selefra.yaml")
	err = os.WriteFile(selefraFullPath, marshal, 0644)
	if err != nil {
		cli_ui.Errorf("Write %s error: %s \n", selefraFullPath, err.Error())
	} else {
		cli_ui.Successf("Write %s success \n", selefraFullPath)
	}

	return selefraBlock
}

func (x *InitCommandExecutor) getCloudBlock(projectName string) *module.CloudBlock {

	cloudBlock := module.NewCloudBlock()
	cloudBlock.Project = projectName

	if x.cloudClient != nil {
		credentials, diagnostics := x.cloudClient.GetCredentials()
		if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
			return nil
		}
		cloudBlock.Organization = credentials.OrgName
		cloudBlock.HostName = credentials.ServerHost
	}

	return cloudBlock
}

//// init module.yaml
//func (x *InitCommandExecutor) initModulesYaml() {
//	const moduleComment = `
//modules:
//  - name: AWS_Security_Demo
//    uses:
//    - ./rules/
//`
//	moduleFullPath := filepath.Join(utils.AbsPath(x.options.ProjectWorkspace), "module.yaml")
//	err := os.WriteFile(moduleFullPath, []byte(moduleComment), 0644)
//	if err != nil {
//		cli_ui.Errorf("Write %s error: %s\n", moduleFullPath, err.Error())
//	} else {
//		cli_ui.Successf("Write %s success\n", moduleFullPath)
//	}
//}


var rulesMap map[string]string

func init() {
	rulesMap = make(map[string]string)
	rulesMap["aws"] = rule_example.Aws
	rulesMap["azure"] = rule_example.Azure
	rulesMap["gcp"] = rule_example.GCP
	rulesMap["k8s"] = rule_example.K8S
}

func (x *InitCommandExecutor) initRulesYaml(providerSlice []*registry.Provider) {
	for _, provider := range providerSlice {
		ruleYamlString, exists := rulesMap[provider.Name]
		if !exists {
			ruleYamlString = rule_example.DefaultTemplate
		}
		ruleFullPath := filepath.Join(utils.AbsPath(x.options.ProjectWorkspace), fmt.Sprintf("rules_%s.yaml", provider.Name))
		err := os.WriteFile(ruleFullPath, []byte(ruleYamlString), 0644)
		if err != nil {
			cli_ui.Errorf("Write %s error: %s \n", ruleFullPath, err.Error())
		} else {
			cli_ui.Successf("Write %s success \n", ruleFullPath)
		}
	}
}

func (x *InitCommandExecutor) initProvidersYaml(ctx context.Context, requiredProviders module.RequireProvidersBlock) {
	if len(requiredProviders) == 0 {
		cli_ui.Infof("No required provider, do not init providers file \n")
		return
	}
	providers, b := x.makeProviders(ctx, requiredProviders)
	if !b {
		return
	}
	out, err := yaml.Marshal(providers)
	if err != nil {
		cli_ui.Errorf("Providers block yaml.Marshal error: %s \n", err.Error())
		return
	}
	//fmt.Println("Providers Yaml string: " + string(out))

	var providersNode yaml.Node
	err = yaml.Unmarshal(out, &providersNode)
	if err != nil {
		cli_ui.Errorf("Providers yaml.Unmarshal error: %s \n", err.Error())
		return
	}
	//fmt.Println(fmt.Sprintf("length: %d", len(providersNode.Content[0].Content[0].Content)))
	documentRoot := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			&yaml.Node{Kind: yaml.ScalarNode, Value: parser.ProvidersBlockName},
			&yaml.Node{Kind: providersNode.Content[0].Kind, Content: providersNode.Content[0].Content},
		},
	}
	marshal, err := yaml.Marshal(documentRoot)
	if err != nil {
		cli_ui.Errorf("Providers yaml.Marshal error: %s \n", err.Error())
		return
	}
	//fmt.Println("Yaml string: " + string(marshal))
	providerFullName := filepath.Join(utils.AbsPath(x.options.ProjectWorkspace), "providers.yaml")
	err = os.WriteFile(providerFullName, marshal, 0644)
	if err != nil {
		cli_ui.Errorf("Write %s error: %s \n", providerFullName, err.Error())
	} else {
		cli_ui.Successf("Write %s success \n", providerFullName)
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// TODO Automatically installs and starts the database and sets connection items
func (x *InitCommandExecutor) GetConnectionBlock() *module.ConnectionBlock {

	//// 1. Try to get the DSN from the cloud
	//if x.cloudClient != nil {
	//	dsn, diagnostics := x.cloudClient.FetchOrgDSN()
	//	if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
	//		return nil
	//	}
	//	if dsn != "" {
	//		return x.parseDsnAsConnectionBlock(dsn)
	//	}
	//}
	//
	//// 2.

	//cli_runtime.Init(x.options.ProjectWorkspace)
	//
	//dsn, diagnostics := cli_runtime.GetDSN()
	//if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
	//	return nil
	//}
	//if dsn != "" {
	//	return module.ParseConnectionBlockFromDSN(dsn)
	//}

	return nil
}

func (x *InitCommandExecutor) getProjectName() (string, bool) {

	// 1. Use the specified one, if any
	if x.options.RelevanceProject != "" {
		return x.options.RelevanceProject, true
	}

	defaultProjectName := filepath.Base(utils.AbsPath(x.options.ProjectWorkspace))

	// 2. Let the user specify from standard input, the default project name is the name of the current folder
	var err error
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Project name:(%s)", defaultProjectName)
	projectName, err := reader.ReadString('\n')
	if err != nil {
		cli_ui.Errorf("Read you project name error: %s\n", err.Error())
		return "", false
	}
	projectName = strings.TrimSpace(strings.Replace(projectName, "\n", "", -1))
	if projectName == "" {
		projectName = defaultProjectName
	}
	return projectName, true
}

// Pull all providers from the remote repository
func (x *InitCommandExecutor) requestProvidersList(ctx context.Context) ([]*registry.Provider, error) {
	githubRegistry, err := registry.NewProviderGithubRegistry(registry.NewProviderGithubRegistryOptions(x.options.DownloadWorkspace))
	if err != nil {
		return nil, err
	}
	providerSlice, err := githubRegistry.List(ctx)
	if err != nil {
		return nil, err
	}
	return providerSlice, nil
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *InitCommandExecutor) makeProviders(ctx context.Context, requiredProvidersBlock module.RequireProvidersBlock) (module.ProvidersBlock, bool) {

	providersBlock := make(module.ProvidersBlock, 0)
	// convert required provider block to
	for _, requiredProvider := range requiredProvidersBlock {

		cli_ui.Infof("Begin install provider %s \n", requiredProvider.Source)

		providerInstallPlan := &planner.ProviderInstallPlan{
			Provider: registry.NewProvider(requiredProvider.Name, requiredProvider.Version),
		}

		// install providers
		hasError := atomic.Bool{}
		messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
			if err := cli_ui.PrintDiagnostics(message); err != nil {
				hasError.Store(true)
			}
		})
		executor, d := executors.NewProviderInstallExecutor(&executors.ProviderInstallExecutorOptions{
			Plans: []*planner.ProviderInstallPlan{
				providerInstallPlan,
			},
			MessageChannel:    messageChannel,
			DownloadWorkspace: x.options.DownloadWorkspace,
			// TODO
			ProgressTracker: nil,
		})
		if err := cli_ui.PrintDiagnostics(d); err != nil {
			messageChannel.SenderWaitAndClose()
			return nil, false
		}
		d = executor.Execute(ctx)
		messageChannel.ReceiverWait()
		if err := cli_ui.PrintDiagnostics(d); err != nil {
			return nil, false
		}
		if hasError.Load() {
			return nil, false
		}
		cli_ui.Infof("Install provider %s success \n", requiredProvider.Source)

		// init
		cli_ui.Infof("Begin init provider %s... \n", requiredProvider.Source)
		configuration, b := x.getProviderInitConfiguration(ctx, executor.GetLocalProviderManager(), providerInstallPlan)
		if !b {
			return nil, false
		}
		providerBlock := module.NewProviderBlock()
		providerBlock.Provider = requiredProvider.Name
		providerBlock.Name = requiredProvider.Name
		providerBlock.Cache = "1d"
		providerBlock.MaxGoroutines = pointer.ToUInt64Pointer(100)
		providerBlock.ProvidersConfigYamlString = configuration
		providersBlock = append(providersBlock, providerBlock)

		//fmt.Println("Provider Block: " + json_util.ToJsonString(providerBlock))

		cli_ui.Infof("Init provider %s done \n", requiredProvider.Source)
	}
	return providersBlock, true
}

// run provider & get it's init configuration
func (x *InitCommandExecutor) getProviderInitConfiguration(ctx context.Context, localProviderManager *local_providers_manager.LocalProvidersManager, plan *planner.ProviderInstallPlan) (string, bool) {

	// start & get information
	cli_ui.Infof("Begin init provider %s \n", plan.String())

	// Find the local path of the provider
	localProvider := &local_providers_manager.LocalProvider{
		Provider: plan.Provider,
	}
	installed, d := localProviderManager.IsProviderInstalled(ctx, localProvider)
	if err := cli_ui.PrintDiagnostics(d); err != nil {
		return "", false
	}
	if !installed {
		cli_ui.Errorf("Provider %s not installed, can not exec init for it! \n", plan.String())
		return "", false
	}

	// Find the local installation location of the provider
	localProviderMeta, d := localProviderManager.Get(ctx, localProvider)
	if err := cli_ui.PrintDiagnostics(d); err != nil {
		return "", false
	}

	// Start provider
	plug, err := plugin.NewManagedPlugin(localProviderMeta.ExecutableFilePath, plan.Name, plan.Version, "", nil)
	if err != nil {
		cli_ui.Errorf("Start provider %s at %s failed: %s \n", plan.String(), localProvider.ExecutableFilePath, err.Error())
		return "", false
	}
	// Close the provider at the end of the method execution
	defer plug.Close()

	cli_ui.Infof("Start provider %s success \n", plan.String())

	// Database connection option
	storageOpt := postgresql_storage.NewPostgresqlStorageOptions(x.options.DSN)
	providerBlock := module.NewProviderBlock()
	providerBlock.Name = plan.Name
	// Because you do not need to actually interact with the database, it is set to public
	pgstorage.WithSearchPath("public")(storageOpt)
	opt, err := json.Marshal(storageOpt)
	if err != nil {
		cli_ui.Errorf("Json marshal postgresql options error: %s \n", err.Error())
		return "", false
	}

	// Initialize the provider
	pluginProvider := plug.Provider()
	//var providerYamlConfiguration string = module.GetDefaultProviderConfigYamlConfiguration(plan.Name, plan.Version)

	providerInitResponse, err := pluginProvider.Init(ctx, &shard.ProviderInitRequest{
		Workspace: pointer.ToStringPointer(utils.AbsPath(x.options.ProjectWorkspace)),
		Storage: &shard.Storage{
			Type:           0,
			StorageOptions: opt,
		},
		IsInstallInit: pointer.FalsePointer(),
		// Without passing in any configuration, there is no interaction with the database
		ProviderConfig: nil,
	})
	if err != nil {
		cli_ui.Errorf("Start provider failed: %s \n", err.Error())
		return "", false
	}
	if err := cli_ui.PrintDiagnostics(providerInitResponse.Diagnostics); err != nil {
		return "", false
	}
	cli_ui.Infof("Provider %s init success \n", plan.String())

	// Get information about the started provider
	information, err := pluginProvider.GetProviderInformation(ctx, &shard.GetProviderInformationRequest{})
	if err != nil {
		cli_ui.Errorf("Provider %s, get provider information failed: %s \n", plan.String(), err.Error())
		return "", false
	}

	// just for debug
	//fmt.Println("Provider Information Name: " + json_util.ToJsonString(information.Name))
	//fmt.Println("Provider Information Version: " + json_util.ToJsonString(information.Version))
	//fmt.Println("Provider Information DefaultConfiguration: " + json_util.ToJsonString(information.DefaultConfigTemplate))

	return information.DefaultConfigTemplate, true
}

// ------------------------------------------------- --------------------------------------------------------------------
