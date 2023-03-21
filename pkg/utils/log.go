package utils

import "github.com/selefra/selefra-provider-sdk/provider/schema"

func HasError(diagnostics *schema.Diagnostics) bool {
	return diagnostics != nil && diagnostics.HasError()
}

func NotHasError(diagnostics *schema.Diagnostics) bool {
	return !HasError(diagnostics)
}

func IsNotEmpty(diagnostics *schema.Diagnostics) bool {
	return diagnostics != nil && !diagnostics.IsEmpty()
}

func IsEmpty(diagnostics *schema.Diagnostics) bool {
	return diagnostics == nil || diagnostics.IsEmpty()
}
