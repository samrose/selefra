package provider

//import (
//	"context"
//	"fmt"
//	"github.com/selefra/selefra-provider-sdk/provider/schema"
//	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
//	"github.com/selefra/selefra-utils/pkg/id_util"
//	"github.com/selefra/selefra/cli_ui"
//	"github.com/selefra/selefra/cmd/fetch"
//	"github.com/selefra/selefra/cmd/test"
//	"github.com/selefra/selefra/cmd/tools"
//	"github.com/selefra/selefra/config"
//	"github.com/selefra/selefra/global"
//	"github.com/selefra/selefra/pkg/cloud_sdk"
//	"github.com/selefra/selefra/pkg/logger"
//	"github.com/selefra/selefra/pkg/registry"
//	"github.com/selefra/selefra/pkg/storage/pgstorage"
//	"github.com/selefra/selefra/pkg/utils"
//	"path/filepath"
//	"time"
//)
//
//type lockStruct struct {
//	SchemaKey string
//	Uuid      string
//	Storage   *postgresql_storage.PostgresqlStorage
//}
//
//// effectiveDecls check provider decls and download provider binary file, return the effective providers
//func effectiveDecls(ctx context.Context, decls []*config.RequireProvider) (effects []*config.RequireProvider, errlogs []string) {
//	namespace, _, err := utils.Home()
//	if err != nil {
//		errlogs = append(errlogs, err.Error())
//		return
//	}
//	provider := registry.NewProviderRegistry(namespace)
//	cli_ui.Infof("Selefra has been successfully installed providers!\n\n")
//	cli_ui.Infof("Checking Selefra provider updates......\n")
//
//	for _, decl := range decls {
//		configVersion := decl.Version
//		prov := registry.Provider{
//			Name:    decl.Name,
//			Version: decl.Version,
//			Source:  "",
//			Path:    decl.Path,
//		}
//		pp, err := provider.Download(ctx, prov, true)
//		if err != nil {
//			cli_ui.Errorf("%s@%s failed updated：%s", decl.Name, decl.Version, err.Error())
//			errlogs = append(errlogs, err.Error())
//			continue
//		} else {
//			decl.Path = pp.Filepath
//			decl.Version = pp.Version
//			err = tools.AppendProviderDecl(pp, nil, configVersion)
//			if err != nil {
//				cli_ui.Errorf("%s@%s failed updated：%s", decl.Name, decl.Version, err.Error())
//				errlogs = append(errlogs, err.Error())
//				continue
//			}
//			effects = append(effects, decl)
//			cli_ui.Infof("	%s@%s all ready updated!\n", decl.Name, decl.Version)
//		}
//	}
//
//	return effects, nil
//}
//
//func Sync(ctx context.Context) (lockSlice []lockStruct, err error) {
//	// load and check config
//	cli_ui.Infof("Initializing provider plugins...\n\n")
//	rootConfig, err := config.GetConfig()
//	if err != nil {
//		return nil, err
//	}
//
//	if err = test.CheckSelefraConfig(ctx, rootConfig); err != nil {
//		_ = http_client.TrySetUpStage(global.RelvPrjName(), http_client.Failed)
//		return nil, err
//	}
//
//	if _, err := cloud_sdk.UploadLogStatus(); err != nil {
//		cli_ui.Errorln(err.Error())
//	}
//
//	var errored bool
//
//	providerDecls, errLogs := effectiveDecls(ctx, rootConfig.Selefra.ProviderDecls)
//
//	cli_ui.Infof("Selefra has been finished update providers!\n")
//
//	global.SetStage("pull")
//	for _, decl := range providerDecls {
//		prvds := tools.ProvidersByID(rootConfig, decl.Name)
//		for _, prvd := range prvds {
//
//			// build a postgresql storage
//			schemaKey := config.GetSchemaKey(decl, *prvd)
//			store, err := pgstorage.PgStorageWithMeta(ctx, &schema.ClientMeta{
//				ClientLogger: logger.NewSchemaLogger(),
//			}, pgstorage.WithSearchPath(config.GetSchemaKey(decl, *prvd)))
//			if err != nil {
//				errored = true
//				cli_ui.Errorf("%s@%s failed updated：%s", decl.Name, decl.Version, err.Error())
//				errLogs = append(errLogs, fmt.Sprintf("%s@%s failed updated：%s", decl.Name, decl.Version, err.Error()))
//				continue
//			}
//
//			// try lock
//			// TODO: check unlock
//			uuid := id_util.RandomId()
//			for {
//				err = store.Lock(ctx, schemaKey, uuid)
//				if err == nil {
//					lockSlice = append(lockSlice, lockStruct{
//						SchemaKey: schemaKey,
//						Uuid:      uuid,
//						Storage:   store,
//					})
//					break
//				}
//				time.Sleep(5 * time.Second)
//			}
//
//			// check if cache expired
//			expired, _ := tools.CacheExpired(ctx, store, prvd.Cache)
//			if !expired {
//				cli_ui.Infof("%s %s@%s pull infrastructure data:\n", prvd.Name, decl.Name, decl.Version)
//				cli_ui.Print(fmt.Sprintf("Pulling %s@%s Please wait for resource information ...", decl.Name, decl.Version), false)
//				cli_ui.Infof("	%s@%s all ready use cache!\n", decl.Name, decl.Version)
//				continue
//			}
//
//			// if expired, fetch new data
//			err = fetch.Fetch(ctx, decl, prvd)
//			if err != nil {
//				cli_ui.Errorf("%s %s Synchronization failed：%s", decl.Name, decl.Version, err.Error())
//				errored = true
//				continue
//			}
//
//			// set fetch time
//			if err := pgstorage.SetStorageValue(ctx, store, config.GetCacheKey(), time.Now().Format(time.RFC3339)); err != nil {
//				cli_ui.Warningf("%s %s set cache time failed：%s", decl.Name, decl.Version, err.Error())
//				errored = true
//				continue
//			}
//		}
//	}
//	if errored {
//		cli_ui.Errorf(`
//This may be exception, view detailed exception in %s .
//`, filepath.Join(global.WorkSpace(), "logs"))
//	}
//
//	return lockSlice, nil
//}
