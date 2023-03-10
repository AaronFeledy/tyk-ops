package cmd

import (
	"github.com/AaronFeledy/tyk-ops/pkg/ops"
	out "github.com/AaronFeledy/tyk-ops/pkg/output"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var (
	Cfg     Config
	cfgFile string
)

var RootCmd = &cobra.Command{
	Use:  "tykops",
	Long: "A tool to manage syncing and deployments of Tyk Gateways and their middleware bundles.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			out.User.Errorln(err)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to config file")
	RootCmd.PersistentFlags().String("target", "", "A target environment to use as defined in your configuration file. You may use @target_name as shorthand for this flag.")

	_ = viper.BindPFlag("target", RootCmd.PersistentFlags().Lookup("target"))

	// Support @ shorthand for target flag in commands (e.g. tykops @dev deploy)
	args := os.Args
	for i, arg := range args {
		if arg[:1] == "@" {
			args[i] = "--target=" + arg[1:]
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
