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

// RenderHelp renders capability-first Markdown help with syntax,
// accepted input guidance, examples, and agent notes.
func (r *MarkdownEnvelopeRenderer) RenderHelp(ctx context.Context, help model.HelpEnvelope) ([]byte, error) {
	_ = ctx
	var builder strings.Builder
	command := helpCommand(help)
	builder.WriteString("# Ayuda: ")
	builder.WriteString(command)
	builder.WriteString("\n\n")
	appendHelpSummary(&builder, help.Summary)
	appendHelpSyntax(&builder, help.Syntax)
	appendHelpList(&builder, "## Qué podés hacer\n", normalizeNonEmpty(help.Capabilities), false)
	appendHelpList(&builder, "## Qué recibe\n", normalizeNonEmpty(help.InputHints), false)
	appendHelpList(&builder, "## Ejemplos\n", normalizeNonEmpty(help.Examples), true)
	appendHelpList(&builder, "## Guía para agentes y automatizaciones\n", normalizeNonEmpty(help.AgentNotes), false)
	appendHelpList(&builder, "## Próximo paso sugerido\n", normalizeNonEmpty(help.NextSteps), false)
	return []byte(strings.TrimRight(builder.String(), "\n")), nil
}

func helpCommand(help model.HelpEnvelope) string {
	command := strings.TrimSpace(help.Command)
	if command == "" {
		return "dlexa"
	}
	return command
}

func appendHelpSummary(builder *strings.Builder, summary string) {
	if summary = strings.TrimSpace(summary); summary != "" {
		builder.WriteString(summary)
		builder.WriteString("\n\n")
	}
}

func appendHelpSyntax(builder *strings.Builder, syntax string) {
	if syntax = strings.TrimSpace(syntax); syntax != "" {
		builder.WriteString("## Sintaxis\n")
		builder.WriteString("`")
		builder.WriteString(syntax)
		builder.WriteString("`\n\n")
	}
}

func appendHelpList(builder *strings.Builder, heading string, items []string, quote bool) {
	if len(items) == 0 {
		return
	}
	builder.WriteString(heading)
	for _, item := range items {
		builder.WriteString("- ")
		if quote {
			builder.WriteString("`")
		}
		builder.WriteString(item)
		if quote {
			builder.WriteString("`")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func normalizeNonEmpty(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
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
