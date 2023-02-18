package main

import (
	"github.com/AaronFeledy/tyk-ops/cmd"
	"os"
)

func main() {
	cmd.RootCmd.SetOut(os.Stdout)
	if err := cmd.RootCmd.Execute(); err != nil {
		cmd.RootCmd.PrintErrln(err)
		os.Exit(1)
	}
}
