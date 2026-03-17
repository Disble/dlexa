// Package app provides the main application entry point for dlexa.
package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/doctor"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/platform"
	"github.com/Disble/dlexa/internal/query"
	"github.com/Disble/dlexa/internal/render"
	searchsvc "github.com/Disble/dlexa/internal/search"
	"github.com/Disble/dlexa/internal/version"
)

// App wires together configuration, lookup, and rendering to drive the CLI.
type App struct {
	platform        platform.CLI
	config          config.Loader
	doctor          doctor.Runner
	lookup          query.Looker
	search          searchsvc.Searcher
	renderers       render.RendererResolver
	searchRenderers render.SearchRendererResolver
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

	positional := flagSet.Args()
	if len(positional) == 0 {
		return a.printUsage()
	}

	runtimeConfig, err := a.config.Load(ctx)
	if err != nil {
		return err
	}

	formatName := resolveFormat(*formatFlag, runtimeConfig.DefaultFormat)
	if positional[0] == "search" {
		return a.runSearch(ctx, positional[1:], formatName, *noCacheFlag || !runtimeConfig.CacheEnabled)
	}

	queryText := strings.TrimSpace(strings.Join(positional, " "))

	request := model.LookupRequest{
		Query:   queryText,
		Format:  formatName,
		Sources: requestedSources(*sourceFlag, runtimeConfig.DefaultSources),
		NoCache: *noCacheFlag || !runtimeConfig.CacheEnabled,
	}

	return a.runLookup(ctx, request, formatName)
}

func (a *App) runSearch(ctx context.Context, args []string, formatName string, noCache bool) error {
	queryText := strings.TrimSpace(strings.Join(args, " "))
	if queryText == "" {
		if err := a.printSearchUsage(); err != nil {
			return err
		}
		return errors.New("search command requires a query")
	}

	request := model.SearchRequest{Query: queryText, Format: formatName, NoCache: noCache}
	result, err := a.search.Search(ctx, request)
	if err != nil {
		return err
	}

	renderer, err := a.searchRenderers.Renderer(formatName)
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
	return a.ensureTrailingNewline(payload)
}

// runLookup executes the lookup, renders the result, and writes the output.
func (a *App) runLookup(ctx context.Context, request model.LookupRequest, formatName string) error {
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

	return a.ensureTrailingNewline(payload)
}

// ensureTrailingNewline writes a newline to stdout if payload does not already end with one.
func (a *App) ensureTrailingNewline(payload []byte) error {
	if len(payload) > 0 && payload[len(payload)-1] == '\n' {
		return nil
	}
	_, err := a.platform.Stdout().Write([]byte("\n"))
	return err
}

// resolveFormat returns the explicit format if non-empty, otherwise the default.
func resolveFormat(explicit, defaultFormat string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	return defaultFormat
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
		"usage: %s [--format markdown|json] [--source name1,name2] [--no-cache] <query>\n       %s [--format markdown|json] [--no-cache] search <query>\n",
		version.BinaryName,
		version.BinaryName,
	)
	return err
}

func (a *App) printSearchUsage() error {
	_, err := fmt.Fprintf(a.platform.Stderr(), "usage: %s [--format markdown|json] [--no-cache] search <query>\n", version.BinaryName)
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
