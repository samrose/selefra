package local_providers_manager

//import (
//	"context"
//	"github.com/selefra/selefra-provider-sdk/provider/schema"
//)
//
//type UpgradeOptions struct {
//}
//
//func (x *LocalProvidersManager) Upgrade(ctx context.Context, providerNameSlice []string, messageChannel chan *schema.Diagnostics) {
//
//	defer func() {
//		close(messageChannel)
//	}()
//
//}

//func Upgrade(ctx context.Context, providerNameSlice []string, messageChannel chan *schema.Diagnostics) {
//
//	defer func() {
//		close(messageChannel)
//	}()
//
//	diagnostics := schema.NewDiagnostics()
//
//	err := config.IsSelefraWorkspace()
//	if err != nil {
//		messageChannel <- diagnostics.AddErrorMsg(err.Error())
//		return
//	}
//
//	var cof = &config.SelefraBlock{}
//	err = cof.Get()
//	if err != nil {
//		messageChannel <- diagnostics.AddErrorMsg(err.Error())
//		return
//	}
//
//	providerNameMap := make(map[string]struct{})
//	for _, providerName := range providerNameSlice {
//		providerNameMap[providerName] = struct{}{}
//	}
//
//	provider := registry.NewProviderGithubRegistry(x.downloadWorkspace)
//	for _, p := range cof.Selefra.Providers {
//		prov := registry.ProviderBinary{
//			Provider: &registry.Provider{
//				Name:    p.Name,
//				Version: p.Version,
//				Source:  "",
//			},
//			FilePath: p.Path,
//		}
//		if len(providerNameSlice) != 0 && !providerNameMap[p.Name] {
//			break
//		}
//
//		pp, err := provider.CheckUpdate(ctx, prov)
//		if err != nil {
//			return err
//		}
//		p.Path = pp.Filepath
//		p.Version = pp.Version
//		confs, err := tools.GetProviders(cof, p.Name)
//		if err != nil {
//			return err
//		}
//		for _, c := range confs {
//			err = fetch.Fetch(ctx, cof, p, c)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
