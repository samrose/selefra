package query

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra-provider-sdk/storage_factory"
	"github.com/selefra/selefra/cli_ui"
	"testing"
)

func TestNewQueryClient(t *testing.T) {
	ctx := context.Background()

	options := postgresql_storage.NewPostgresqlStorageOptions(env.GetDatabaseDsn())
	storage, diagnostics := storage_factory.NewStorage(context.Background(), storage_factory.StorageTypePostgresql, options)
	if err := cli_ui.PrintDiagnostics(diagnostics); err != nil {
		return
	}
	queryClient, _ := NewQueryClient(ctx, storage_factory.StorageTypePostgresql, storage)
	if queryClient == nil {
		t.Error("queryClient is nil")
	}
	queryClient.Run(context.Background())
}

//func TestNewQueryClientOnline(t *testing.T) {
//	ctx := context.Background()
//	global.Init("query", global.WithWorkspace("../../tests/workspace/online"))
//	global.SetToken("xxxxxxxxxxxxxxxxxxxxxx")
//	global.SERVER = "dev-api.selefra.io"
//
//	queryClient, _ := NewQueryClient(ctx)
//	if queryClient == nil {
//		t.Error("queryClient is nil")
//	}
//}

//func TestCreateColumnsSuggest(t *testing.T) {
//	ctx := context.Background()
//	global.Init("go_test", global.WithWorkspace("../../tests/workspace/offline"))
//	cof, err := config.GetConfig()
//	if err != nil {
//		cli_ui.Errorln(err)
//	}
//	for i := range cof.Selefra.ProviderDecls {
//		confs, err := tools.ProviderConfigStrs(cof, cof.Selefra.ProviderDecls[i].Name)
//		if err != nil {
//			cli_ui.Errorln(err.Error())
//		}
//		for _, conf := range confs {
//			var cp config.ProviderBlock
//			err := json.Unmarshal([]byte(conf), &cp)
//			if err != nil {
//				cli_ui.Errorln(err.Error())
//				continue
//			}
//			//ctx, c, err := createCtxAndClient(*cof, cof.Selefra.RequireProvidersBlock[i], cp)
//			//if err != nil {
//			//	t.Error(err)
//			//}
//			sto, _ := pgstorage.Storage(ctx)
//			columns := initColumnsSuggest(ctx, sto)
//			if columns == nil {
//				t.Error("Columns is nil")
//			}
//		}
//	}
//}
//
//func TestCreateTablesSuggest(t *testing.T) {
//	ctx := context.Background()
//	global.Init("go_test", global.WithWorkspace("../../tests/workspace/offline"))
//	cof, err := config.GetConfig()
//	if err != nil {
//		cli_ui.Errorln(err)
//	}
//	for i := range cof.Selefra.ProviderDecls {
//		confs, err := tools.ProviderConfigStrs(cof, cof.Selefra.ProviderDecls[i].Name)
//		if err != nil {
//			cli_ui.Errorln(err.Error())
//		}
//		for _, conf := range confs {
//			var cp config.ProviderBlock
//			err := json.Unmarshal([]byte(conf), &cp)
//			if err != nil {
//				cli_ui.Errorln(err.Error())
//				continue
//			}
//			sto, _ := pgstorage.Storage(ctx)
//			tables := initTablesSuggest(ctx, sto)
//			if tables == nil {
//				t.Error("Tables is nil")
//			}
//		}
//	}
//}
