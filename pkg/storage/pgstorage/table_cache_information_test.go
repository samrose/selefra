package pgstorage

import (
	"context"
	"github.com/selefra/selefra-utils/pkg/id_util"
	"github.com/selefra/selefra-utils/pkg/json_util"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReadTableCacheInformation(t *testing.T) {
	testStorage := getTestStorage(t)

	tableName := "test_foo_bar"

	// save
	d := SaveTableCacheInformation(context.Background(), testStorage, &TableCacheInformation{
		TableName:    tableName,
		LastPullTime: time.Now(),
		LastPullId:   id_util.RandomId(),
	})
	if utils.HasError(d) {
		t.Log(d.String())
	}
	assert.False(t, utils.HasError(d))

	// read
	information, diagnostics := ReadTableCacheInformation(context.Background(), testStorage, tableName)
	if utils.HasError(diagnostics) {
		t.Log(diagnostics.String())
	}
	assert.False(t, utils.HasError(diagnostics))

	t.Log(json_util.ToJsonString(information))

}
