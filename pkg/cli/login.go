package cli

import (
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/ops"
	out "github.com/AaronFeledy/tyk-ops/pkg/output"
	"github.com/containerd/console"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var loginTriedBrowser bool

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
	loginCmd.Flags().StringP("secret", "s", "", "The dashboard admin auth token to use")
	loginCmd.Flags().BoolP("no-browser", "n", false, "Don't open the login link in a browser")
	loginCmd.Flags().BoolP("interactive", "i", false, "Prompt for details. This is the default in TTY mode.")

	// It's safe to use the default environment as the target for this command.
	viper.SetDefault("target", "default")
}

// cmdLogin is a function which implements the `tykops login` CLI command
func cmdLogin(cmd *cobra.Command, args []string) {
	// A target is required
	if Cfg.TargetEnv == nil {
		out.User.Fatal("No target environment specified")
	}
	// Update the target environment config with the flags
	if secret, _ := cmd.Flags().GetString("secret"); secret != "" {
		Cfg.TargetEnv.Dashboard.Secret = secret
	}

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

	loginLink, err := dashAdmin.SSO("dashboard", orgId, "", "")
	if err != nil {
		out.User.Error(err.Error())
		os.Exit(1)
	}

	// Print the link
	out.DataWithFlair(loginLink).
		Pre(labelColor("You may now log in to Tyk using the following link:\n")).
		Print()

	// We only want the interactive bits to happen when we are in a terminal.
	// Scripts should just dump the link and move on.
	interactive, _ := cmd.Flags().GetBool("interactive")
	if !interactive && !isatty.IsTerminal(os.Stdout.Fd()) {
		out.User.Printf("\n")
		return
	} else {
		// Block input from echoing to the terminal during the countdown
		current := console.Current()
		defer func(current console.Console) {
			err := current.Reset()
			if err != nil {
				out.User.Debug(err.Error())
			}
		}(current)
		err := current.DisableEcho()
		if err != nil {
			out.User.Debug(err.Error())
		}

		// Capture keystrokes so we can exit countdown on key press
		input := make(chan rune, 1)
		go readKey(input)

		wrapCount := 0
		trailingSpace := 0
		padding := " "
		cliClearLine := "\033[2K\r"
		// Add a countdown to link expiry
	CountDown:
		for i := 59; i >= 0; i-- {
			expires := fmt.Sprintf("[Expires in %02ds]", i)

			// First run of the loop, print the link
			if i == 59 {
				// Determine whether the link will wrap to the next line
				termSize, err := current.Size()
				if err != nil {
					out.User.Debug(err.Error())
				}
				wrapCount = len(loginLink) / int(termSize.Width)

				// Clear the previous output
				for w := 0; w < wrapCount; w++ {
					out.User.Printf("%s\033[1A", cliClearLine)
				}

				// Print the link
				out.User.Printf("%s%s", cliClearLine, loginLink)

				// If the expires string is longer than the terminal width, print a newline so it doesn't wrap
				trailingSpace = int(termSize.Width) - (len(loginLink) % int(termSize.Width))
				if len(expires) >= trailingSpace {
					out.User.Printf("\n")
					padding = ""
					wrapCount++
				}
			}
			if i == 0 {
				expires = fmt.Sprintf("[ LINK EXPIRED ]")
				out.User.Printf("%s", cliClearLine)
				// Clear the number of lines equal to wrapCount
				for w := 0; w <= wrapCount; w++ {
					out.User.Printf("\033[1A%s", cliClearLine)
				}

				// Print the expired link in strikethrough
				out.User.Printf("\033[9m%s\033[0m", loginLink)

				if len(expires) >= trailingSpace {
					out.User.Printf("\n")
				}
			}
			expires = padding + expires

			var expiresColored string
			switch {
			case i == 0:
				expiresColored = color.HiRedString(expires)
			case i > 30:
				expiresColored = color.GreenString(expires)
			case i > 10:
				expiresColored = color.YellowString(expires)
			default:
				expiresColored = color.RedString(expires)
			}
			out.User.Printf("%s", expiresColored)

			select {
			case <-time.After(time.Second * time.Duration(i*2000)): // Immediately break when i==0
				out.User.Printf("\n")
				break CountDown
			case opened := <-openLink(cmd, loginLink): // Try to open the link in the browser automatically
				if opened {
					// Back up the cursor to the beginning of expiry countdown
					out.User.Printf("\033[%dD", len(expires))
					// Clear everything after the cursor
					out.User.Printf("\033[J")
					// Add a final line break if we don't already have one
					if len(expires) < trailingSpace {
						out.User.Printf("\n")
					}
					break CountDown
				}
			case <-time.After(time.Second): // Update every second
				// Back up the cursor to the beginning of expiry countdown
				out.User.Printf("\033[%dD", len(expires))
				continue
			case <-input: // End countdown when a key is pressed
				// Back up the cursor to the beginning of expiry countdown
				out.User.Printf("\033[%dD", len(expires))
				// Clear everything after the cursor
				out.User.Printf("\033[J")
				// Add a final line break if we don't already have one
				if len(expires) < trailingSpace {
					out.User.Printf("\n")
				}
				break CountDown
			}
		}
	}
}

// readKey opens a channel that sends any key when it is pressed
func readKey(input chan rune) {
	defer close(input)
	err := keyboard.Open()
	if err != nil {
		out.User.Fatal(err)
	}
	defer keyboard.Close()

	for {
		char, _, err := keyboard.GetSingleKey()
		if err != nil {
			out.User.Fatal(err)
		}
		input <- char
	}
}

// openLink opens the specified link in the default web browser of the user's operating system,
// unless the "no-browser" flag is set to true, and returns a channel that receives a boolean value
// when the browser is opened or an error occurs. Only one attempt to open the link is made, and subsequent
// calls to this function will return a nil channel.
func openLink(cmd *cobra.Command, link string) chan bool {
	c := make(chan bool, 1)
	// Only try this once
	if loginTriedBrowser {
		return nil
	}
	loginTriedBrowser = true

	// Don't open a link if this is an ssh session
	if os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_TTY") != "" {
		return nil
	}

	// User can specify not to open the link in the browser
	if noBrowser, _ := cmd.Flags().GetBool("no-browser"); noBrowser {
		return nil
	}
	var execCmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		execCmd = exec.Command("cmd", "/C", "start", link)
	case "darwin":
		execCmd = exec.Command("open", link)
	default: // assume Linux or other Unix-like OS
		execCmd = exec.Command("xdg-open", link)
	}
	go func() {
		err := execCmd.Start()
		if err != nil {
			out.User.Debug("error opening link: ", err)
			return
		}
		c <- true
	}()
	return c
}

// init registers the `tykops login` CLI command
func init() {
	loginOpt()
	RootCmd.AddCommand(loginCmd)
}
