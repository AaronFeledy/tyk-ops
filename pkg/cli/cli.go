// Package cli provides a command line interface for the application.
package cli

import (
	"github.com/AaronFeledy/tyk-ops/cmd"
	"github.com/fatih/color"
)

const VERSION = cmd.VERSION

var (
	RootCmd = cmd.RootCmd

	Cfg = &cmd.Cfg

	labelColor = color.New(color.FgMagenta).SprintFunc()
)
