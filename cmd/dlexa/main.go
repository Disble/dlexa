// Package main is the entry point for the dlexa CLI binary.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Disble/dlexa/internal/app"
	"github.com/Disble/dlexa/internal/platform"
)

func main() {
	cli := platform.NewOSCLI(os.Args, os.Stdout, os.Stderr)
	application := app.New(cli)

	if err := application.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
