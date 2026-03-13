package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gentleman-programming/dlexa/internal/app"
	"github.com/gentleman-programming/dlexa/internal/platform"
)

func main() {
	cli := platform.NewOSCLI(os.Args, os.Stdout, os.Stderr)
	application := app.New(cli)

	if err := application.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
