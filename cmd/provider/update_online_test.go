package provider

//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestUpdateOnline(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping test in short mode.")
//		return
//	}
//	global.Init("TestRemoveOnline", global.WithWorkspace("../../tests/workspace/online"))
//	global.SetToken("xxxxxxxxxxxxxxxxxxxxxx")
//	global.SERVER = "dev-api.selefra.io"
//	ctx := context.Background()
//	arg := []string{"aws"}
//	err := update(ctx, arg)
//	if err != nil {
//		t.Error(err)
//	}
//}
