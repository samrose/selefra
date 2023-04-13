package executors

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/grpc/pb/log"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/modules/module_loader"
	"github.com/selefra/selefra/pkg/modules/planner"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/selefra/selefra/pkg/utils"
	"os"
)

// ------------------------------------------------- --------------------------------------------------------------------

// ProjectLifeCycleStep You can control the execution of a project until it ends at a certain point in the lifecycle
type ProjectLifeCycleStep int

const (

	// ProjectLifeCycleStepQuery At what point in the project's life cycle, The order is reversed
	// Proceed to the module query step
	ProjectLifeCycleStepQuery ProjectLifeCycleStep = iota

	// ProjectLifeCycleStepFetch Go to the pull data step
	ProjectLifeCycleStepFetch

	// ProjectLifeCycleStepInstall Proceed to the installation step
	ProjectLifeCycleStepInstall

	// ProjectLifeCycleStepModuleCheck Module validity check
	ProjectLifeCycleStepModuleCheck

	// ProjectLifeCycleStepCloudInit Initialization phase
	ProjectLifeCycleStepCloudInit

	// ProjectLifeCycleStepLoadModule Just load the module of the project and do nothing else
	ProjectLifeCycleStepLoadModule
)

// ------------------------------------------------ ---------------------------------------------------------------------

// ProjectLocalLifeCycleExecutorOptions The local life cycle of the project
type ProjectLocalLifeCycleExecutorOptions struct {
	// Gpt query string
	Instruction map[string]interface{}

	// project path
	ProjectWorkspace string

	// download things put where
	DownloadWorkspace string

	// The channel through which messages are received externally
	MessageChannel *message.Channel[*schema.Diagnostics]

	// Used to control where the project goes
	ProjectLifeCycleStep ProjectLifeCycleStep

	// Used to control which step is executed when pulling, the pull has its own separate lifecycle step partition
	FetchStep FetchStep

	// if set this options, then enable cloud project
	ProjectCloudLifeCycleExecutorOptions *ProjectCloudLifeCycleExecutorOptions

	// The database to which data is to be written, May be copied by a higher priority setting
	DSN string

	// The number of concurrences during the fetch phase
	FetchWorkerNum uint64

	// The number of concurrent queries executed
	QueryWorkerNum uint64
}

// ------------------------------------------------ ---------------------------------------------------------------------

// ProjectLifeCycleExecutorName The life cycle of the project
const ProjectLifeCycleExecutorName = "project-local-life-cycle-executor"

// ProjectLocalLifeCycleExecutor Used to fully run the entire project lifecycle
type ProjectLocalLifeCycleExecutor struct {

	// Some options required for the local life cycle
	options *ProjectLocalLifeCycleExecutorOptions

	// project module path
	rootModule *module.Module

	// for sync to cloud, If you log in, it has a real effect. If you do not log in, it has no real effect
	cloudExecutor *ProjectCloudLifeCycleExecutor
}

var _ Executor = &ProjectLocalLifeCycleExecutor{}

// NewProjectLocalLifeCycleExecutor Create a project executor
func NewProjectLocalLifeCycleExecutor(options *ProjectLocalLifeCycleExecutorOptions) *ProjectLocalLifeCycleExecutor {
	return &ProjectLocalLifeCycleExecutor{
		options: options,
	}
}

// Name of project
func (x *ProjectLocalLifeCycleExecutor) Name() string {
	return ProjectLifeCycleExecutorName
}

// Execute Actually execute the project
func (x *ProjectLocalLifeCycleExecutor) Execute(ctx context.Context) *schema.Diagnostics {
	defer func() {

		// close cloud
		if x.cloudExecutor != nil {
			x.cloudExecutor.ShutdownAndWait(ctx)
		}

		// cloud self
		x.options.MessageChannel.SenderWaitAndClose()

	}()

	// load module & check
	if !x.loadModule(ctx) {
		return nil
	}

	// init cloud
	if x.options.ProjectLifeCycleStep > ProjectLifeCycleStepCloudInit {
		return nil
	}
	_ = x.initCloudClient(ctx)
	//if !ok {
	//	_ = x.cloudExecutor.UploadLog(ctx, schema.NewDiagnostics().AddErrorMsg("Selefra Cloud init failed, exit."))
	//	return nil
	//}
	//_ = x.cloudExecutor.UploadLog(ctx, schema.NewDiagnostics().AddInfo("Selefra Cloud init success"))

	// fix dsn
	if !x.fixDsn(ctx) {
		return nil
	}

	// validate module is ok
	if x.options.ProjectLifeCycleStep > ProjectLifeCycleStepModuleCheck {
		return nil
	}
	validatorContext := module.NewValidatorContext()
	d := x.rootModule.Check(x.rootModule, validatorContext)
	if x.cloudExecutor.UploadLog(ctx, d) {
		return nil
	}

	// install provider
	if x.options.ProjectLifeCycleStep > ProjectLifeCycleStepInstall {
		return nil
	}
	providersInstallPlan, providerLocalManager, b := x.install(ctx)
	if !b {
		x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_INITIALIZING, log.Status_STATUS_FAILED)
		return nil
	}

	// check update
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\nChecking Selefra provider updates...\n"))
	for _, plan := range providersInstallPlan {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\t- %s all ready updated!", plan.String()))
	}
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\nSelefra has been finished update providers!"))

	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\nTesting Selefra operation environment...\n"))
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\t- Client verification completed"))
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\t- Providers verification completed"))
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\t- Profile verification completed"))
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\nComplete the Selefra runtime environment test!"))

	x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_INITIALIZING, log.Status_STATUS_SUCCESS)
	x.cloudExecutor.ChangeLogStage(log.StageType_STAGE_TYPE_PULL_INFRASTRUCTURE)

	// fetch data
	if x.options.ProjectLifeCycleStep > ProjectLifeCycleStepFetch {
		return nil
	}
	fetchExecutor, fetchPlans, b := x.fetch(ctx, providersInstallPlan, providerLocalManager)
	if !b {
		x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_PULL_INFRASTRUCTURE, log.Status_STATUS_FAILED)
		return nil
	}
	// A value of 0 indicates that none of the providers has been successfully pulled, so there is no need to start subsequent pull tasks
	if len(fetchExecutor.GetProviderInformationMap()) == 0 {
		x.cloudExecutor.UploadLog(ctx, schema.NewDiagnostics().AddErrorMsg("Fetch failed, can not get provider information"))
		x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_PULL_INFRASTRUCTURE, log.Status_STATUS_FAILED)
		return nil
	}
	x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_PULL_INFRASTRUCTURE, log.Status_STATUS_SUCCESS)
	x.cloudExecutor.ChangeLogStage(log.StageType_STAGE_TYPE_INFRASTRUCTURE_ANALYSIS)

	// exec query
	if x.options.ProjectLifeCycleStep > ProjectLifeCycleStepQuery {
		return nil
	}
	if !x.query(ctx, fetchExecutor, fetchPlans) {
		x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_INFRASTRUCTURE_ANALYSIS, log.Status_STATUS_FAILED)
		return nil
	}
	x.cloudExecutor.ReportTaskStatus(log.StageType_STAGE_TYPE_INFRASTRUCTURE_ANALYSIS, log.Status_STATUS_SUCCESS)

	return nil
}

func (x *ProjectLocalLifeCycleExecutor) fixDsn(ctx context.Context) bool {

	// 1. first take from local module
	if x.rootModule != nil && x.rootModule.SelefraBlock != nil && x.rootModule.SelefraBlock.ConnectionBlock != nil {
		logger.InfoF("fix dsn from selefra block")
		x.options.DSN = x.rootModule.SelefraBlock.ConnectionBlock.BuildDSN()
		return true
	}

	// 2. if is login, take from
	if x.cloudExecutor != nil && x.cloudExecutor.cloudClient != nil && x.cloudExecutor.cloudClient.IsLoggedIn() {
		dsn, diagnostics := x.cloudExecutor.cloudClient.FetchOrgDSN()
		x.options.MessageChannel.Send(diagnostics)
		if utils.HasError(diagnostics) {
			return false
		}
		if dsn != "" {
			logger.InfoF("fix dsn from cloud")
			x.options.DSN = dsn
			return true
		}
	}

	// 3. from options
	if x.options.DSN != "" {
		logger.InfoF("fix dsn from options")
		return true
	}

	// 4. from env
	if os.Getenv(env.DatabaseDsn) != "" {
		logger.InfoF("fix dsn from env")
		x.options.DSN = os.Getenv(env.DatabaseDsn)
		return true
	}

	// 5. start default postgresql instance
	dsn := pgstorage.DefaultPostgreSQL(x.options.DownloadWorkspace, x.options.MessageChannel.MakeChildChannel())
	if dsn != "" {
		x.options.DSN = dsn
		return true
	}

	return false
}

// Load the module to be apply
func (x *ProjectLocalLifeCycleExecutor) loadModule(ctx context.Context) bool {
	moduleLoaderOptions := &module_loader.LocalDirectoryModuleLoaderOptions{
		Instruction: x.options.Instruction,
		ModuleLoaderOptions: &module_loader.ModuleLoaderOptions{
			Source:            x.options.ProjectWorkspace,
			Version:           "",
			DownloadDirectory: x.options.DownloadWorkspace,
			// TODO
			ProgressTracker:  nil,
			MessageChannel:   x.options.MessageChannel.MakeChildChannel(),
			DependenciesTree: []string{x.options.ProjectWorkspace},
		},
		ModuleDirectory: x.options.ProjectWorkspace,
	}

	loader, err := module_loader.NewLocalDirectoryModuleLoader(moduleLoaderOptions)
	if err != nil {
		moduleLoaderOptions.MessageChannel.SenderWaitAndClose()
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("create local directory module loader from %s error: %s", x.options.ProjectWorkspace, err.Error()))
		return false
	}

	rootModule, b := loader.Load(ctx)
	if !b {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("local directory module loader load  %s failed.", x.options.ProjectWorkspace))
		return false
	}

	x.rootModule = rootModule
	return true
}

// ------------------------------------------------- --------------------------------------------------------------------

// install need providers
func (x *ProjectLocalLifeCycleExecutor) install(ctx context.Context) (planner.ProvidersInstallPlan, *local_providers_manager.LocalProvidersManager, bool) {

	// Make an installation plan
	providersInstallPlan, diagnostics := planner.MakeProviderInstallPlan(ctx, x.rootModule)
	if x.cloudExecutor.UploadLog(ctx, diagnostics) {
		return nil, nil, false
	}
	if len(providersInstallPlan) == 0 {
		_ = x.cloudExecutor.UploadLog(ctx, schema.NewDiagnostics().AddErrorMsg("no providers"))
		return nil, nil, false
	}

	// Installation-dependent dependency
	installMessageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			_ = x.cloudExecutor.UploadLog(ctx, message)
		}
	})
	executor, diagnostics := NewProviderInstallExecutor(&ProviderInstallExecutorOptions{
		Plans:             providersInstallPlan,
		MessageChannel:    installMessageChannel,
		DownloadWorkspace: x.options.DownloadWorkspace,
		// TODO
		ProgressTracker: nil,
	})
	if x.cloudExecutor.UploadLog(ctx, diagnostics) {
		installMessageChannel.SenderWaitAndClose()
		return nil, nil, false
	}
	d := executor.Execute(context.Background())
	installMessageChannel.ReceiverWait()
	if x.cloudExecutor.UploadLog(ctx, d) {
		return nil, nil, false
	}
	return providersInstallPlan, executor.GetLocalProviderManager(), true
}

// ------------------------------------------------- --------------------------------------------------------------------

// Start pulling data
func (x *ProjectLocalLifeCycleExecutor) fetch(ctx context.Context, providersInstallPlan planner.ProvidersInstallPlan, localProviderManager *local_providers_manager.LocalProvidersManager) (*ProviderFetchExecutor, planner.ProvidersFetchPlan, bool) {
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\nSelefra will start infrastructure data collection...\n"))
	// Develop a data pull plan
	providerFetchPlans, d := planner.NewProviderFetchPlanner(&planner.ProviderFetchPlannerOptions{
		DSN:                          x.options.DSN,
		Module:                       x.rootModule,
		ProviderVersionVoteWinnerMap: providersInstallPlan.ToMap(),
		MessageChannel:               x.options.MessageChannel.MakeChildChannel(),
	}).MakePlan(ctx)
	if x.cloudExecutor.UploadLog(ctx, d) {
		return nil, nil, false
	}

	fetchMessageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			_ = x.cloudExecutor.UploadLog(ctx, message)
		}
	})
	fetchExecutor := NewProviderFetchExecutor(&ProviderFetchExecutorOptions{
		LocalProviderManager: localProviderManager,
		Plans:                providerFetchPlans,
		MessageChannel:       fetchMessageChannel,
		WorkerNum:            x.options.FetchWorkerNum,
		Workspace:            x.options.ProjectWorkspace,
		DSN:                  x.options.DSN,
		FetchStepTo:          x.options.FetchStep,
	})
	d = fetchExecutor.Execute(context.Background())
	fetchMessageChannel.ReceiverWait()
	if x.cloudExecutor.UploadLog(ctx, d) {
		return nil, nil, false
	}
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Complete Selefra infrastructure data collection!\n"))
	return fetchExecutor, providerFetchPlans, true
}

// ------------------------------------------------- --------------------------------------------------------------------

// Start querying the policy and output the query results to the console and upload them to the cloud
func (x *ProjectLocalLifeCycleExecutor) query(ctx context.Context, fetchExecutor *ProviderFetchExecutor, providerFetchPlans planner.ProvidersFetchPlan) bool {
	plan, d := planner.MakeModuleQueryPlan(ctx, &planner.ModulePlannerOptions{
		Instruction:        x.options.Instruction,
		Module:             x.rootModule,
		TableToProviderMap: fetchExecutor.GetTableToProviderMap(),
	})
	if x.cloudExecutor.UploadLog(ctx, d) {
		return false
	}
	queryMessageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		_ = x.cloudExecutor.UploadLog(ctx, message)
	})
	resultQueryResultChannel := message.NewChannel[*RuleQueryResult](func(index int, message *RuleQueryResult) {
		x.cloudExecutor.UploadIssue(ctx, message)
	})
	contextMap, d := providerFetchPlans.BuildProviderContextMap(ctx, x.options.DSN)
	if x.cloudExecutor.UploadLog(ctx, d) {
		return false
	}
	queryExecutor := NewModuleQueryExecutor(&ModuleQueryExecutorOptions{
		Plan:                   plan,
		DownloadWorkspace:      x.options.DownloadWorkspace,
		MessageChannel:         queryMessageChannel,
		RuleQueryResultChannel: resultQueryResultChannel,
		ProviderInformationMap: fetchExecutor.GetProviderInformationMap(),
		ProviderExpandMap:      contextMap,
		WorkerNum:              x.options.QueryWorkerNum,
		// TODO
		ProgressTracker: nil,
	})
	d = queryExecutor.Execute(ctx)
	resultQueryResultChannel.ReceiverWait()
	queryMessageChannel.ReceiverWait()
	return !x.cloudExecutor.UploadLog(ctx, d)
}

func (x *ProjectLocalLifeCycleExecutor) initCloudClient(ctx context.Context) bool {

	// Projects on the cloud share the same module as local projects
	if x.options.ProjectCloudLifeCycleExecutorOptions == nil {
		x.options.ProjectCloudLifeCycleExecutorOptions = &ProjectCloudLifeCycleExecutorOptions{
			IsNeedLogin:       false,
			EnableConsoleTips: true,
		}
	}

	if x.options.ProjectCloudLifeCycleExecutorOptions.Module == nil {
		x.options.ProjectCloudLifeCycleExecutorOptions.Module = x.rootModule
	}

	// The message queue is connected
	if x.options.ProjectCloudLifeCycleExecutorOptions.MessageChannel == nil {
		x.options.ProjectCloudLifeCycleExecutorOptions.MessageChannel = x.options.MessageChannel.MakeChildChannel()
	}

	// if module set cloud host, use it first
	if x.rootModule != nil && x.rootModule.SelefraBlock != nil && x.rootModule.SelefraBlock.CloudBlock != nil && x.rootModule.SelefraBlock.CloudBlock.HostName != "" {
		x.options.ProjectCloudLifeCycleExecutorOptions.CloudServerHost = x.rootModule.SelefraBlock.CloudBlock.HostName
	}

	x.cloudExecutor = NewProjectCloudLifeCycleExecutor(x.options.ProjectCloudLifeCycleExecutorOptions)
	return x.cloudExecutor.InitCloudClient(ctx)
}

// ------------------------------------------------ ---------------------------------------------------------------------
