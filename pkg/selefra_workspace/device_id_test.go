package selefra_workspace

import (
	"github.com/selefra/selefra/pkg/utils"
	"testing"
)

func TestGetDeviceID(t *testing.T) {
	id, diagnostics := GetDeviceID()
	if utils.HasError(diagnostics) {
		t.Log(diagnostics.String())
	}
	t.Log(id)
}
