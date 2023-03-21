package local_providers_manager
//
//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestRemoveOnline(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping test in short mode.")
//		return
//	}
//	global.SERVER = "dev-api.selefra.io"
//	global.LOGINTOKEN = "xxxxxxxxxxxxxxxxxxxxxx"
//	*global.WORKSPACE = "../../tests/workspace/online"
//	err := RemoveProviders([]string{"aws"})
//	if err != nil {
//		t.Error(err)
//	}
//	err = install(context.Background(), []string{"aws@latest"})
//	if err != nil {
//		t.Error(err)
//	}
//}
