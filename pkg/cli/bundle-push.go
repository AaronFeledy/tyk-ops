package cli

import (
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/clients/mserv"
	out "github.com/AaronFeledy/tyk-ops/pkg/output"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"path"
	"strings"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:     "bundle:push bundle_zip",
	Short:   "Pushes a middleware bundle to mserv",
	Long:    "Uploads a bundle file created with tyk CLI to mserv",
	Example: rootCmd.Use + "@dev bundle:push /path/to/bundle.zip",
	Args:    cobra.ExactArgs(1),
	Run:     pushBundle,
}

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().BoolP("storeonly", "s", false, "Don't process, just store it")
	pushCmd.Flags().StringP("apiid", "a", "", "Optional API ID")
	pushCmd.PersistentFlags().StringP("token", "t", "", "Mserv security token")
	pushCmd.PersistentFlags().StringP("endpoint", "e", "", "Mserv endpoint")
	pushCmd.PersistentFlags().StringP("dashboard", "d", "", "The dashboard proxying to mserv")
	pushCmd.PersistentFlags().BoolP("insecure-tls", "k", false, "allow insecure TLS for mserv client")

	_ = viper.BindPFlag("mserv-url", pushCmd.PersistentFlags().Lookup("endpoint"))
	_ = viper.BindPFlag("mserv-secret", pushCmd.PersistentFlags().Lookup("token"))

}

func pushBundle(cmd *cobra.Command, args []string) {
	if Cfg.TargetEnv != nil {
		viper.SetDefault("mserv-url", Cfg.TargetEnv.Mserv.Url)
		viper.SetDefault("mserv-secret", Cfg.TargetEnv.Mserv.Secret)
	}

	file, err := os.Open(args[0])
	if err != nil {
		cmd.PrintErr(fmt.Sprintln("Couldn't open the bundle file"))
		cmd.PrintErr(fmt.Sprintln(err.Error()))
		os.Exit(1)
	}

	endpoint := viper.GetString("mserv-url")
	if endpoint == "" {
		dbUrl, _ := cmd.Flags().GetString("dashboard")
		if dbUrl == "" {
			cmd.PrintErr(fmt.Sprintln(cmd.Use, "requires an endpoint or dashboard URL to be set"))
			os.Exit(1)
		} else {
			urlObj, err := url.Parse(dbUrl)
			if err != nil {
				cmd.PrintErr(fmt.Sprintln("Couldn't parse dashboard URL"))
				os.Exit(1)
			}
			urlObj.Path = path.Join(urlObj.Path, "mserv")
			endpoint = urlObj.String()
		}
	}

	secret := viper.GetString("mserv-secret")

	if secret == "" {
		cmd.PrintErr(fmt.Sprintln("Please set the --token flag to your mserv access token"))
		os.Exit(1)
	}

	client, err := mserv.NewMservClient(endpoint, secret)
	if err != nil {
		cmd.PrintErr(fmt.Sprintln(err.Error()))
		os.Exit(1)
	}

	apiid, _ := cmd.Flags().GetString("apiid")
	storeOnly, _ := cmd.Flags().GetBool("storeonly")

	fileParts := strings.Split(file.Name(), ".")
	ext := fileParts[len(fileParts)-1]
	if ext != "zip" {
		cmd.PrintErr(fmt.Sprintf("File type must be zip: %v\n", ext))
		os.Exit(1)
	}

	cmd.PrintErrln("")
	fileInfo, err := file.Stat()
	bar := progressbar.DefaultBytes(
		fileInfo.Size(),
		"Uploading bundle:",
	)

	params := &mserv.BundlePushParams{
		APIID:          &apiid,
		StoreOnly:      &storeOnly,
		UploadFile:     file,
		ProgressWriter: bar,
	}

	data, err := client.BundlePush(params)
	if err != nil {
		_ = bar.Exit()
		cmd.PrintErr(fmt.Sprintln("Failed to push bundle to mserv"))
		cmd.PrintErr(fmt.Sprintln(err.Error()))
		os.Exit(1)
	}
	err = bar.Finish()
	err = bar.Clear()
	if err != nil {
		cmd.PrintErr(fmt.Sprintln(err.Error()))
		os.Exit(1)
	}

	successMsg := "Bundle successfully pushed to mserv with ID: "
	out.DataWithFlair(data.Id).Pre(successMsg).Println()
}
