package cli_ui

//import (
//	"context"
//	"errors"
//	"github.com/google/uuid"
//	"github.com/selefra/selefra-provider-sdk/storage"
//	"github.com/selefra/selefra/config"
//	"github.com/selefra/selefra/pkg/registry"
//	"github.com/selefra/selefra/pkg/storage/pgstorage"
//	"github.com/selefra/selefra/ui"
//)
//
//type Client struct {
//	Registry      interface{}
//	PluginManager interface{}
//	Storage       storage.Storage
//	instanceId    uuid.UUID
//}
//
//func CreateClientFromConfig(ctx context.Context, cfg *config.SelefraConfig, instanceId uuid.UUID, provider *config.ProviderDecl, cp config.Provider) (*Client, error) {
//
//	hub := new(interface{})
//	pm := new(interface{})
//
//	c := &Client{
//		Storage:       nil,
//		cfg:           cfg,
//		Registry:      hub,
//		PluginManager: pm,
//		instanceId:    instanceId,
//	}
//
//	schema := config.GetSchemaKey(provider, cp)
//	sto, diag := pgstorage.Storage(ctx, pgstorage.WithSearchPath(schema))
//	if diag != nil {
//		err := ui.PrintDiagnostic(diag.GetDiagnosticSlice())
//		if err != nil {
//			return nil, errors.New("failed to create pgstorage")
//		}
//	}
//	if sto != nil {
//		c.Storage = sto
//	}
//
//	c.Providers = registry.Providers{}
//	for _, rp := range cfg.RequireProvidersBlock {
//		c.Providers.Set(registry.Provider{Name: rp.Name, Version: rp.Version})
//	}
//
//	return c, nil
//}
