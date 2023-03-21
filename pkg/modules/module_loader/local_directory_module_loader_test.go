package module_loader

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalDirectoryModuleLoader_Load(t *testing.T) {

	source := "./test_data/module_mixed"

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})
	loader, err := NewLocalDirectoryModuleLoader(&LocalDirectoryModuleLoaderOptions{
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source:            source,
			Version:           "",
			DownloadDirectory: testDownloadDirectory,
			MessageChannel:    messageChannel,
			DependenciesTree:  []string{source},
		},
		ModuleDirectory: source,
	})
	assert.Nil(t, err)
	rootModule, isLoadSuccess := loader.Load(context.Background())
	assert.True(t, isLoadSuccess)
	assert.NotNil(t, rootModule)
	messageChannel.ReceiverWait()
}
