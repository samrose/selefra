package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

//import (
//	"github.com/selefra/selefra/global"
//	"testing"
//)
//
//func TestList(t *testing.T) {
//	*global.WORKSPACE = "../../tests/workspace/offline"
//	err := list()
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestLocalProvidersManager_ListProviderVersions(t *testing.T) {

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

	versions, diagnostics := getTestLocalProviderManager().ListProviderVersions(testProviderName)
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.GreaterOrEqual(t, 1, len(versions.ProviderVersionMap))
}

func TestLocalProvidersManager_ListProviders(t *testing.T) {
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

	providers, diagnostics := getTestLocalProviderManager().ListProviders()
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.GreaterOrEqual(t, 1, len(providers))
}
