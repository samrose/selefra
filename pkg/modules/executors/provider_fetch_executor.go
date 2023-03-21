package executors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/grpc/shard"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra-provider-sdk/storage_factory"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra-utils/pkg/pointer"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/modules/planner"
	"github.com/selefra/selefra/pkg/plugin"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/selefra/selefra/pkg/utils"
	"io"
	"path/filepath"
	"sync"
	"time"
)

// ------------------------------------------------ ---------------------------------------------------------------------

// FetchStep The pull is broken down into small steps, and you can control where to stop
type FetchStep int

const (

	// FetchStepFetch Notice that the order is reversed
	// The default level is the data after fetch
	FetchStepFetch FetchStep = iota

	// FetchStepCreateAllTable Go to create all tables
	FetchStepCreateAllTable FetchStep = iota

	// FetchStepDropAllTable Go to delete all tables
	FetchStepDropAllTable FetchStep = iota

	// FetchStepGetInformation Go to get the Provider information
	FetchStepGetInformation FetchStep = iota

	// FetchStepGetInit Perform Provider initialization
	FetchStepGetInit FetchStep = iota

	// FetchStepGetStart Just start the Provider up and quit
	FetchStepGetStart FetchStep = iota
)

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderFetchExecutorOptions Various parameter options when pulling data
type ProviderFetchExecutorOptions struct {

	// Used to find the Provider and start the instance
	LocalProviderManager *local_providers_manager.LocalProvidersManager

	// The pull plan to execute
	Plans []*planner.ProviderFetchPlan

	// Receive message feedback in real time
	MessageChannel *message.Channel[*schema.Diagnostics]

	// Number of providers that are concurrently pulled
	WorkerNum uint64

	// Working directory
	Workspace string

	// Connect to database
	DSN string

	// At which stage to exit
	FetchStepTo FetchStep
}

// ------------------------------------------------- --------------------------------------------------------------------

const FetchExecutorName = "provider-fetch-executor"

// ProviderFetchExecutor An actuator for pulling data
type ProviderFetchExecutor struct {
	options *ProviderFetchExecutorOptions

	// After the Provider is started, information about the Provider is collected
	providerInformationMap map[string]*shard.GetProviderInformationResponse
}

var _ Executor = &ProviderFetchExecutor{}

func NewProviderFetchExecutor(options *ProviderFetchExecutorOptions) *ProviderFetchExecutor {
	return &ProviderFetchExecutor{
		options: options,
	}
}

func (x *ProviderFetchExecutor) GetProviderInformationMap() map[string]*shard.GetProviderInformationResponse {
	return x.providerInformationMap
}

func (x *ProviderFetchExecutor) GetTableToProviderMap() map[string]string {
	tableToProviderMap := make(map[string]string)
	for providerName, providerInformation := range x.providerInformationMap {
		for _, table := range providerInformation.Tables {
			flatTableToProviderMap(providerName, table, tableToProviderMap)
		}
	}
	return tableToProviderMap
}

// Generate a mapping of a single table to the provider
func flatTableToProviderMap(providerName string, table *schema.Table, m map[string]string) {
	m[table.TableName] = providerName

	for _, subTable := range table.SubTables {
		flatTableToProviderMap(providerName, subTable, m)
	}
}

func (x *ProviderFetchExecutor) Name() string {
	return FetchExecutorName
}

func (x *ProviderFetchExecutor) Execute(ctx context.Context) *schema.Diagnostics {

	defer func() {
		logger.InfoF("fetch MessageChannel.SenderWaitAndClose begin")
		x.options.MessageChannel.SenderWaitAndClose()
		logger.InfoF("fetch MessageChannel.SenderWaitAndClose end")
	}()

	// TODO Scheduling algorithm, Minimize waiting
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Make fetch queue begin..."))
	fetchPlanChannel := make(chan *planner.ProviderFetchPlan, len(x.options.Plans))
	for _, plan := range x.options.Plans {
		fetchPlanChannel <- plan
	}
	close(fetchPlanChannel)
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Make fetch queue done..."))

	// The concurrent pull starts
	providerInformationChannel := make(chan *shard.GetProviderInformationResponse, len(x.options.Plans))
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Run fetch worker, worker num %d...", x.options.WorkerNum))
	wg := sync.WaitGroup{}
	for i := uint64(0); i < x.options.WorkerNum; i++ {
		wg.Add(1)
		NewProviderFetchExecutorWorker(x, fetchPlanChannel, providerInformationChannel, &wg).Run()
	}
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Start fetch worker done, wait queue consumer done."))
	wg.Wait()
	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Fetch queue done"))

	// Sort the provider information
	close(providerInformationChannel)
	providerInformationMap := make(map[string]*shard.GetProviderInformationResponse)
	for response := range providerInformationChannel {
		providerInformationMap[response.Name] = response
	}
	x.providerInformationMap = providerInformationMap

	return nil
}

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderFetchExecutorWorker A working coroutine used to perform a pull task
type ProviderFetchExecutorWorker struct {

	// Is the task in which actuator is executed
	executor *ProviderFetchExecutor

	// Task queue
	planChannel chan *planner.ProviderFetchPlan

	// Exit signal
	wg *sync.WaitGroup

	// Collect information about the started providers
	providerInformationCollector chan *shard.GetProviderInformationResponse
}

func NewProviderFetchExecutorWorker(executor *ProviderFetchExecutor, planChannel chan *planner.ProviderFetchPlan, providerInformationCollector chan *shard.GetProviderInformationResponse, wg *sync.WaitGroup) *ProviderFetchExecutorWorker {
	return &ProviderFetchExecutorWorker{
		executor:                     executor,
		planChannel:                  planChannel,
		wg:                           wg,
		providerInformationCollector: providerInformationCollector,
	}
}

func (x *ProviderFetchExecutorWorker) Run() {
	go func() {
		defer func() {
			x.wg.Done()
		}()
		for plan := range x.planChannel {
			// The drop-down time limit for a single Provider is a month. If it is insufficient, adjust it again
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Hour*24*30)
			x.executePlan(ctx, plan)
			cancelFunc()
		}
	}()
}

// Execute a provider fetch task plan
func (x *ProviderFetchExecutorWorker) executePlan(ctx context.Context, plan *planner.ProviderFetchPlan) {

	diagnostics := schema.NewDiagnostics()

	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Begin fetch provider %s", plan.String())))

	// Find the local path of the provider
	localProvider := &local_providers_manager.LocalProvider{
		Provider: plan.Provider,
	}
	installed, d := x.executor.options.LocalProviderManager.IsProviderInstalled(ctx, localProvider)
	if diagnostics.AddDiagnostics(d).HasError() {
		x.sendMessage(x.addProviderNameForMessage(plan, diagnostics))
		return
	}
	if !installed {
		x.sendMessage(x.addProviderNameForMessage(plan, diagnostics.AddErrorMsg("Provider %s not installed, can not exec fetch for it", plan.String())))
		return
	}

	// Find the local installation location of the provider
	localProviderMeta, d := x.executor.options.LocalProviderManager.Get(ctx, localProvider)
	if diagnostics.AddDiagnostics(d).HasError() {
		x.sendMessage(x.addProviderNameForMessage(plan, diagnostics))
		return
	}

	// Start provider
	plug, err := plugin.NewManagedPlugin(localProviderMeta.ExecutableFilePath, plan.Name, plan.Version, "", nil)
	if err != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Start provider %s at %s failed: %s", plan.String(), localProviderMeta.ExecutableFilePath, err.Error())))
		return
	}
	// Close the provider at the end of the method execution
	defer func() {
		plug.Close()
		x.sendMessage(schema.NewDiagnostics().AddInfo("Stop provider %s at %s ", plan.String(), localProviderMeta.ExecutableFilePath))
	}()

	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Start provider %s success", plan.String())))

	// init
	if x.executor.options.FetchStepTo > FetchStepGetInit {
		// TODO log
		return
	}

	// Database connection option
	storageOpt := postgresql_storage.NewPostgresqlStorageOptions(x.executor.options.DSN)
	pgstorage.WithSearchPath(plan.FetchToDatabaseSchema)(storageOpt)
	opt, err := json.Marshal(storageOpt)
	if err != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Json marshal postgresql options error: %s", err.Error())))
		return
	}

	// Get the lock first
	databaseStorage, d := storage_factory.NewStorage(ctx, storage_factory.StorageTypePostgresql, storageOpt)
	x.sendMessage(x.addProviderNameForMessage(plan, d))
	if utils.HasError(d) {
		return
	}
	defer func() {
		databaseStorage.Close()
	}()
	ownerId := utils.BuildLockOwnerId()
	tryTimes := 0
	for {

		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner %s, fetch data, try get fetch lock...", plan.String(), plan.FetchToDatabaseSchema, ownerId)))

		tryTimes++
		err := databaseStorage.Lock(ctx, pgstorage.LockId, ownerId)
		if err != nil {
			x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, owner %s, fetch data, get fetch lock error: %s, will sleep & retry, tryTimes = %d", plan.String(), plan.FetchToDatabaseSchema, ownerId, err.Error(), tryTimes)))
		} else {
			x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner %s, fetch data, get fetch lock success", plan.String(), plan.FetchToDatabaseSchema, ownerId)))
			break
		}
		time.Sleep(time.Second * 10)
	}
	defer func() {
		for tryTimes := 0; tryTimes < 10; tryTimes++ {
			err := databaseStorage.UnLock(ctx, pgstorage.LockId, ownerId)
			if err != nil {
				if errors.Is(err, postgresql_storage.ErrLockNotFound) {
					x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner %s, fetch data, release fetch lock success", plan.String(), plan.FetchToDatabaseSchema, ownerId)))
				} else {
					x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, owner %s, fetch data, release fetch lock error: %s, will sleep & retry, tryTimes = %d", plan.String(), plan.FetchToDatabaseSchema, ownerId, err.Error(), tryTimes)))
				}
			} else {
				x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner %s, fetch data, release fetch lock success", plan.String(), plan.FetchToDatabaseSchema, ownerId)))
				break
			}
		}
	}()

	// TODO Default values for processing parameters
	// Initialize the provider
	pluginProvider := plug.Provider()
	var providerYamlConfiguration string
	if plan.ProviderConfigurationBlock == nil {
		providerYamlConfiguration = module.GetDefaultProviderConfigYamlConfiguration(plan.Name, plan.Version)
	} else {
		providerYamlConfiguration = plan.GetProvidersConfigYamlString()
	}

	workspace, _ := filepath.Abs(x.executor.options.Workspace)
	providerInitResponse, err := pluginProvider.Init(ctx, &shard.ProviderInitRequest{
		Workspace: pointer.ToStringPointer(workspace),
		Storage: &shard.Storage{
			Type:           0,
			StorageOptions: opt,
		},
		IsInstallInit:  pointer.FalsePointer(),
		ProviderConfig: pointer.ToStringPointerOrNilIfEmpty(providerYamlConfiguration),
	})
	if err != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Start provider failed: %s", err.Error())))
		return
	}
	// TODO There is a problem with process interruption here
	if utils.IsNotEmpty(providerInitResponse.Diagnostics) {
		x.sendMessage(x.addProviderNameForMessage(plan, providerInitResponse.Diagnostics))
		if utils.HasError(providerInitResponse.Diagnostics) {
			return
		}
	}
	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s init success", plan.String())))

	// get information
	if x.executor.options.FetchStepTo > FetchStepGetInformation {
		return
	}

	// Get information about the started provider
	information, err := pluginProvider.GetProviderInformation(ctx, &shard.GetProviderInformationRequest{})
	if err != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, get provider information failed: %s", plan.String(), plan.FetchToDatabaseSchema, err.Error())))
		return
	}
	x.providerInformationCollector <- information
	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Get provider %s information success", plan.String())))

	if x.executor.options.FetchStepTo > FetchStepDropAllTable {
		return
	}

	// Check whether the cache can be removed
	cache, needFetchTableSet := x.tryHitCache(ctx, databaseStorage, plan, information)
	if cache {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s pull data hit cache", plan.String())))
		return
	}

	// Delete the table before provider
	dropRes, err := pluginProvider.DropTableAll(ctx, &shard.ProviderDropTableAllRequest{})
	if err != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, drop all table failed: %s", plan.String(), plan.FetchToDatabaseSchema, err.Error())))
		return
	}
	x.sendMessage(x.addProviderNameForMessage(plan, dropRes.Diagnostics))
	if utils.HasError(dropRes.Diagnostics) {
		return
	}
	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s drop database schema clean success", plan.String())))

	if x.executor.options.FetchStepTo > FetchStepCreateAllTable {
		return
	}

	// create all tables
	createRes, err := pluginProvider.CreateAllTables(ctx, &shard.ProviderCreateAllTablesRequest{})
	if err != nil {
		cli_ui.Errorln(err.Error())
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, create all table failed: %s", plan.String(), plan.FetchToDatabaseSchema, err.Error())))
		return
	}
	if createRes.Diagnostics != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, createRes.Diagnostics))
		if utils.HasError(createRes.Diagnostics) {
			return
		}
	}
	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s create tables success", plan.String())))

	if x.executor.options.FetchStepTo > FetchStepFetch {
		return
	}
	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s begin fetch...", plan.String())))

	// being pull data
	needFetchTableNameSlice := make([]string, 0)
	for tableName := range needFetchTableSet {
		needFetchTableNameSlice = append(needFetchTableNameSlice, tableName)
	}
	recv, err := pluginProvider.PullTables(ctx, &shard.PullTablesRequest{
		Tables:        needFetchTableNameSlice,
		MaxGoroutines: plan.GetMaxGoroutines(),
		Timeout:       0,
	})
	if err != nil {
		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, pull table failed: %s", plan.String(), plan.FetchToDatabaseSchema, err.Error())))
		return
	}
	//progbar := progress.DefaultProgress()
	//progbar.Add(decl.Name+"@"+decl.Version, -1)
	//success := 0
	//errorsN := 0
	//var total int64
	//for {
	//	res, err := recv.Recv()
	//	if err != nil {
	//		if errors.Is(err, io.EOF) {
	//			progbar.Current(decl.Name+"@"+decl.Version, total, "Done")
	//			progbar.Done(decl.Name + "@" + decl.Version)
	//			break
	//		}
	//		return err
	//	}
	//	progbar.SetTotal(decl.Name+"@"+decl.Version, int64(res.TableCount))
	//	progbar.Current(decl.Name+"@"+decl.Version, int64(len(res.FinishedTables)), res.Table)
	//	total = int64(res.TableCount)
	//	if res.Diagnostics != nil {
	//		if res.Diagnostics.HasError() {
	//			cli_ui.SaveLogToDiagnostic(res.Diagnostics.GetDiagnosticSlice())
	//		}
	//	}
	//	success = len(res.FinishedTables)
	//	errorsN = 0
	//}
	//progbar.ReceiverWait(decl.Name + "@" + decl.Version)
	//if errorsN > 0 {
	//	cli_ui.Errorf("\nPull complete! Total Resources pulled:%d        Errors: %d\n", success, errorsN)
	//	return nil
	//}
	//cli_ui.Infof("\nPull complete! Total Resources pulled:%d        Errors: %d\n", success, errorsN)
	//return nil

	success := 0
	errorsN := 0
	var total int64
	recordCount := 0
	for {
		res, err := recv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddErrorMsg(err.Error())))
			return
		}
		//progbar.SetTotal(decl.Name+"@"+decl.Version, int64(res.TableCount))
		//progbar.Current(decl.Name+"@"+decl.Version, int64(len(res.FinishedTables)), res.Table)
		total = int64(res.TableCount)
		if res.Diagnostics != nil {
			//if res.Diagnostics.HasError() {
			//	cli_ui.SaveLogToDiagnostic(res.Diagnostics.GetDiagnosticSlice())
			//}
			x.sendMessage(x.addProviderNameForMessage(plan, res.Diagnostics))
		}

		// count record pull
		if utils.NotHasError(res.Diagnostics) {
			recordCount++
		}

		success = len(res.FinishedTables)
		errorsN = 0

		x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s resource fetch %d/%d, finished task count %d ...", plan.String(), success, total, recordCount)))
	}
	_ = success
	_ = total
	//x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s fetch %d/%d, record count %d ...", plan.String(), success, total, recordCount)))
	//progbar.ReceiverWait(decl.Name + "@" + decl.Version)
	if errorsN > 0 {
		//cli_ui.Errorf("\nPull complete! Total Resources pulled:%d        Errors: %d\n", success, errorsN)
		//return nil
		return
	}
	//cli_ui.Infof("\nPull complete! Total Resources pulled:%d        Errors: %d\n", success, errorsN)
	//return nil
	x.sendMessage(x.addProviderNameForMessage(plan, schema.NewDiagnostics().AddInfo("Provider %s fetch done", plan.String())))

	// save table pull time
	d = x.refreshPullTableTime(ctx, databaseStorage, plan, needFetchTableSet)
	if utils.IsNotEmpty(d) {
		x.executor.options.MessageChannel.Send(d)
	}

	return
}

func (x *ProviderFetchExecutorWorker) addProviderNameForMessage(plan *planner.ProviderFetchPlan, d *schema.Diagnostics) *schema.Diagnostics {
	if d == nil {
		return nil
	}
	diagnostics := schema.NewDiagnostics()
	for _, item := range d.GetDiagnosticSlice() {
		diagnostics.AddDiagnostic(schema.NewDiagnostic(item.Level(), fmt.Sprintf("Provider %s say: %s", plan.String(), item.Content())))
	}
	return diagnostics
}

func (x *ProviderFetchExecutorWorker) sendMessage(message *schema.Diagnostics) {
	x.executor.options.MessageChannel.Send(message)
}

// ------------------------------------------------- --------------------------------------------------------------------

// An attempt is made to hit the cache of the data pull, and if the cache can be hit, the previous data is used instead of a repeat pull
func (x *ProviderFetchExecutorWorker) tryHitCache(ctx context.Context, databaseStorage storage.Storage, plan *planner.ProviderFetchPlan, providerInformation *shard.GetProviderInformationResponse) (bool, map[string]struct{}) {

	// Step 01. Calculate all the root tables that need to be pulled
	tooRootTableMap := x.makeToRootTableMap(providerInformation)
	needFetchTableNameSet := map[string]struct{}{}
	//  If resource is specified, only the given resource is pulled
	if plan.ProviderConfigurationBlock != nil && len(plan.ProviderConfigurationBlock.Resources) != 0 {
		for _, tableName := range plan.ProviderConfigurationBlock.Resources {
			needFetchTableNameSet[tooRootTableMap[tableName]] = struct{}{}
		}
	} else {
		// Otherwise, all resources of this provider are pulled by default
		for _, table := range providerInformation.Tables {
			needFetchTableNameSet[table.TableName] = struct{}{}
		}
	}

	//  If caching is not enabled, return directly
	if !x.isEnableFetchCache(ctx, databaseStorage, plan) {
		return false, needFetchTableNameSet
	}

	cache, diagnostics := x.computeAllNeedPullTableCanHitCache(ctx, databaseStorage, plan, needFetchTableNameSet)
	x.executor.options.MessageChannel.Send(diagnostics)
	return cache, needFetchTableNameSet
}

func (x *ProviderFetchExecutorWorker) isEnableFetchCache(ctx context.Context, storage storage.Storage, plan *planner.ProviderFetchPlan) bool {
	if plan == nil || plan.ProviderConfigurationBlock == nil || plan.ProviderConfigurationBlock.Cache == "" {
		return false
	}
	return true
}

// Calculate all the tables that need to be pulled
func (x *ProviderFetchExecutorWorker) computeAllNeedPullTableCanHitCache(ctx context.Context, storage storage.Storage, plan *planner.ProviderFetchPlan, needFetchTableNameSet map[string]struct{}) (bool, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// Step 02. Resolve whether it is expired
	duration, err := module.ParseDuration(plan.ProviderConfigurationBlock.Cache)
	if err != nil {
		return false, schema.NewDiagnostics().AddErrorMsg("Parse cache duration failed: %s", err.Error())
	}
	databaseTime, err := storage.GetTime(ctx)
	if err != nil {
		return false, schema.NewDiagnostics().AddErrorMsg("Get database time failed: %s", err.Error())
	}

	// The expiration time of the cache in the table

	// Step 03.
	pullTaskId := ""
	for tableName := range needFetchTableNameSet {
		information, d := pgstorage.ReadTableCacheInformation(ctx, storage, tableName)
		if utils.HasError(d) {
			logger.ErrorF("read table cache information error: %s", d.String())
			return false, d
		}
		if information == nil {
			logger.ErrorF("read table cache information nil")
			return false, x.addProviderNameForMessage(plan, diagnostics.AddInfo("Table %s did not find cache information, still need pull table", tableName))
		}

		// It has to be from the same batch
		if pullTaskId == "" {
			pullTaskId = information.LastPullId
		} else if pullTaskId != information.LastPullId {
			return false, x.addProviderNameForMessage(plan, diagnostics.AddInfo("Table %s is not in the same period as the previous data pull, so the cache cannot be hit, still need pull table", tableName))
		}

		if information.LastPullTime.Add(duration).Before(databaseTime) {
			return false, x.addProviderNameForMessage(plan, diagnostics.AddInfo("Table %s pulls data that is out of date, still need pull table, last pull time %s, database now time %s, cache %s",
				tableName, information.LastPullTime.String(), databaseTime.String(), duration.String()))
		}

		// ok, this table can hit cache

	}

	// ok, all table can hit cache
	return true, nil
}

// Expand the forest of all tables of the provider into a mapping table from the current table name to the root table name
func (x *ProviderFetchExecutorWorker) makeToRootTableMap(providerInformation *shard.GetProviderInformationResponse) map[string]string {
	tableRootMap := make(map[string]string, 0)
	for rootTableName, rootTable := range providerInformation.Tables {
		for _, tableName := range x.flatTable(rootTable) {
			tableRootMap[tableName] = rootTableName
		}
	}
	return tableRootMap
}

func (x *ProviderFetchExecutorWorker) flatTable(table *schema.Table) []string {
	if table == nil {
		return nil
	}
	tableNameSlice := []string{table.TableName}
	for _, subTables := range table.SubTables {
		tableNameSlice = append(tableNameSlice, x.flatTable(subTables)...)
	}
	return tableNameSlice
}

func (x *ProviderFetchExecutorWorker) refreshPullTableTime(ctx context.Context, databaseStorage storage.Storage, plan *planner.ProviderFetchPlan, needFetchTableNameSet map[string]struct{}) *schema.Diagnostics {
	diagnostics := schema.NewDiagnostics()
	pullId := id_util.RandomId()
	storageTime, err := databaseStorage.GetTime(ctx)
	if err != nil {
		return diagnostics.AddErrorMsg("Get storage time error: %s", err.Error())
	}
	for tableName := range needFetchTableNameSet {
		information := &pgstorage.TableCacheInformation{
			TableName:    tableName,
			LastPullId:   pullId,
			LastPullTime: storageTime,
		}
		d := pgstorage.SaveTableCacheInformation(ctx, databaseStorage, information)
		if diagnostics.AddDiagnostics(d).HasError() {
			return diagnostics
		}
	}
	return diagnostics
}

// ------------------------------------------------- --------------------------------------------------------------------
