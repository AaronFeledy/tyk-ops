package cli

import (
	"github.com/AaronFeledy/tyk-ops/pkg/ops"
	out "github.com/AaronFeledy/tyk-ops/pkg/output"
	"github.com/fatih/color"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:  "tykops",
	Long: "A tool to manage syncing and deployments of Tyk Gateways and their middleware bundles.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			out.User.Errorln(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd.SetOut(os.Stdout)

	cc.Init(&cc.Config{
		RootCmd:       rootCmd,
		Headings:      cc.HiCyan + cc.Bold + cc.Underline,
		Commands:      cc.HiYellow + cc.Bold,
		CmdShortDescr: cc.HiBlue,
		Example:       cc.HiGreen + cc.Italic,
		ExecName:      cc.Bold,
		Flags:         cc.HiMagenta + cc.Bold,
		FlagsDescr:    cc.HiBlue,
	})

	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErrln(color.RedString(err.Error()))
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to config file")
	rootCmd.PersistentFlags().String("target", "", "A target environment to use as defined in your configuration file. You may use @target_name as shorthand for this flag.")

	_ = viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))

	// Support @ shorthand for target flag in commands (e.g. tykops @dev deploy)
	args := os.Args
	for i, arg := range args {
		if arg[:1] == "@" {
			segments := strings.Split(arg[1:], ".")
			switch len(segments) {
			case 2:
				// One segment after the first dot is the server type
				viper.Set("target-server.type", segments[1])
				args[i] = "--target=" + segments[0]
			case 3:
				// Two segments after the first dot is the server type and name
				viper.Set("target-server.name", segments[2])
				viper.Set("target-server.type", segments[1])
				args[i] = "--target=" + segments[0]
			case 4:
				// Three segments after the first dot
				targetServer := arg[strings.Index(arg, ".")+1:]
				viper.Set("target-server.type", strings.Split(targetServer, ".")[0])
				if len(strings.Split(targetServer, ".")) > 1 {
					viper.Set("target-server.name", strings.Join(strings.Split(targetServer, ".")[1:], "."))
				}
				arg = arg[:strings.Index(arg, ".")]
				args[i] = "--target=" + arg[1:] + ":" + strings.Split(arg, ".")[3]
			default:
				// Target is an environment name
				args[i] = "--target=" + arg[1:]
			}
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Config file has name ".tykops.yml"
		viper.SetConfigName(".tykops")

		// Walk up the directory tree looking for a config file
		wd, err := os.Getwd()
		if err != nil {
			out.User.Errorln("Couldn't get current working directory")
			os.Exit(1)
		}
		for wd != "/" {
			viper.AddConfigPath(wd)
			wd = path.Dir(wd)
		}

		// Config file may also be in the home directory
		home, err := homedir.Dir()
		if err != nil {
			out.User.Errorln(err)
		}
		viper.AddConfigPath(home)
	}

	viper.SetEnvPrefix("tykops")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		out.User.Debugln("Using config file:", viper.ConfigFileUsed())

		// Check for target environment in config
		var targetEnv bool
		target := viper.GetString("target")

		// A subcommand will set target to "default" when it wants to use the default environment from the config file
		if target == "default" {
			if viper.IsSet("environment_default") {
				target = viper.GetString("environment_default")
			}
		}

		// Don't continue if the target environment is not found in the config
		if viper.IsSet("environments." + target) {
			out.User.Debugln("Using target environment '" + target + "'")
			targetEnv = true
		} else {
			out.User.Errorln("Target environment '" + target + "' not found in " + viper.ConfigFileUsed())
			os.Exit(1)
		}

		// Load the config into the global variable
		Cfg.Environments = &ops.Environments
		if err := viper.Unmarshal(&Cfg); err != nil {
			out.User.Errorln(err.Error())
			return
		}
		// Add shorthand for the target environment
		if targetEnv {
			Cfg.TargetEnv = ops.Environments[target]
		}
	}
}

// Config is the configuration for the TykOps client.
type Config struct {
	// Environments is a map of available environments keyed by name.
	Environments *map[string]*ops.Environment `mapstructure:"environments"`
	// EnvironmentDefault is a map of available environments keyed by name.
	EnvironmentDefault string `mapstructure:"environment_default,omitempty"`
	// Target is the name of the target environment.
	Target string `mapstructure:"target"`
	// TargetEnv is the target environment to act on.
	TargetEnv *ops.Environment `mapstructure:"-"` // This is not a config value, it's a convenience for the target environment
}
