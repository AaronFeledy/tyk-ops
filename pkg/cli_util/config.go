package cli_util

import "github.com/AaronFeledy/tyk-ops/pkg/ops"

var Config *ConfigData

// ConfigData is the configuration for the TykOps client.
type ConfigData struct {
	// Environments is a map of available environments keyed by name.
	Environments *map[string]*ops.Environment `mapstructure:"environments"`
	// EnvironmentDefault is a map of available environments keyed by name.
	EnvironmentDefault string `mapstructure:"environment_default,omitempty"`
	// Target is the name of the target environment.
	Target string `mapstructure:"target"`
	// TargetEnv is the target environment to act on.
	TargetEnv *ops.Environment `mapstructure:"-"` // This is not a config value, it's a convenience for the target environment
}

func init() {
	Config = new(ConfigData)
}
