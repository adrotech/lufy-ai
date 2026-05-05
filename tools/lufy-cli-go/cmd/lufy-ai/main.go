package main

import (
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.Dependencies{Stdout: os.Stdout, Stderr: os.Stderr}))
}
