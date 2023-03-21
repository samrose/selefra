package pgstorage

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/env"
	"github.com/selefra/selefra-provider-sdk/storage"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra-provider-sdk/storage_factory"
	"github.com/selefra/selefra-utils/pkg/json_util"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getTestStorage(t *testing.T) storage.Storage {
	options := postgresql_storage.NewPostgresqlStorageOptions(env.GetDatabaseDsn())
	storage, diagnostics := storage_factory.NewStorage(context.Background(), storage_factory.StorageTypePostgresql, options)
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.String())
	}
	assert.False(t, utils.HasError(diagnostics))
	return storage
}

func TestSaveSchemaOwner(t *testing.T) {
	hostname, _ := os.Hostname()
	storage := getTestStorage(t)
	information := &SchemaOwnerInformation{
		Hostname:          hostname,
		HolderID:          "test",
		ConfigurationName: "",
		ConfigurationMD5:  "",
	}
	d := SaveSchemaOwner(context.Background(), storage, information)
	if utils.IsNotEmpty(d) {
		t.Log(d.String())
	}
	assert.False(t, utils.HasError(d))

	owner, diagnostics := GetSchemaOwner(context.Background(), storage)
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.String())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.NotNil(t, owner)
	t.Log(json_util.ToJsonString(owner))
}
