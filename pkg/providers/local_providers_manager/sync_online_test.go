package local_providers_manager

//import (
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestSyncOnline(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping test in short mode.")
//		return
//	}
//	global.SERVER = "dev-api.selefra.io"
//	global.LOGINTOKEN = "xxxxxxxxxxxxxxxxxxxxxx"
//	*global.WORKSPACE = "../../tests/workspace/online"
//	errLogs, _, err := Sync()
//	if err != nil {
//		t.Error(err)
//	}
//	if len(errLogs) != 0 {
//		t.Error(errLogs)
//	}
//}
