package module

import (
	"github.com/selefra/selefra-provider-sdk/provider/schema"
)

// ------------------------------------------------- --------------------------------------------------------------------

// Block each block should implement this interface
type Block interface {

	// Validator block should be able to check that it's configuration is correct
	Validator

	// Locatable every block should be addressable, so you have to be able to figure out where the text is in that block
	// yaml to module parser fills in the location of the Block, so you can get the original location and content of the Block when you need it
	// The location information should not change and should be fixed once parsed
	Locatable

	// IsEmpty Determines whether the block is empty
	IsEmpty() bool
}

// ------------------------------------------------- --------------------------------------------------------------------

// Validator A validator that supports checking
type Validator interface {

	// Check whether the node configuration is correct
	Check(module *Module, validatorContext *ValidatorContext) *schema.Diagnostics
}

// ------------------------------------------------- --------------------------------------------------------------------

// MergableBlock Used to indicate that a block is merge
type MergableBlock[T Block] interface {

	// Merge Used to merge two identical blocks
	Merge(other T) (T, *schema.Diagnostics)
}

// ------------------------------------------------- --------------------------------------------------------------------

// HaveRuntime Some blocks may have a runtime to handle more complex logic
type HaveRuntime[T any] interface {

	// Runtime Returns the runtime corresponding to the block
	Runtime() T
}

// ------------------------------------------------- --------------------------------------------------------------------
