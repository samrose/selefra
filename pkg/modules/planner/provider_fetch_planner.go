package planner

import (
	"context"
	"errors"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra-provider-sdk/storage_factory"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/selefra_workspace"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"time"
)

// ------------------------------------------------- --------------------------------------------------------------------

// ProvidersFetchPlan The installation plan of a batch of providers
type ProvidersFetchPlan []*ProviderFetchPlan

// BuildProviderContextMap Create an execution context for the provider installation plan
func (x ProvidersFetchPlan) BuildProviderContextMap(ctx context.Context, DSN string) (map[string][]*ProviderContext, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	m := make(map[string][]*ProviderContext, 0)
	for _, plan := range x {

		//databaseSchema := pgstorage.GetSchemaKey(plan.Name, plan.Version, plan.ProviderConfigurationBlock)
		//options := postgresql_storage.NewPostgresqlStorageOptions(DSN)
		//options.SearchPath = databaseSchema

		options := postgresql_storage.NewPostgresqlStorageOptions(DSN)
		options.SearchPath = plan.FetchToDatabaseSchema

		databaseStorage, d := storage_factory.NewStorage(ctx, storage_factory.StorageTypePostgresql, options)
		if diagnostics.AddDiagnostics(d).HasError() {
			return nil, diagnostics
		}

		providerContext := &ProviderContext{
			ProviderName:          plan.Name,
			ProviderVersion:       plan.Version,
			DSN:                   DSN,
			Schema:                plan.FetchToDatabaseSchema,
			Storage:               databaseStorage,
			ProviderConfiguration: plan.ProviderConfigurationBlock,
		}
		m[plan.Name] = append(m[plan.Name], providerContext)
	}

	return m, diagnostics
}

// ProviderContext Ready execution strategy
type ProviderContext struct {

	// Which provider is it?
	ProviderName string

	// Which version
	ProviderVersion string

	DSN string

	// The database stored to
	Schema string

	// A connection to a database instance
	Storage storage.Storage

	// The provider configuration block
	ProviderConfiguration *module.ProviderBlock
}

// ------------------------------------------------- --------------------------------------------------------------------

const (
	DefaultMaxGoroutines = uint64(100)
)

// ProviderFetchPlan Indicates the pull plan of a provider
type ProviderFetchPlan struct {
	*ProviderInstallPlan

	// provider Configuration information used for fetching
	ProviderConfigurationBlock *module.ProviderBlock

	// Which schema to write data to
	FetchToDatabaseSchema string

	// The name of the configuration block to be used, which is left blank if not configured using a configuration file
	ProviderConfigurationName string

	// What is the MD5 of the configuration block if the provider configuration is used
	ProviderConfigurationMD5 string
}

func NewProviderFetchPlan(providerName, providerVersion string, providerBlock *module.ProviderBlock) *ProviderFetchPlan {
	return &ProviderFetchPlan{
		ProviderInstallPlan: &ProviderInstallPlan{
			Provider: registry.NewProvider(providerName, providerVersion),
		},
		ProviderConfigurationBlock: providerBlock,
	}
}

// GetProvidersConfigYamlString Obtain the configuration file for running the Provider
func (x *ProviderFetchPlan) GetProvidersConfigYamlString() string {
	if x.ProviderConfigurationBlock != nil {
		return x.ProviderConfigurationBlock.ProvidersConfigYamlString
	}
	return ""
}

// GetNeedPullTablesName Gets which tables to pull when pulling
func (x *ProviderFetchPlan) GetNeedPullTablesName() []string {
	tables := make([]string, 0)
	if x.ProviderConfigurationBlock != nil {
		tables = x.ProviderConfigurationBlock.Resources
	}
	if len(tables) == 0 {
		tables = append(tables, provider.AllTableNameWildcard)
	}
	return tables
}

// GetMaxGoroutines How many concurrency is used to pull the table data
func (x *ProviderFetchPlan) GetMaxGoroutines() uint64 {
	if x.ProviderConfigurationBlock != nil && x.ProviderConfigurationBlock.MaxGoroutines != nil {
		return *x.ProviderConfigurationBlock.MaxGoroutines
	} else {
		return DefaultMaxGoroutines
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderFetchPlannerOptions This parameter is required when creating the provider execution plan
type ProviderFetchPlannerOptions struct {

	// Which module is the execution plan being generated for
	Module *module.Module

	// Provider version that wins the vote
	ProviderVersionVoteWinnerMap map[string]string

	// DSNS are used to connect to the database to determine which schema to use when using environment variables
	DSN string

	// A place to send messages to the outside world
	MessageChannel *message.Channel[*schema.Diagnostics]
}

// ------------------------------------------------- --------------------------------------------------------------------

type ProviderFetchPlanner struct {
	options *ProviderFetchPlannerOptions
}

var _ Planner[ProvidersFetchPlan] = &ProviderFetchPlanner{}

func NewProviderFetchPlanner(options *ProviderFetchPlannerOptions) *ProviderFetchPlanner {
	return &ProviderFetchPlanner{
		options: options,
	}
}

func (x *ProviderFetchPlanner) Name() string {
	return "provider-fetch-planner"
}

func (x *ProviderFetchPlanner) MakePlan(ctx context.Context) (ProvidersFetchPlan, *schema.Diagnostics) {

	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
	}()

	return x.expandByConfiguration(ctx)
}

// Expand to multiple tasks based on the configuration
func (x *ProviderFetchPlanner) expandByConfiguration(ctx context.Context) ([]*ProviderFetchPlan, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()
	providerFetchPlanSlice := make([]*ProviderFetchPlan, 0)

	if x.options.Module.SelefraBlock == nil {
		return nil, diagnostics.AddErrorMsg("Module %s must have selefra block for make fetch plan", x.options.Module.BuildFullName())
	} else if len(x.options.Module.SelefraBlock.RequireProvidersBlock) == 0 {
		return nil, diagnostics.AddErrorMsg("Module %s selefra block not have providers block", x.options.Module.BuildFullName())
	}

	// Start a task for those that have a task written, some join by fetch start rule
	providerNamePlanCountMap := make(map[string]int, 0)
	nameToProviderMap := x.options.Module.SelefraBlock.RequireProvidersBlock.BuildNameToProviderBlockMap()
	for _, providerBlock := range x.options.Module.ProvidersBlock {

		// find required provider block
		requiredProviderBlock, exists := nameToProviderMap[providerBlock.Provider]
		if !exists {
			// selefra.providers block not found that name in providers[index] configuration
			errorTips := fmt.Sprintf("Provider name %s not found", providerBlock.Provider)
			diagnostics.AddErrorMsg(module.RenderErrorTemplate(errorTips, providerBlock.GetNodeLocation("provider"+module.NodeLocationSelfValue)))
			continue
		}

		// find use provider version
		providerWinnerVersion, exists := x.options.ProviderVersionVoteWinnerMap[requiredProviderBlock.Source]
		if !exists {
			errorTips := fmt.Sprintf("Provider version %s not found", requiredProviderBlock.Source)
			diagnostics.AddErrorMsg(module.RenderErrorTemplate(errorTips, requiredProviderBlock.GetNodeLocation("version")))
			continue
		}

		// Start a plan for the provider
		providerNamePlanCountMap[requiredProviderBlock.Source]++
		providerFetchPlan := NewProviderFetchPlan(requiredProviderBlock.Source, providerWinnerVersion, providerBlock)

		fetchToDatabaseSchema := pgstorage.GetSchemaKey(requiredProviderBlock.Source, providerWinnerVersion, providerBlock)
		providerFetchPlan.FetchToDatabaseSchema = fetchToDatabaseSchema
		providerFetchPlanSlice = append(providerFetchPlanSlice, providerFetchPlan)

	}
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	deviceID, d := selefra_workspace.GetDeviceID()
	if diagnostics.AddDiagnostics(d).HasError() {
		return nil, diagnostics
	}

	// See if there is another project that has not been activated, and if there is, start a pull plan for it as well
	for providerName, providerVersion := range x.options.ProviderVersionVoteWinnerMap {
		if providerNamePlanCountMap[providerName] > 0 {
			continue
		}

		providerFetchPlan := NewProviderFetchPlan(providerName, providerVersion, nil)
		fetchToDatabaseSchema, d := x.decideDatabaseSchemaForNoProviderBlockPlan(ctx, providerFetchPlan, deviceID)
		if diagnostics.AddDiagnostics(d).HasError() {
			continue
		}
		providerFetchPlan.FetchToDatabaseSchema = fetchToDatabaseSchema
		providerFetchPlanSlice = append(providerFetchPlanSlice, providerFetchPlan)
	}

	return providerFetchPlanSlice, diagnostics
}

// Generate schema names for pull plans that do not have provider blocks
func (x *ProviderFetchPlanner) decideDatabaseSchemaForNoProviderBlockPlan(ctx context.Context, plan *ProviderFetchPlan, deviceID string) (string, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// Verify that the database is available
	fetchToDatabaseSchema := pgstorage.GetSchemaKey(plan.Name, plan.Version, nil)
	pgstorage.WithSearchPath(fetchToDatabaseSchema)
	postgresqlOptions := postgresql_storage.NewPostgresqlStorageOptions(x.options.DSN)
	databaseStorage, d := storage_factory.NewStorage(ctx, storage_factory.StorageTypePostgresql, postgresqlOptions)
	if diagnostics.AddDiagnostics(d).HasError() {
		return "", diagnostics
	}
	// storage created must remember to close
	defer func() {
		databaseStorage.Close()
	}()
	owner, d := pgstorage.GetSchemaOwner(ctx, databaseStorage)
	if diagnostics.AddDiagnostics(d).HasError() {
		return "", diagnostics
	}
	if owner == nil {
		// This schema is still in unowned state. Try to get its attribution
		d := x.grabDatabaseSchema(ctx, plan, deviceID, databaseStorage)
		if diagnostics.AddDiagnostics(d).HasError() {
			return "", diagnostics
		}
		return fetchToDatabaseSchema, diagnostics
	}

	// If the schema is already occupied by someone, check to see if that person is yourself
	if owner.HolderID == deviceID {
		// If that person is yourself, then you can continue to use it
		return fetchToDatabaseSchema, diagnostics
	}

	// The previous schema is occupied, so you have to use your own separate schema
	fetchToDatabaseSchema = fetchToDatabaseSchema + "_" + deviceID
	return fetchToDatabaseSchema, diagnostics
}

// Use the database schema
// When a schema is assigned to the provider in the execution plan, the ownership of the schema is also marked for the provider to avoid schema ownership disputes during the execution phase
func (x *ProviderFetchPlanner) grabDatabaseSchema(ctx context.Context, plan *ProviderFetchPlan, deviceID string, storage storage.Storage) *schema.Diagnostics {

	lockOwnerId := utils.BuildLockOwnerId()
	tryTimes := 0

	for {

		x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner %s, make execute plan, begin try get database schema lock...", plan.String(), plan.FetchToDatabaseSchema, lockOwnerId))

		tryTimes++
		err := storage.Lock(ctx, pgstorage.LockId, lockOwnerId)
		if err != nil {
			x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, owner %s, make execute plan, get database schema lock error: %s, will sleep & retry, tryTimes = %d", plan.String(), plan.FetchToDatabaseSchema, lockOwnerId, err.Error(), tryTimes))
		} else {
			x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner %s, make execute plan, get database schema lock success", plan.String(), plan.FetchToDatabaseSchema, lockOwnerId))
			break
		}
		time.Sleep(time.Second * 10)
	}
	defer func() {
		for tryTimes := 0; tryTimes < 10; tryTimes++ {
			err := storage.UnLock(ctx, pgstorage.LockId, lockOwnerId)
			if err != nil {
				if errors.Is(err, postgresql_storage.ErrLockNotFound) {
					x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner = %s, release database schema lock success", plan.String(), plan.FetchToDatabaseSchema, lockOwnerId))
				} else {
					x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("Provider %s, schema %s, owner = %s, release database schema lock error: %s, will sleep & retry, tryTimes = %d", plan.String(), plan.FetchToDatabaseSchema, lockOwnerId, err.Error(), tryTimes))
				}
			} else {
				x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Provider %s, schema %s, owner = %s, release database schema lock success", plan.String(), plan.FetchToDatabaseSchema, lockOwnerId))
				break
			}
		}
	}()

	// You can hold this database, It's okay to hold the database, because you were the first one there
	// First set a tag bit to occupy this schema
	hostname, _ := os.Hostname()
	return pgstorage.SaveSchemaOwner(ctx, storage, &pgstorage.SchemaOwnerInformation{
		Hostname: hostname,
		HolderID: deviceID,
		// TODO If you are using a configuration file, put these two fields on the Settings
		ConfigurationName: "",
		ConfigurationMD5:  "",
	})
}

// ------------------------------------------------- --------------------------------------------------------------------
