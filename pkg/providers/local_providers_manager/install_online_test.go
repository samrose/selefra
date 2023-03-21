package local_providers_manager

//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestInstallOnline(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping test in short mode.")
//		return
//	}
//	global.SERVER = "dev-api.selefra.io"
//	global.LOGINTOKEN = "xxxxxxxxxxxxxxxxxxxxxx"
//	*global.WORKSPACE = "../../tests/workspace/online"
//	ctx := context.Background()
//	err := install(ctx, []string{"aws@latest"})
//	if err != nil {
//		t.Error(err)
//	}
//}
