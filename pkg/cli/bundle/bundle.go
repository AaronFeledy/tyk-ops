// Package bundle contains the bundle subcommand
package bundle

import (
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/cli_util"
	"github.com/spf13/cobra"
	"strings"
)

const cmdName = "bundle"

var (
	rootCmd *cobra.Command
	cfg     = cli_util.Config
)

// BundleCmd is the root command for the bundle subcommand
var BundleCmd = &cobra.Command{
	Use:   cmdName + " [command]",
	Short: "Manage bundle resources",
	Long:  `Manage bundle resources`,
	RunE:  bundleRun,
}

// bundleRun is the function that is run when the bundle command is executed
func bundleRun(cmd *cobra.Command, args []string) error {
	fmt.Println("bundleRun")
	_ = cmd.Help()
	return nil
}

func init() {
	rootCmd = BundleCmd.Root()
	BundleCmd.Example = strings.Join([]string{rootCmd.Name(), BundleCmd.Name()}, " ")

	BundleCmd.AddCommand(pushCmd)
}
