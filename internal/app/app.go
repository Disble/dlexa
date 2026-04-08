// Package app provides the composition-root runtime boundary for dlexa.
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/config"
	"github.com/Disble/dlexa/internal/doctor"
	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/modules"
	"github.com/Disble/dlexa/internal/platform"
	"github.com/Disble/dlexa/internal/render"
	"github.com/Disble/dlexa/internal/version"
)

// App owns the runtime-facing application services required by the CLI surface.
type App struct {
	platform platform.CLI
	config   config.Loader
	doctor   doctor.Runner
	registry *modules.Registry
	envelope render.EnvelopeRenderer
}

// NewWithDependencies creates an App with explicit collaborators for tests and wiring.
func NewWithDependencies(cli platform.CLI, loader config.Loader, doctorRunner doctor.Runner, registry *modules.Registry, envelope render.EnvelopeRenderer) *App {
	return &App{
		platform: cli,
		config:   loader,
		doctor:   doctorRunner,
		registry: registry,
		envelope: envelope,
	}
}

// validFormats lists the formats supported by the render layer.
var validFormats = map[string]bool{
	"markdown": true,
	"json":     true,
}

// ExecuteModule runs a registered module and writes its rendered payload to stdout.
func (a *App) ExecuteModule(ctx context.Context, moduleName string, req modules.Request) error {
	if a == nil {
		return fmt.Errorf("application is not configured")
	}
	runtimeConfig, err := a.loadConfig(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.Format) == "" {
		req.Format = runtimeConfig.DefaultFormat
	}
	if len(req.Sources) == 0 {
		req.Sources = append([]string(nil), runtimeConfig.DefaultSources...)
	}
	if !runtimeConfig.CacheEnabled {
		req.NoCache = true
	}

	if !validFormats[strings.TrimSpace(strings.ToLower(req.Format))] {
		return a.HandleSyntaxError(ctx,
			fmt.Errorf("formato %q no soportado; usá markdown o json", req.Format),
			version.BinaryName+" <query> --format markdown|json",
		)
	}

	module, ok := a.registry.Module(moduleName)
	if !ok {
		return a.HandleSyntaxError(ctx, fmt.Errorf("unknown command %q for %q", moduleName, version.BinaryName), version.BinaryName+" --help")
	}

	response, err := module.Execute(ctx, req)
	if err != nil {
		fallback := modules.FallbackFromError(module.Name(), req.Query, req.Format, err)
		return a.writeFallback(ctx, *fallback)
	}
	if response.Fallback != nil {
		if strings.TrimSpace(response.Fallback.Format) == "" {
			response.Fallback.Format = req.Format
		}
		if strings.TrimSpace(response.Fallback.Module) == "" {
			response.Fallback.Module = module.Command()
		}
		return a.writeFallback(ctx, *response.Fallback)
	}

	payload, err := a.envelope.RenderSuccess(ctx, model.Envelope{
		Module:     module.Command(),
		Title:      response.Title,
		Source:     response.Source,
		CacheState: response.CacheState,
		Format:     response.Format,
	}, response.Body)
	if err != nil {
		return err
	}
	return a.writePayload(payload)
}

// RunModule is an alias that keeps the CLI runtime boundary focused on intent.
func (a *App) RunModule(ctx context.Context, moduleName string, req modules.Request) error {
	return a.ExecuteModule(ctx, moduleName, req)
}

// RenderHelp writes Markdown help through the shared envelope renderer.
func (a *App) RenderHelp(ctx context.Context, help model.HelpEnvelope) error {
	payload, err := a.envelope.RenderHelp(ctx, help)
	if err != nil {
		return err
	}
	return a.writePayload(payload)
}

// HandleSyntaxError writes a Level 1 fallback and keeps CLI usage guidance explicit.
func (a *App) HandleSyntaxError(ctx context.Context, err error, syntax string) error {
	message := "El comando es inválido. Corregí la forma antes de volver a intentar."
	if err != nil && strings.TrimSpace(err.Error()) != "" {
		message = err.Error()
	}
	return a.writeFallback(ctx, model.FallbackEnvelope{
		Kind:       model.FallbackKindSyntax,
		Module:     "root",
		Title:      version.BinaryName,
		Message:    message,
		Syntax:     syntax,
		Suggestion: "Usá `--help` para ver sintaxis válida y ejemplos copiables.",
	})
}

// RunDoctor executes the configured doctor runner and writes a plain diagnostic report.
func (a *App) RunDoctor(ctx context.Context) error {
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

// PrintVersion writes version metadata to stdout.
func (a *App) PrintVersion() error {
	_, err := fmt.Fprintf(a.platform.Stdout(), "%s %s\n", version.BinaryName, version.Version)
	return err
}

func (a *App) loadConfig(ctx context.Context) (config.RuntimeConfig, error) {
	if a.config == nil {
		return config.RuntimeConfig{DefaultFormat: "markdown", DefaultSources: []string{"dpd"}, CacheEnabled: true}, nil
	}
	return a.config.Load(ctx)
}

func (a *App) writeFallback(ctx context.Context, fb model.FallbackEnvelope) error {
	payload, err := a.envelope.RenderFallback(ctx, fb)
	if err != nil {
		return err
	}
	return a.writePayload(payload)
}

func (a *App) writePayload(payload []byte) error {
	if _, err := a.platform.Stdout().Write(payload); err != nil {
		return err
	}
	if len(payload) > 0 && payload[len(payload)-1] == '\n' {
		return nil
	}
	_, err := a.platform.Stdout().Write([]byte("\n"))
	return err
}
