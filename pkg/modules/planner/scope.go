package planner

import (
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/utils"
)

// Scope Used to represent the scope of a module, scope have some variables can use
type Scope struct {

	// Variable in scope, may be is self declare, or extend from parent scope
	variablesMap map[string]any

	// The provider configuration information in scope, now extend from parent module, do not support custom by self
	providerConfigBlockSlice []*module.ProviderBlock
}

// ExtendScope create scope from exists scope
func ExtendScope(scope *Scope) *Scope {
	subScope := NewScope()
	subScope.Extend(scope)
	return subScope
}

// NewScope create new scope
func NewScope() *Scope {
	return &Scope{
		variablesMap: make(map[string]any),
	}
}

// Extend current scope extend other scope
func (x *Scope) Extend(scope *Scope) {
	for key, value := range scope.variablesMap {
		if _, exists := x.variablesMap[key]; exists {
			continue
		}
		x.variablesMap[key] = value
	}
}

// Clone Make a copy of the current scope
func (x *Scope) Clone() *Scope {

	newVariablesMap := make(map[string]any)
	for key, value := range x.variablesMap {
		newVariablesMap[key] = value
	}

	return &Scope{
		variablesMap:             newVariablesMap,
		providerConfigBlockSlice: x.providerConfigBlockSlice,
	}
}

// GetVariable Gets the value of a variable
func (x *Scope) GetVariable(variableName string) (any, bool) {
	value, exists := x.variablesMap[variableName]
	return value, exists
}

// SetVariable Declare a variable
func (x *Scope) SetVariable(variableName string, variableValue any) any {
	oldValue := x.variablesMap[variableName]
	x.variablesMap[variableName] = variableValue
	return oldValue
}

// SetVariables Batch declaration variable
func (x *Scope) SetVariables(variablesMap map[string]any) {
	for variableName, variableValue := range variablesMap {
		x.variablesMap[variableName] = variableValue
	}
}

// SetVariableIfNotExists Declared only if the variable does not exist
func (x *Scope) SetVariableIfNotExists(variableName string, variableValue any) bool {
	if _, exists := x.variablesMap[variableName]; exists {
		return false
	}
	x.variablesMap[variableName] = variableValue
	return true
}

// RenderingTemplate Rendering the template using the moduleScope of the current module
func (x *Scope) RenderingTemplate(templateName, templateString string) (string, error) {
	// TODO a problem in here, they call it "no value" ?
	return utils.RenderingTemplate(templateName, templateString, x.variablesMap)
}
