package pgstorage

import (
	"context"
	"encoding/json"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage"
	"github.com/selefra/selefra/pkg/utils"
)

// ------------------------------------------------- --------------------------------------------------------------------

const SchemaOwnerKey = "schema-owner-key"

// SchemaOwnerInformation schema is held by whom, information about the holder
type SchemaOwnerInformation struct {

	// The host name of the holder
	Hostname string `json:"hostname"`

	// Holder's ID
	HolderID string `json:"holder_id"`

	// The name of the holder's configuration file
	ConfigurationName string `json:"configuration_name"`

	// MD5 configured when pulling data from this schema
	ConfigurationMD5 string `json:"configuration_md5"`
}

func GetSchemaOwner(ctx context.Context, storage storage.Storage) (*SchemaOwnerInformation, *schema.Diagnostics) {
	value, diagnostics := storage.GetValue(ctx, SchemaOwnerKey)
	if utils.HasError(diagnostics) {
		return nil, diagnostics
	}
	if value == "" {
		return nil, nil
	}
	ownerInformation := &SchemaOwnerInformation{}
	err := json.Unmarshal([]byte(value), &ownerInformation)
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg("failed to unmarshal schema owner: %s, s = %s", err.Error(), value)
	}
	return ownerInformation, nil
}

func SaveSchemaOwner(ctx context.Context, storage storage.Storage, owner *SchemaOwnerInformation) *schema.Diagnostics {
	marshal, err := json.Marshal(owner)
	if err != nil {
		return schema.NewDiagnostics().AddErrorMsg("failed to marshal schema owner: %s", err.Error())
	}
	return storage.SetKey(ctx, SchemaOwnerKey, string(marshal))
}

// ------------------------------------------------- --------------------------------------------------------------------
