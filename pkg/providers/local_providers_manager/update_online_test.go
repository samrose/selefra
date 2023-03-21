package local_providers_manager

//import (
//	"context"
//	"github.com/selefra/selefra/cmd/provider"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestUpdateOnline(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping test in short mode.")
//		return
//	}
//	global.SERVER = "dev-api.selefra.io"
//	global.LOGINTOKEN = "xxxxxxxxxxxxxxxxxxxxxx"
//	*global.WORKSPACE = "../../tests/workspace/online"
//	ctx := context.Background()
//	arg := []string{"aws"}
//	err := provider.Upgrade(ctx, arg)
//	if err != nil {
//		t.Error(err)
//	}
//}
