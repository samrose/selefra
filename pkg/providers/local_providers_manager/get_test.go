package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalProvidersManager_Get(t *testing.T) {

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
			RequiredProvider: NewLocalProvider("aws", "v0.0.5"),
			MessageChannel:   messageChannel,
		})
		messageChannel.ReceiverWait()
	}

	localProvider, diagnostics := manager.Get(context.Background(), NewLocalProvider("aws", "v0.0.5"))
	if utils.IsNotEmpty(diagnostics) {
		t.Log(diagnostics.ToString())
	}
	assert.False(t, utils.HasError(diagnostics))
	assert.Nil(t, localProvider)
}
