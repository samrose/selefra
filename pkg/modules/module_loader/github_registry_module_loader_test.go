package module_loader

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testDownloadDirectory = "./test_download"

func TestGitHubRegistryModuleLoader_Load(t *testing.T) {

	source := "rules-aws-misconfiguration-s3@v0.0.4"

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			t.Log(message.ToString())
		}
	})

	loader, err := NewGitHubRegistryModuleLoader(&GitHubRegistryModuleLoaderOptions{
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source:            source,
			Version:           "",
			MessageChannel:    messageChannel,
			DownloadDirectory: testDownloadDirectory,
			DependenciesTree:  []string{source},
		},
		RegistryRepoFullName: "selefra/registry",
	})
	assert.Nil(t, err)
	rootModule, b := loader.Load(context.Background())
	messageChannel.ReceiverWait()
	assert.True(t, b)
	assert.NotNil(t, rootModule)

}
