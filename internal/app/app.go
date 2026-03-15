// Package app provides the main application entry point for dlexa.
package app

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gentleman-programming/dlexa/internal/config"
	"github.com/gentleman-programming/dlexa/internal/doctor"
	"github.com/gentleman-programming/dlexa/internal/model"
	"github.com/gentleman-programming/dlexa/internal/platform"
	"github.com/gentleman-programming/dlexa/internal/query"
	"github.com/gentleman-programming/dlexa/internal/render"
	"github.com/gentleman-programming/dlexa/internal/version"
)

// App wires together configuration, lookup, and rendering to drive the CLI.
type App struct {
	platform  platform.CLI
	config    config.Loader
	doctor    doctor.Service
	lookup    query.Service
	renderers render.Registry
}

// Run parses CLI flags, performs the lookup, and writes rendered output.
func (a *App) Run(ctx context.Context) error {
	flagSet := flag.NewFlagSet(version.BinaryName, flag.ContinueOnError)
	flagSet.SetOutput(a.platform.Stderr())

	formatFlag := flagSet.String("format", "", "render format: markdown|json")
	sourceFlag := flagSet.String("source", "", "comma-separated source names")
	noCacheFlag := flagSet.Bool("no-cache", false, "skip cache reads and writes")
	doctorFlag := flagSet.Bool("doctor", false, "run environment checks")
	versionFlag := flagSet.Bool("version", false, "print version information")

	args := a.platform.Args()
	if len(args) > 1 {
		if err := flagSet.Parse(args[1:]); err != nil {
			return err
		}
	}

	if *versionFlag {
		_, err := fmt.Fprintf(a.platform.Stdout(), "%s %s\n", version.BinaryName, version.Version)
		return err
	}

	if *doctorFlag {
		return a.runDoctor(ctx)
	}

	queryText := strings.TrimSpace(strings.Join(flagSet.Args(), " "))
	if queryText == "" {
		return a.printUsage()
	}

	runtimeConfig, err := a.config.Load(ctx)
	if err != nil {
		return err
	}

	formatName := runtimeConfig.DefaultFormat
	if strings.TrimSpace(*formatFlag) != "" {
		formatName = strings.TrimSpace(*formatFlag)
	}

	request := model.LookupRequest{
		Query:   queryText,
		Format:  formatName,
		Sources: requestedSources(*sourceFlag, runtimeConfig.DefaultSources),
		NoCache: *noCacheFlag || !runtimeConfig.CacheEnabled,
	}

	result, err := a.lookup.Lookup(ctx, request)
	if err != nil {
		return err
	}

	renderer, err := a.renderers.Renderer(formatName)
	if err != nil {
		return err
	}

	payload, err := renderer.Render(ctx, result)
	if err != nil {
		return err
	}

	if _, err := a.platform.Stdout().Write(payload); err != nil {
		return err
	}

	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		if _, err := a.platform.Stdout().Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) runDoctor(ctx context.Context) error {
	report, err := a.doctor.Run(ctx)
	if err != nil {
		return err
	}

	status := "ok"
	if !report.Healthy {
		status = "degraded"
	}

	if _, err := fmt.Fprintf(a.platform.Stdout(), "doctor: %s\n", status); err != nil {
		return err
	}

	for _, check := range report.Checks {
		if _, err := fmt.Fprintf(a.platform.Stdout(), "- %s [%s] %s\n", check.Name, check.Status, check.Detail); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) printUsage() error {
	_, err := fmt.Fprintf(
		a.platform.Stderr(),
		"usage: %s [--format markdown|json] [--source name1,name2] [--no-cache] <query>\n",
		version.BinaryName,
	)
	return err
}

func requestedSources(raw string, fallback []string) []string {
	if strings.TrimSpace(raw) == "" {
		return append([]string(nil), fallback...)
	}

	parts := strings.Split(raw, ",")
	selected := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			selected = append(selected, trimmed)
		}
	}

	return selected
}
