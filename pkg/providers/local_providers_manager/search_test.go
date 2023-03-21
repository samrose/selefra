package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalProvidersManager_SearchLocal(t *testing.T) {

	testProviderName := "aws"
	testProviderVersion := "v0.0.5"

	manager := getTestLocalProviderManager()

	isInstalled, d := manager.IsProviderInstalled(context.Background(), NewLocalProvider(testProviderName, testProviderVersion))
	if utils.IsNotEmpty(d) {
		t.Log(d.ToString())
	}
	assert.False(t, utils.HasError(d))
	if !isInstalled {
		messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
			t.Log(message.ToString())
		})
		manager.InstallProvider(context.Background(), &InstallProvidersOptions{
			RequiredProvider: NewLocalProvider(testProviderName, testProviderVersion),
			MessageChannel:   messageChannel,
		})
		messageChannel.ReceiverWait()
	}

	hitProviders, diagnostics := manager.SearchLocal(context.Background(), testProviderName)
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.Len(t, hitProviders, 1)
	isContains := false
	for _, provider := range hitProviders {
		if provider.ProviderName == testProviderName {
			isContains = true
		}
	}
	assert.True(t, isContains)
}

func TestLocalProvidersManager_SearchRegistry(t *testing.T) {

	testProviderName := "aws"

	hitProviders, diagnostics := getTestLocalProviderManager().SearchRegistry(context.Background(), testProviderName)
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.Len(t, hitProviders, 1)
	isContains := false
	for _, provider := range hitProviders {
		if provider.Name == testProviderName {
			isContains = true
		}
	}
	assert.True(t, isContains)
}
