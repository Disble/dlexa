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
	runtime := app.New(cli)
	if err := executeRootCommand(context.Background(), runtime, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
