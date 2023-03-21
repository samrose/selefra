package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestRemove(t *testing.T) {
//	*global.WORKSPACE = "../../tests/workspace/offline"
//	err := RemoveProviders([]string{"aws"})
//	if err != nil {
//		t.Error(err)
//	}
//	err = install(context.Background(), []string{"aws@latest"})
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestLocalProvidersManager_RemoveProviders(t *testing.T) {
	d := getTestLocalProviderManager().RemoveProviders(context.Background(), "aws@v0.0.5")
	if utils.HasError(d) {
		t.Log(d.ToString())
	}
	assert.False(t, utils.HasError(d))
}
