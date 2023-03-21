package telemetry

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/json_util"
	"github.com/selefra/selefra/pkg/logger"
	"github.com/selefra/selefra/pkg/utils"
	"sync"
)

// ------------------------------------------------ ---------------------------------------------------------------------

// TelemetryEnable Whether to enable usage data reporting
var TelemetryEnable = true

// ------------------------------------------------ ---------------------------------------------------------------------

// Analytics Represents an interface for analysis
type Analytics interface {

	// Init Initialization analyzer
	Init(ctx context.Context) *schema.Diagnostics

	// Submit the information to be collected
	Submit(ctx context.Context, event *Event) *schema.Diagnostics

	// Close Turn off analyzer
	Close(ctx context.Context) *schema.Diagnostics
}

// ------------------------------------------------ ---------------------------------------------------------------------

type Event struct {
	Name       string         `json:"name"`
	PayloadMap map[string]any `json:"payload_map"`
}

func NewEvent(name string) *Event {
	return &Event{
		Name:       name,
		PayloadMap: make(map[string]any, 0),
	}
}

func (x *Event) SetName(name string) *Event {
	x.Name = name
	return x
}

func (x *Event) Add(name string, value any) *Event {
	x.PayloadMap[name] = value
	return x
}

func (x *Event) ToJsonString() string {
	return json_util.ToJsonString(x)
}

// ------------------------------------------------ ---------------------------------------------------------------------

var DefaultAnalytics Analytics
var InitOnce sync.Once

func Init(ctx context.Context) *schema.Diagnostics {

	if !TelemetryEnable {
		return nil
	}

	DefaultAnalytics = &RudderstackAnalytics{}
	return DefaultAnalytics.Init(ctx)
}

func Submit(ctx context.Context, event *Event) *schema.Diagnostics {

	if !TelemetryEnable {
		return nil
	}

	InitOnce.Do(func() {
		d := Init(context.Background())
		if utils.HasError(d) {
			logger.ErrorF("init telemetry, msg: %s", d.String())
		} else {
			logger.InfoF("init telemetry success")
		}
	})

	return DefaultAnalytics.Submit(ctx, event)
}

func Close(ctx context.Context) *schema.Diagnostics {

	if !TelemetryEnable {
		return nil
	}

	if DefaultAnalytics != nil {
		return DefaultAnalytics.Close(ctx)
	} else {
		return nil
	}
}
