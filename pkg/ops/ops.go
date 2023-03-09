// Package ops provides a set of tools for managing, developing, and testing Tyk APIs and Gateways.
package ops

var (
	Environments map[string]*Environment
)

type Server struct {
	// Type is the type of server (e.g. "dashboard", "gateway", "mserv").
	Type string `mapstructure:"type" json:"type"`
	// Url is the URL of the server.
	Url string `mapstructure:"url" json:"url"`
	// Secret is the secret used to authenticate with the server.
	Secret string `mapstructure:"secret" json:"secret"`
	// AllowInsecure is a flag that indicates whether or not to allow insecure connections.
	AllowInsecure bool `mapstructure:"insecure" json:"insecure"`
}

// Environment is the configuration for a Tyk environment.
type Environment struct {
	// Name is the name of the environment.
	Name      string `mapstructure:"name" json:"name"`
	Dashboard Server `mapstructure:"dashboard" json:"dashboard"`
	Gateway   Server `mapstructure:"gateway" json:"gateway"`
	Mserv     Server `mapstructure:"mserv" json:"mserv"`
}
