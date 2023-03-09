package cli

import (
	"github.com/AaronFeledy/tyk-ops/pkg/ops"
	out "github.com/AaronFeledy/tyk-ops/pkg/output"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// loginCmd defines the `tykops login` CLI command
var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Log in to the Tyk Dashboard",
	Long:    "Generate a login link for the Tyk Dashboard and open it in your default browser.",
	Example: RootCmd.Use + " login",
	Args:    cobra.NoArgs,
	Run:     cmdLogin,
}

// loginOpt defines the flags for the `tykops login` CLI command
func loginOpt() {
	loginCmd.Flags().BoolP("insecure", "k", false, "Override TLS certificate validation")
	loginCmd.Flags().StringP("org", "o", "", "The ID of the organization to log in to")
	loginCmd.Flags().StringP("user", "u", "", "The email address of the user to log in as")

	// It's safe to use the default environment as the target for this command.
	viper.SetDefault("target", "default")
}

// cmdLogin is a function which implements the `tykops login` CLI command
func cmdLogin(cmd *cobra.Command, args []string) {
	allowInsecure := false
	if allowInsecure, _ = cmd.Flags().GetBool("insecure"); !allowInsecure {
		allowInsecure = Cfg.TargetEnv.Dashboard.AllowInsecure
	}

	dashAdmin := ops.DashboardAdmin{
		Server: ops.Server{
			Type:          "dashboard",
			Url:           Cfg.TargetEnv.Dashboard.Url,
			Secret:        Cfg.TargetEnv.Dashboard.Secret,
			AllowInsecure: allowInsecure,
		},
		Client: resty.New(),
	}

	orgId := ""
	if orgId, _ = cmd.Flags().GetString("org"); orgId == "" {
		orgs, err := dashAdmin.GetOrganizations()
		if err != nil {
			out.User.Error(err.Error())
			os.Exit(1)
		}
		if len(*orgs) == 0 {
			out.User.Error("no organizations found")
			os.Exit(1)
		}
		if len(*orgs) > 1 {
			out.User.Error("multiple organizations found, please specify one with the -o flag")
			os.Exit(1)
		}
		orgId = (*orgs)[0].Id
	}

	sso, err := dashAdmin.SSO("dashboard", orgId, "", "")
	if err != nil {
		out.User.Error(err.Error())
		os.Exit(1)
	}
	out.User.Info("Login link: " + sso)
}

// init registers the `tykops login` CLI command
func init() {
	loginOpt()
	RootCmd.AddCommand(loginCmd)
}
