package main

import (
	"github.com/AaronFeledy/tyk-ops/pkg/cli"
	"os"
)

func main() {
	cli.RootCmd.SetOut(os.Stdout)
	if err := cli.RootCmd.Execute(); err != nil {
		cli.RootCmd.PrintErrln(err)
		os.Exit(1)
	}
}
