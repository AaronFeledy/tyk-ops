package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish API definitions from a Git repo or file system to a gateway or dashboard",
	Long: `Publish API definitions from a Git repo to a gateway or dashboard, this
	will not update existing APIs, and if it detects a collision, will stop.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cfg.TargetEnv != nil {
			url := cfg.TargetEnv.Dashboard.Url
			secret := cfg.TargetEnv.Dashboard.Secret
			urlFlag := "dashboard"
			serverType := viper.GetString("target-server.type")
			if serverType == "gateway" {
				url = cfg.TargetEnv.Gateway.Url
				secret = cfg.TargetEnv.Gateway.Secret
				urlFlag = "gateway"
			}
			if val, _ := cmd.Flags().GetString(urlFlag); val == "" {
				cmd.Flags().Lookup(urlFlag).Value.Set(url)
			}
			if val, _ := cmd.Flags().GetString("secret"); val == "" {
				cmd.Flags().Lookup("secret").Value.Set(secret)
			}
		}
		verificationError := verifyArguments(cmd)
		if verificationError != nil {
			fmt.Println(verificationError)
			os.Exit(1)
		}

		err := processPublish(cmd, args)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.
	publishCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	publishCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	publishCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	publishCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	publishCmd.Flags().StringP("secret", "s", "", "Your API secret")
	publishCmd.Flags().StringP("path", "p", "", "Source directory for definition files (optional)")
	publishCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
	publishCmd.Flags().StringSlice("policies", []string{}, "Specific Policies ids to publish")
	publishCmd.Flags().StringSlice("apis", []string{}, "Specific Apis ids to publish")
	publishCmd.Flags().BoolP("skip-existing", "n", false, "Skip creating APIs if they already exist")
	publishCmd.Flags().BoolP("insecure", "", false, "Override TLS certificate validation")
}
