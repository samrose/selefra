package telemetry

import (
	"context"
	"github.com/rudderlabs/analytics-go/v4"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra/pkg/cli_env"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/selefra_workspace"
)

type RudderstackAnalytics struct {
	client   analytics.Client
	deviceId string
}

var _ Analytics = &RudderstackAnalytics{}

func (x *RudderstackAnalytics) Init(ctx context.Context) *schema.Diagnostics {
	// Instantiates a client to use send messages to the Rudder API.
	token := cli_env.GetSelefraTelemetryToken()
	if token == "" {
		logger.ErrorF("can not find SELEFRA_TELEMETRY_TOKEN")
		return schema.NewDiagnostics().AddErrorMsg("you must use env SELEFRA_TELEMETRY_TOKEN set you Rudderstack write key")
	}
	client := analytics.New(token, "https://selefralefsm.dataplane.rudderstack.com")

	x.client = client

	deviceId, diagnostics := selefra_workspace.GetDeviceID()
	x.deviceId = deviceId

	return diagnostics
}

func (x *RudderstackAnalytics) Submit(ctx context.Context, event *Event) *schema.Diagnostics {
	if x.client == nil {
		return nil
	}
	err := x.client.Enqueue(analytics.Track{
		AnonymousId: x.deviceId,
		MessageId:   id_util.RandomId(),
		Event:       event.Name,
		Properties:  event.PayloadMap,
	})
	return schema.NewDiagnostics().AddError(err)
}

func (x *RudderstackAnalytics) Close(ctx context.Context) *schema.Diagnostics {
	if x.client != nil {
		err := x.client.Close()
		if err != nil {
			return schema.NewDiagnostics().AddErrorMsg("close Rudderstack client failed: %s", err.Error())
		}
	}
	return nil
}
