package cmd

import (
	"github.com/spf13/cobra"
)

func init() {

}

var RootCmd = &cobra.Command{
	Use:   "tykops",
	Short: "A tool to manage Tyk environments",
	Long: `A tool to manage syncing and deployments of Tyk Gateways and their
           middleware bundles.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
