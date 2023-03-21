package module

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-utils/pkg/reflect_util"
	"reflect"
)

// ------------------------------------------------- --------------------------------------------------------------------

// Module Represents information about a module
type Module struct {

	// Several root-level blocks of a module
	SelefraBlock   *SelefraBlock
	ModulesBlock   ModulesBlock
	ProvidersBlock ProvidersBlock
	RulesBlock     RulesBlock
	VariablesBlock VariablesBlock

	// Parent of the current module
	ParentModule *Module

	// What are the submodules of the current module, [subModuleName, *subModule]
	// Keep the order of references
	SubModules []*Module

	// The source of the module, in fact, is the string written inside use
	// The source of the root module is the current path
	Source string
	// Local path of the module
	ModuleLocalDirectory string

	// How is the dependency from the top-level module to the current module, in fact, all the way to use the concatenation
	DependenciesPath []string
}

func NewModule() *Module {
	return &Module{}
}

// ------------------------------------------------- --------------------------------------------------------------------

// BuildFullName The full path name of the module, which can be understood at a glance
func (x *Module) BuildFullName() string {
	if x.Source == "" {
		return x.ModuleLocalDirectory
	} else {
		return fmt.Sprintf("%s @ %s", x.Source, x.ModuleLocalDirectory)
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// HasRequiredProviderName check whether the required provider name is available
func (x *Module) HasRequiredProviderName(requiredProviderName string) bool {
	for _, requiredProvider := range x.SelefraBlock.RequireProvidersBlock {
		if requiredProvider.Name == requiredProviderName {
			return true
		}
	}
	return false
}

// ListRequiredProvidersName List the names of all required providers
func (x *Module) ListRequiredProvidersName() []string {
	requiredProviderNameSlice := make([]string, len(x.SelefraBlock.RequireProvidersBlock))
	for index, requiredProvider := range x.SelefraBlock.RequireProvidersBlock {
		requiredProviderNameSlice[index] = requiredProvider.Name
	}
	return requiredProviderNameSlice
}

// ------------------------------------------------- --------------------------------------------------------------------

// Merge the two modules into a new module
func (x *Module) Merge(other *Module) (*Module, *schema.Diagnostics) {

	diagnostics := schema.NewDiagnostics()

	// Only independent, unrelated modules can be merged, as if they were different configuration files in the same path
	if x.ParentModule != nil || len(x.SubModules) != 0 || other.ParentModule != nil || len(other.SubModules) != 0 {
		return nil, diagnostics.AddErrorMsg("can not merge module it have parent module or submodules")
	}

	mergedModule := NewModule()
	// The blocks at the root level are merged one by one
	mergedModule.SelefraBlock = MergeBlockWithDiagnostics(x.SelefraBlock, other.SelefraBlock, diagnostics)
	mergedModule.ModulesBlock = MergeBlockWithDiagnostics(x.ModulesBlock, other.ModulesBlock, diagnostics)
	mergedModule.ProvidersBlock = MergeBlockWithDiagnostics(x.ProvidersBlock, other.ProvidersBlock, diagnostics)
	mergedModule.RulesBlock = MergeBlockWithDiagnostics(x.RulesBlock, other.RulesBlock, diagnostics)
	mergedModule.VariablesBlock = MergeBlockWithDiagnostics(x.VariablesBlock, other.VariablesBlock, diagnostics)

	return mergedModule, diagnostics
}

// MergeBlockWithDiagnostics Merge two blocks
func MergeBlockWithDiagnostics[T Block](blockA, blockB T, diagnostics *schema.Diagnostics) T {
	var zero T
	if !reflect_util.IsNil(blockA) && !reflect_util.IsNil(blockB) {
		reflectValueA := reflect.ValueOf(blockA)
		if reflectValueA.CanInterface() {
			mergableBlockA, ok := reflectValueA.Interface().(MergableBlock[T])
			if !ok {
				// TODO error message
				diagnostics.AddErrorMsg("can not convert block to MergableBlock")
				return zero
			}
			merge, d := mergableBlockA.Merge(blockB)
			diagnostics.AddDiagnostics(d)
			if d == nil || !d.HasError() {
				return merge
			} else {
				return zero
			}
		} else {
			// TODO build humanreadable error message
			diagnostics.AddErrorMsg("can not convert block to MergableBlock")
			return zero
		}
	} else if reflect_util.IsNil(blockA) {
		return blockB
	} else {
		return blockA
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

func (x *Module) Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	if x.ModulesBlock != nil {
		diagnostics.AddDiagnostics(x.ModulesBlock.Check(x, validatorContext))
	}

	if x.ProvidersBlock != nil {
		diagnostics.AddDiagnostics(x.ProvidersBlock.Check(x, validatorContext))
	}

	if x.SelefraBlock != nil {
		diagnostics.AddDiagnostics(x.SelefraBlock.Check(x, validatorContext))
	} else {
		errorTips := fmt.Sprintf("module %s selefra block must can not lack", x.BuildFullName())
		diagnostics.AddErrorMsg(errorTips)
	}

	if x.RulesBlock != nil {
		diagnostics.AddDiagnostics(x.RulesBlock.Check(x, validatorContext))
	}

	if x.VariablesBlock != nil {
		diagnostics.AddDiagnostics(x.VariablesBlock.Check(x, validatorContext))
	}

	// check submodules
	for _, subModule := range x.SubModules {
		diagnostics.AddDiagnostics(subModule.Check(subModule, validatorContext))
	}

	return diagnostics
}

// ------------------------------------------------- --------------------------------------------------------------------

type TraversalContext struct {
	ParentTraversalContext *TraversalContext

	ParentModule *Module
	Module       *Module
}

func (x *Module) Traversal(ctx context.Context, traversalFunc func(ctx context.Context, traversalContext *TraversalContext) bool) {
	x.internalTraversal(ctx, &TraversalContext{ParentTraversalContext: nil, ParentModule: nil, Module: x}, traversalFunc)
}

func (x *Module) internalTraversal(ctx context.Context, traversalContext *TraversalContext, traversalFunc func(ctx context.Context, traversalContext *TraversalContext) bool) {

	if !traversalFunc(ctx, traversalContext) {
		return
	}

	for _, subModule := range traversalContext.Module.SubModules {
		x.internalTraversal(ctx, &TraversalContext{ParentTraversalContext: traversalContext, ParentModule: traversalContext.Module, Module: subModule}, traversalFunc)
	}

}

// ------------------------------------------------ ---------------------------------------------------------------------
