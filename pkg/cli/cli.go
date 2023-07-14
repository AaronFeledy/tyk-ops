// Package cli provides a command line interface for the application.
package cli

import (
	"github.com/fatih/color"
)

const VERSION = "0.0.1"

var (
	cfgFile    string
	labelColor = color.New(color.FgMagenta).SprintFunc()
)
