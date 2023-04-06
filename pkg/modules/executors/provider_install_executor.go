package executors

import (
	"context"
	"github.com/hashicorp/go-getter"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/planner"
	"github.com/selefra/selefra/pkg/providers/local_providers_manager"
	"github.com/selefra/selefra/pkg/utils"
)

// ------------------------------------------------- --------------------------------------------------------------------

// ProviderInstallExecutorOptions Install the provider's actuator
type ProviderInstallExecutorOptions struct {

	// The installation plan to execute
	Plans []*planner.ProviderInstallPlan

	// The path to install to
	DownloadWorkspace string

	// Receive real-time message feedback
	MessageChannel *message.Channel[*schema.Diagnostics]

	// Tracking installation progress
	ProgressTracker getter.ProgressTracker
}

// ------------------------------------------------- --------------------------------------------------------------------

const ProviderInstallExecutorName = "provider-install-executor"

type ProviderInstallExecutor struct {
	options *ProviderInstallExecutorOptions

	localProviderManager *local_providers_manager.LocalProvidersManager
}

var _ Executor = &ProviderInstallExecutor{}

func NewProviderInstallExecutor(options *ProviderInstallExecutorOptions) (*ProviderInstallExecutor, *schema.Diagnostics) {
	diagnostics := schema.NewDiagnostics()

	manager, err := local_providers_manager.NewLocalProvidersManager(options.DownloadWorkspace)
	if err != nil {
		return nil, diagnostics.AddErrorMsg(err.Error())
	}

	return &ProviderInstallExecutor{
		options:              options,
		localProviderManager: manager,
	}, diagnostics
}

// GetLocalProviderManager This way we can reuse the local provider manager
func (x *ProviderInstallExecutor) GetLocalProviderManager() *local_providers_manager.LocalProvidersManager {
	return x.localProviderManager
}

func (x *ProviderInstallExecutor) Name() string {
	return ProviderInstallExecutorName
}

func (x *ProviderInstallExecutor) Execute(ctx context.Context) *schema.Diagnostics {

	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
	}()

	diagnostics := schema.NewDiagnostics()
	for _, plan := range x.options.Plans {
		diagnostics.AddDiagnostics(x.executePlan(ctx, plan))
	}
	return diagnostics
}

func (x *ProviderInstallExecutor) executePlan(ctx context.Context, plan *planner.ProviderInstallPlan) *schema.Diagnostics {
	requiredProvider := &local_providers_manager.LocalProvider{
		Provider: plan.Provider,
	}
	installed, diagnostics := x.localProviderManager.IsProviderInstalled(ctx, requiredProvider)
	if utils.HasError(diagnostics) {
		return diagnostics
	}
	if installed {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\tProvider %s has installed", plan.String()))
		return nil
	}

	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("\tBegin downloading provider %s ...", plan.String()))

	x.localProviderManager.InstallProvider(ctx, &local_providers_manager.InstallProvidersOptions{
		RequiredProvider: requiredProvider,
		MessageChannel:   x.options.MessageChannel.MakeChildChannel(),
		ProgressTracker:  x.options.ProgressTracker,
	})

	// TODO init

	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Download & install provider %s success", plan.String()))

	return nil
}

// ------------------------------------------------- --------------------------------------------------------------------
