package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/Disble/dlexa/internal/model"
)

// MarkdownEnvelopeRenderer materializes the agent-facing Markdown envelope and fallback ladder.
type MarkdownEnvelopeRenderer struct{}

// NewEnvelopeRenderer creates a shared envelope renderer.
func NewEnvelopeRenderer() *MarkdownEnvelopeRenderer {
	return &MarkdownEnvelopeRenderer{}
}

// RenderSuccess wraps Markdown payloads with a canonical envelope and bypasses JSON untouched.
func (r *MarkdownEnvelopeRenderer) RenderSuccess(ctx context.Context, env model.Envelope, body []byte) ([]byte, error) {
	_ = ctx
	if strings.EqualFold(strings.TrimSpace(env.Format), "json") {
		return append([]byte(nil), body...), nil
	}

	moduleName := strings.TrimSpace(env.Module)
	if moduleName == "" {
		moduleName = "unknown"
	}
	title := strings.TrimSpace(env.Title)
	if title == "" {
		title = "resultado"
	}
	source := strings.TrimSpace(env.Source)
	if source == "" {
		source = "sin fuente"
	}
	cacheState := strings.TrimSpace(env.CacheState)
	if cacheState == "" {
		cacheState = "UNKNOWN"
	}
	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return []byte(fmt.Sprintf("# [dlexa:%s] %s\n*Fuente: %s | Caché: %s*\n", moduleName, title, source, cacheState)), nil
	}
	return []byte(fmt.Sprintf("# [dlexa:%s] %s\n*Fuente: %s | Caché: %s*\n\n---\n\n%s", moduleName, title, source, cacheState, bodyText)), nil
}

// RenderHelp renders agent-oriented Markdown help with syntax, examples, and recovery guidance.
func (r *MarkdownEnvelopeRenderer) RenderHelp(ctx context.Context, help model.HelpEnvelope) ([]byte, error) {
	_ = ctx
	var builder strings.Builder
	command := strings.TrimSpace(help.Command)
	if command == "" {
		command = "dlexa"
	}
	builder.WriteString("# Ayuda: ")
	builder.WriteString(command)
	builder.WriteString("\n\n")
	if summary := strings.TrimSpace(help.Summary); summary != "" {
		builder.WriteString(summary)
		builder.WriteString("\n\n")
	}
	if syntax := strings.TrimSpace(help.Syntax); syntax != "" {
		builder.WriteString("## Sintaxis\n")
		builder.WriteString("`")
		builder.WriteString(syntax)
		builder.WriteString("`\n\n")
	}
	if len(help.Examples) > 0 {
		builder.WriteString("## Ejemplos\n")
		for _, example := range help.Examples {
			if strings.TrimSpace(example) == "" {
				continue
			}
			builder.WriteString("- `")
			builder.WriteString(strings.TrimSpace(example))
			builder.WriteString("`\n")
		}
		builder.WriteString("\n")
	}
	if len(help.NextSteps) > 0 {
		builder.WriteString("## Siguiente paso sugerido\n")
		for _, step := range help.NextSteps {
			if strings.TrimSpace(step) == "" {
				continue
			}
			builder.WriteString("- ")
			builder.WriteString(strings.TrimSpace(step))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	if recovery := strings.TrimSpace(help.RecoveryTip); recovery != "" {
		builder.WriteString("## Si falla\n")
		builder.WriteString(recovery)
		builder.WriteString("\n")
	}
	return []byte(strings.TrimRight(builder.String(), "\n")), nil
}

// RenderFallback renders the explicit fallback tiers.
func (r *MarkdownEnvelopeRenderer) RenderFallback(ctx context.Context, fb model.FallbackEnvelope) ([]byte, error) {
	_ = ctx
	if strings.EqualFold(strings.TrimSpace(fb.Format), "json") {
		return marshalNoEscape(fb)
	}
	var builder strings.Builder
	moduleName := strings.TrimSpace(fb.Module)
	if moduleName == "" {
		moduleName = "fallback"
	}
	title := strings.TrimSpace(fb.Title)
	if title == "" {
		title = strings.TrimSpace(fb.Query)
	}
	if title == "" {
		title = fallbackLabel(fb.Kind)
	}
	fmt.Fprintf(&builder, "# [dlexa:%s] Nivel %d · %s\n\n", moduleName, fallbackLevel(fb.Kind), fallbackLabel(fb.Kind))
	if fb.Message != "" {
		builder.WriteString(fb.Message)
		builder.WriteString("\n\n")
	}
	if title != "" && title != fallbackLabel(fb.Kind) {
		builder.WriteString("- objetivo: `")
		builder.WriteString(title)
		builder.WriteString("`\n")
	}
	if detail := strings.TrimSpace(fb.Detail); detail != "" {
		builder.WriteString("- detalle: ")
		builder.WriteString(detail)
		builder.WriteString("\n")
	}
	if syntax := strings.TrimSpace(fb.Syntax); syntax != "" {
		builder.WriteString("- sintaxis correcta: `")
		builder.WriteString(syntax)
		builder.WriteString("`\n")
	}
	if next := strings.TrimSpace(fb.NextCommand); next != "" {
		builder.WriteString("- siguiente comando: `")
		builder.WriteString(next)
		builder.WriteString("`\n")
	}
	if suggestion := strings.TrimSpace(fb.Suggestion); suggestion != "" {
		builder.WriteString("- acción: ")
		builder.WriteString(suggestion)
		builder.WriteString("\n")
	}
	return []byte(strings.TrimRight(builder.String(), "\n")), nil
}

func fallbackLevel(kind model.FallbackKind) int {
	switch kind {
	case model.FallbackKindSyntax:
		return 1
	case model.FallbackKindNotFound:
		return 2
	case model.FallbackKindRateLimited:
		return 3
	case model.FallbackKindUpstreamUnavailable:
		return 3
	case model.FallbackKindParseFailure:
		return 4
	default:
		return 4
	}
}

func fallbackLabel(kind model.FallbackKind) string {
	switch kind {
	case model.FallbackKindSyntax:
		return "Syntax"
	case model.FallbackKindNotFound:
		return "Not Found"
	case model.FallbackKindRateLimited:
		return "Rate Limited"
	case model.FallbackKindUpstreamUnavailable:
		return "Upstream Unavailable"
	case model.FallbackKindParseFailure:
		return "Parse Failure"
	default:
		return "Fallback"
	}
}
