package planner

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
)

// Planner Represents a planner that can generate a plan
type Planner[T any] interface {

	// Name The name of the planner
	Name() string

	// MakePlan Make a plan
	MakePlan(ctx context.Context) (T, *schema.Diagnostics)
}
