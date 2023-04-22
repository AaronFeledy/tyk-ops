package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "TykOps version",
	Long:    `This command will show the current TykOps version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v" + VERSION)
	},
}
