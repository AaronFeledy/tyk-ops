// Package cli provides a command line interface for the application.
package cli

import (
	"github.com/AaronFeledy/tyk-ops/cmd"
)

const VERSION = cmd.VERSION

var (
	RootCmd = cmd.RootCmd

	Cfg = &cmd.Cfg
)
