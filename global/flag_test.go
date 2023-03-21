package global

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Variable(t *testing.T) {
	Init("cmd", WithWorkspace("/path/to/workspace"))

	require.Equal(t, "cmd", Cmd())

	require.Equal(t, "/path/to/workspace", WorkSpace())

	// Init only do once
	Init("cmd1", WithWorkspace("/fake/workspace"))

	require.Equal(t, "cmd", Cmd())

	require.Equal(t, "/path/to/workspace", WorkSpace())
}

func Test_parentCmdNames(t *testing.T) {
	cmd_a := &cobra.Command{
		Use: "a",
	}

	cmd_b := &cobra.Command{
		Use: "b",
	}

	//cmd_c := &cobra.Command{
	//	Use: "c",
	//}

	cmd_a.AddCommand(cmd_b)

	require.Equal(t, []string{"a", "b"}, parentCmdNames(cmd_b))
}
