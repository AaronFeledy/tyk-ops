package cmd

import (
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
	Use:   "tykops",
	Short: "A tool to manage Tyk environments",
	Long: `A tool to manage syncing and deployments of Tyk Gateways and their
           middleware bundles.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file)")
	RootCmd.PersistentFlags().String("target", "", "a target environment to use as defined in .tykops.yml")

	_ = viper.BindPFlag("target", RootCmd.PersistentFlags().Lookup("target"))
	viper.SetDefault("target", "default")
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
		if viper.IsSet("environments." + target) {
			out.User.Debugln("Using target environment '" + target + "'")
			targetEnv = true
		} else {
			if target == "default" {
				target = ""
			} else {
				out.User.Errorln("Target environment '" + target + "' not found in " + viper.ConfigFileUsed())
				os.Exit(1)
			}
		}

		// Load the config into the global variable
		if err := viper.Unmarshal(&Cfg); err != nil {
			out.User.Errorln(err.Error())
			return
		}
		// Add shorthand for the target environment
		if targetEnv {
			Cfg.TargetEnv = Cfg.Environments[target]
		}
	}
}

type Config struct {
	Environments map[string]*Environment `mapstructure:"environments"`
	Target       string                  `mapstructure:"target"`
	TargetEnv    *Environment            `mapstructure:"target_env,omitempty"`
}

type Environment struct {
	Dashboard struct {
		Url    string `mapstructure:"url"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"dashboard"`
	Gateway struct {
		Url    string `mapstructure:"url"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"gateway"`
	Mserv struct {
		Url    string `mapstructure:"url"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"mserv"`
}
