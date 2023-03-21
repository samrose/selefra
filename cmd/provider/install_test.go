package provider

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

//import (
//	"context"
//	"github.com/selefra/selefra/global"
//	"github.com/spf13/cobra"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func TestInstall(t *testing.T) {
//	global.Init("TestInstall", global.WithWorkspace("../../tests/workspace/offline"))
//
//	ctx := context.Background()
//	err := install(ctx, []string{"aws@latest"})
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestInstallCmd(t *testing.T) {
//	rootCmd := &cobra.Command{
//		Use: "provider",
//	}
//	installCmd := newCmdProviderInstall()
//	rootCmd.AddCommand(installCmd)
//
//	require.Equal(t, "provider install", global.Cmd())
//}

func Test_install(t *testing.T) {
	err := Install(context.Background(), "./test_download", "mock")
	assert.Nil(t, err)
}
