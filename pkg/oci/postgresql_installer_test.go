package oci

import (
	"context"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Failed to pass the test on Windows. Procedure
func TestPostgreSQLInstaller_Run1(t *testing.T) {

	//downloadWorkspace := "./test_download"
	downloadWorkspace, err := config.GetDefaultDownloadCacheDirectory()
	assert.Nil(t, err)
	err = utils.EnsureDirectoryNotExists(downloadWorkspace)
	assert.Nil(t, err)

	messageChannel := message.NewChannel[*schema.Diagnostics](func(index int, message *schema.Diagnostics) {
		if utils.IsNotEmpty(message) {
			fmt.Println(message.ToString())
		}
	})
	downloader := NewPostgreSQLDownloader(&PostgreSQLDownloaderOptions{
		MessageChannel:    messageChannel,
		DownloadDirectory: downloadWorkspace,
	})
	isRunSuccess := downloader.Run(context.Background())
	messageChannel.ReceiverWait()
	assert.True(t, isRunSuccess)
}
