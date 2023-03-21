package executors

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
)

// Executor Used to execute Module
type Executor interface {
	Name() string

	Execute(ctx context.Context) *schema.Diagnostics
}
