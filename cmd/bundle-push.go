package cmd

import (
	"fmt"
	"github.com/AaronFeledy/tyk-ops/clients/mserv"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"path"
	"strings"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:     "bundle:push",
	Short:   "Pushes a middleware bundle to mserv",
	Long:    "Uploads a bundle file created with tyk CLI to mserv",
	Example: RootCmd.Use + " bundle:push /path/to/bundle.zip",
	Args:    cobra.ExactArgs(1),
	Run:     pushBundle,
}

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().BoolP("storeonly", "s", false, "Don't process, just store it")
	pushCmd.Flags().StringP("apiid", "a", "", "Optional API ID")
	pushCmd.PersistentFlags().StringP("token", "t", "", "Mserv security token")
	pushCmd.PersistentFlags().StringP("endpoint", "e", "", "Mserv endpoint")
	pushCmd.PersistentFlags().StringP("dashboard", "d", "", "The dashboard proxying to mserv")
	pushCmd.PersistentFlags().BoolP("insecure-tls", "k", false, "allow insecure TLS for mserv client")
}

func pushBundle(cmd *cobra.Command, args []string) {
	verificationError := verifyArguments(cmd)
	if verificationError != nil {
		cmd.PrintErr(fmt.Sprintln(verificationError))
		os.Exit(1)
	}

	file, err := os.Open(args[0])
	if err != nil {
		cmd.PrintErr(fmt.Sprintln("Couldn't open the bundle file"))
		cmd.PrintErr(fmt.Sprintln(err.Error()))
		os.Exit(1)
	}

	endpoint, _ := cmd.Flags().GetString("endpoint")
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

	secret, _ := cmd.Flags().GetString("secret")
	token, _ := cmd.Flags().GetString("token")

	if secret == "" && token == "" {
		cmd.PrintErr(fmt.Sprintln("Please set the --secret or --token flag to your mserv access token"))
		os.Exit(1)
	}

	if secret != "" {
		token = secret
	}

	client, err := mserv.NewMservClient(endpoint, token)
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

	// Print only the bundle ID to stdout so it can be piped to other commands without including the surrounding text.
	successMsg := "Bundle successfully pushed to mserv with ID: "
	cmd.PrintErr(fmt.Sprintf("%v", successMsg))
	cmd.Printf("%v", data.Id)
	// Rewrite the previous line to stderr so that it is also displayed to the user.
	cmd.PrintErr(fmt.Sprintf("\r%v%v\n", successMsg, data.Id))

}
