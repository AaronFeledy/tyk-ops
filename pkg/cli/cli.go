// Package cli provides a command line interface for the application.
package cli

import (
	"github.com/AaronFeledy/tyk-ops/cmd"
)

var (
	RootCmd = cmd.RootCmd

	Cfg = &cmd.Cfg
)
