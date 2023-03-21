package version

import (
	"fmt"
	"github.com/selefra/selefra/global"
	"github.com/spf13/cobra"
)

var Version = "{{version}}"

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "version",
		Short:            "Print Selefra's version number",
		Long:             "Print Selefra's version number",
		PersistentPreRun: global.DefaultWrappedInit(),
		Run: func(cmd *cobra.Command, args []string) {
			version()
		},
	}
	return cmd
}

func version() {
	fmt.Println(Version)
}
