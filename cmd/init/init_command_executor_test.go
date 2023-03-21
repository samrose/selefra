package init

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInitCommandExecutor_Run(t *testing.T) {

	_ = os.Setenv(SelefraInputInitForceConfirm, "y")

	err := NewInitCommandExecutor(&InitCommandExecutorOptions{
		DownloadWorkspace: "./test_download",
		ProjectWorkspace:  "./test_data",
		IsForceInit:       true,
		RelevanceProject:  "",
		DSN:               "",
	}).Run(context.Background())
	assert.NotNil(t, err)

}
