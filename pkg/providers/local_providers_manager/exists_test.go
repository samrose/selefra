package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getTestLocalProviderManager() *LocalProvidersManager {
	manager, err := NewLocalProvidersManager("./test_download")
	if err != nil {
		panic(err)
	}
	return manager
}

func TestLocalProvidersManager_IsProviderInstalled(t *testing.T) {
	installed, diagnostics := getTestLocalProviderManager().IsProviderInstalled(context.Background(), NewLocalProvider("aws", "v0.0.5"))
	assert.False(t, utils.HasError(diagnostics))
	assert.False(t, installed)
}
