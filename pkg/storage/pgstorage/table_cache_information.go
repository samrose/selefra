package pgstorage

import (
	"context"
	"encoding/json"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage"
	"github.com/selefra/selefra/pkg/utils"
	"time"
)

// ------------------------------------------------- --------------------------------------------------------------------

// TableCacheInformation Some information about the table level cache
type TableCacheInformation struct {

	// Table name
	TableName string `json:"table_name"`

	// The last time this table was pulled
	LastPullTime time.Time `json:"last_pull_time"`

	// Which batch was pulled off
	LastPullId string `json:"last_pull_id"`
}

// ReadTableCacheInformation Reads the table cache information from the database
func ReadTableCacheInformation(ctx context.Context, storage storage.Storage, tableName string) (*TableCacheInformation, *schema.Diagnostics) {
	cacheKey := BuildCacheKey(tableName)
	value, diagnostics := storage.GetValue(ctx, cacheKey)
	if utils.HasError(diagnostics) {
		return nil, diagnostics
	}
	if value == "" {
		return nil, nil
	}
	information := &TableCacheInformation{}
	err := json.Unmarshal([]byte(value), information)
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg("table name = %s, read table cache information unmarshal failed: %s, s = %s", tableName, err.Error(), value)
	}
	return information, nil
}

// SaveTableCacheInformation Save the table cache information to the kv database
func SaveTableCacheInformation(ctx context.Context, storage storage.Storage, information *TableCacheInformation) *schema.Diagnostics {
	cacheKey := BuildCacheKey(information.TableName)
	marshal, err := json.Marshal(information)
	if err != nil {
		return schema.NewDiagnostics().AddErrorMsg("failed to marshal cache information: %s", err.Error())
	}
	return storage.SetKey(ctx, cacheKey, string(marshal))
}

func BuildCacheKey(tableName string) string {
	return "cache:table:pull:" + tableName
}

// ------------------------------------------------- --------------------------------------------------------------------
