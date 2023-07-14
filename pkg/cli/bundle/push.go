package bundle

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
	Use:     "push bundle_zip",
	Short:   "Pushes a middleware bundle to mserv",
	Long:    "Uploads a bundle file created with tyk CLI to mserv",
	Args:    cobra.ExactArgs(1),
	RunE:    pushBundle,
	Example: "/path/to/bundle.zip",
}

func init() {
	pushCmd.Example = strings.Join([]string{
		rootCmd.Name(), "@dev", BundleCmd.Name(), pushCmd.Name(), " /path/to/bundle.zip",
	}, " ")

	pushCmd.Flags().BoolP("storeonly", "s", false, "Don't process, just store it")
	pushCmd.Flags().StringP("apiid", "a", "", "Optional API ID")
	pushCmd.PersistentFlags().StringP("token", "t", "", "Mserv security token")
	pushCmd.PersistentFlags().StringP("endpoint", "e", "", "Mserv endpoint")
	pushCmd.PersistentFlags().StringP("dashboard", "d", "", "The dashboard proxying to mserv")
	pushCmd.PersistentFlags().BoolP("insecure-tls", "k", false, "allow insecure TLS for mserv client")

	_ = viper.BindPFlag("mserv-url", pushCmd.PersistentFlags().Lookup("endpoint"))
	_ = viper.BindPFlag("mserv-secret", pushCmd.PersistentFlags().Lookup("token"))

}

// pushBundle is a function which implements the `tykops bundle:push` CLI command to handle uploading a bundle to mserv.
func pushBundle(cmd *cobra.Command, args []string) error {
	if cfg.TargetEnv != nil {
		viper.SetDefault("mserv-url", cfg.TargetEnv.Mserv.Url)
		viper.SetDefault("mserv-secret", cfg.TargetEnv.Mserv.Secret)
	}

	fileName := args[0]

	fileParts := strings.Split(fileName, ".")
	ext := fileParts[len(fileParts)-1]
	if ext != "zip" {
		return fmt.Errorf("file type must be zip: '%v' given", ext)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("couldn't read the bundle file at '%s': %s", fileName, err.Error())
	}

	endpoint := viper.GetString("mserv-url")
	if endpoint == "" {
		dbUrl, _ := cmd.Flags().GetString("dashboard")
		if dbUrl == "" {
			return fmt.Errorf("'%s' requires an endpoint or dashboard URL to be set", cmd.Use)
		} else {
			urlObj, err := url.Parse(dbUrl)
			if err != nil {
				return fmt.Errorf("couldn't parse dashboard URL")
			}
			urlObj.Path = path.Join(urlObj.Path, "mserv")
			endpoint = urlObj.String()
		}
	}

	secret := viper.GetString("mserv-secret")

	if secret == "" {
		return fmt.Errorf("please set the --token flag to your mserv access token")
	}

	client, err := mserv.NewMservClient(endpoint, secret)
	if err != nil {
		return fmt.Errorf("failed to init mserv client: %s", err.Error())
	}

	apiid, _ := cmd.Flags().GetString("apiid")
	storeOnly, _ := cmd.Flags().GetBool("storeonly")

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
		return err
	}
	err = bar.Finish()
	err = bar.Clear()
	if err != nil {
		_ = bar.Exit()
		cmd.PrintErrln()
	}

	successMsg := "Bundle successfully pushed to mserv with ID: "
	out.DataWithFlair(data.Id).Pre(successMsg).Println()

	return nil
}
