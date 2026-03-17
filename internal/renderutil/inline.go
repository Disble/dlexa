// Package renderutil provides shared inline and table rendering helpers
// used by both the normalize and render packages. It depends only on
// model and stdlib packages.
package renderutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Disble/dlexa/internal/model"
)

// NeedsInlineSpace returns true if a space separator is needed between
// the current accumulated string and the next piece.
func NeedsInlineSpace(current, next string) bool {
	if current == "" || next == "" {
		return false
	}
	last, _ := utf8.DecodeLastRuneInString(current)
	first, _ := utf8.DecodeRuneInString(next)
	if strings.ContainsRune(" [{«", first) || strings.ContainsRune(")]}.;,:!?»", first) {
		return false
	}
	if strings.ContainsRune(" ([{«", last) {
		return false
	}
	if unicode.IsSpace(last) || unicode.IsSpace(first) {
		return false
	}
	return true
}

// ShouldGlueInlineWordBoundary returns true if two adjacent pieces should
// be glued without space (both end/start with letters across markdown markers).
func ShouldGlueInlineWordBoundary(current, next string) bool {
	last, ok := LastInlineWordRune(current)
	if !ok {
		return false
	}
	first, ok := FirstInlineWordRune(next)
	if !ok {
		return false
	}
	return unicode.IsLetter(last) && unicode.IsLetter(first)
}

// ShouldWrapStyledBuffer returns true if a buffer of inlines should be
// wrapped with style markers.
func ShouldWrapStyledBuffer(buffer []model.Inline) bool {
	if len(buffer) != 1 {
		return true
	}
	switch buffer[0].Kind {
	case model.InlineKindMention, model.InlineKindEmphasis, model.InlineKindWorkTitle, model.InlineKindCorrection, model.InlineKindExample:
		return false
	default:
		return true
	}
}

// LastInlineWordRune returns the last significant rune in a string,
// skipping markdown markers.
func LastInlineWordRune(raw string) (rune, bool) {
	trimmed := strings.TrimRightFunc(raw, unicode.IsSpace)
	for trimmed != "" {
		r, size := utf8.DecodeLastRuneInString(trimmed)
		if r == utf8.RuneError && size == 0 {
			return 0, false
		}
		if strings.ContainsRune("*_`~", r) {
			trimmed = trimmed[:len(trimmed)-size]
			continue
		}
		if strings.ContainsRune("])}>", r) {
			return 0, false
		}
		return r, true
	}
	return 0, false
}

// FirstInlineWordRune returns the first significant rune in a string,
// skipping markdown markers.
func FirstInlineWordRune(raw string) (rune, bool) {
	trimmed := strings.TrimLeftFunc(raw, unicode.IsSpace)
	for trimmed != "" {
		r, size := utf8.DecodeRuneInString(trimmed)
		if r == utf8.RuneError && size == 0 {
			return 0, false
		}
		if strings.ContainsRune("*_`~", r) {
			trimmed = trimmed[size:]
			continue
		}
		if strings.ContainsRune("[(<{→", r) {
			return 0, false
		}
		return r, true
	}
	return 0, false
}

// RenderInlineMarkdown renders a slice of model.Inline to markdown string.
// This is the normalize variant: examples are wrapped with ‹›.
func RenderInlineMarkdown(inlines []model.Inline) string {
	var builder strings.Builder
	for _, inline := range inlines {
		piece := renderInlineMarkdownItem(inline)
		if builder.Len() > 0 && NeedsInlineSpace(builder.String(), piece) {
			builder.WriteString(" ")
		}
		builder.WriteString(piece)
	}
	return strings.TrimSpace(builder.String())
}

func renderInlineMarkdownItem(inline model.Inline) string {
	children := RenderInlineMarkdown(inline.Children)
	text := inline.Text
	if children != "" {
		text = children
	}
	switch inline.Kind {
	case model.InlineKindExample:
		return "‹" + text + "›"
	case model.InlineKindMention, model.InlineKindEmphasis, model.InlineKindWorkTitle, model.InlineKindCorrection:
		if len(inline.Children) > 0 {
			return renderStyledInlineMarkdown(inline.Children, "*")
		}
		return "*" + text + "*"
	case model.InlineKindReference:
		if text == "" {
			return ""
		}
		return "→ [" + text + "](" + inline.Target + ")"
	case model.InlineKindScaffold:
		return text
	case model.InlineKindCitationQuote:
		return "«" + text + "»"
	// Phase 1: VALIDATED signs with real HTML evidence
	case model.InlineKindDigitalEdition:
		return text // @ sign preserved as-is
	case model.InlineKindConstructionMarker:
		return text // + sign with construction phrase preserved as-is
	// Phase 2: Bracket semantic contexts (VALIDATED)
	case model.InlineKindBracketDefinition,
		model.InlineKindBracketPronunciation,
		model.InlineKindBracketInterpolation:
		// Markdown output: brackets as plain text
		// JSON output: semantic distinction preserved via InlineKind
		return text
	// WARNING: Speculative signs - no real HTML validation exists yet. These
	// render branches are inferred and MUST be updated when real examples are found.
	case model.InlineKindAgrammatical:
		return text // * sign (SPECULATIVE - synthetic coverage only)
	case model.InlineKindHypothetical:
		return text // ‖ sign (SPECULATIVE - synthetic coverage only)
	case model.InlineKindPhoneme:
		return text // // sign (SPECULATIVE - synthetic coverage only)
	default:
		return text
	}
}

func renderStyledInlineMarkdown(children []model.Inline, marker string) string {
	var builder strings.Builder
	buffer := make([]model.Inline, 0, len(children))
	for _, child := range children {
		if child.Kind != model.InlineKindScaffold {
			buffer = append(buffer, child)
			continue
		}
		flushStyledInlineBuffer(&builder, buffer, marker)
		buffer = buffer[:0]
		plain := strings.TrimSpace(RenderInlineMarkdown(child.Children))
		if plain == "" {
			plain = strings.TrimSpace(child.Text)
		}
		appendInlinePiece(&builder, plain)
	}
	flushStyledInlineBuffer(&builder, buffer, marker)
	return strings.TrimSpace(builder.String())
}

func appendInlinePiece(builder *strings.Builder, piece string) {
	if piece == "" {
		return
	}
	if builder.Len() > 0 && !ShouldGlueInlineWordBoundary(builder.String(), piece) && NeedsInlineSpace(builder.String(), piece) {
		builder.WriteString(" ")
	}
	builder.WriteString(piece)
}

func flushStyledInlineBuffer(builder *strings.Builder, buffer []model.Inline, marker string) {
	text := strings.TrimSpace(RenderInlineMarkdown(buffer))
	if text == "" {
		return
	}
	piece := text
	if ShouldWrapStyledBuffer(buffer) {
		piece = marker + text + marker
	}
	appendInlinePiece(builder, piece)
}

// RenderMarkdownInlines renders inlines for the render layer.
// This is the render variant: examples are wrapped with * (not ‹›).
// It also handles emphasis unwrapping for mention/correction children.
func RenderMarkdownInlines(inlines []model.Inline) string {
	var builder strings.Builder
	for _, inline := range inlines {
		piece := renderMarkdownInline(inline)
		if piece == "" {
			continue
		}
		if builder.Len() > 0 && NeedsInlineSpace(builder.String(), piece) {
			builder.WriteString(" ")
		}
		builder.WriteString(piece)
	}
	return strings.TrimSpace(builder.String())
}

func renderMarkdownInline(inline model.Inline) string {
	text := strings.TrimSpace(inline.Text)
	if len(inline.Children) > 0 {
		if inline.Kind == model.InlineKindEmphasis && len(inline.Children) == 1 {
			child := inline.Children[0]
			if child.Kind == model.InlineKindMention || child.Kind == model.InlineKindCorrection {
				return renderMarkdownInline(child)
			}
		}
		text = RenderMarkdownInlines(inline.Children)
	}
	if text == "" {
		return ""
	}

	switch inline.Kind {
	case model.InlineKindExample,
		model.InlineKindMention,
		model.InlineKindEmphasis,
		model.InlineKindWorkTitle,
		model.InlineKindCorrection:
		if len(inline.Children) > 0 {
			return renderStyledMarkdownInline(inline.Children, "*")
		}
		return "*" + text + "*"
	case model.InlineKindReference:
		return "→ [" + text + "](" + inline.Target + ")"
	case model.InlineKindScaffold:
		return text
	default:
		return text
	}
}

func renderStyledMarkdownInline(children []model.Inline, marker string) string {
	var builder strings.Builder
	buffer := make([]model.Inline, 0, len(children))
	for _, child := range children {
		if child.Kind != model.InlineKindScaffold {
			buffer = append(buffer, child)
			continue
		}
		buffer = flushStyledMarkdownBuffer(&builder, buffer, marker)
		piece := strings.TrimSpace(RenderMarkdownInlines(child.Children))
		if piece == "" {
			piece = strings.TrimSpace(child.Text)
		}
		appendInlinePiece(&builder, piece)
	}
	flushStyledMarkdownBuffer(&builder, buffer, marker)
	return strings.TrimSpace(builder.String())
}

func flushStyledMarkdownBuffer(builder *strings.Builder, buffer []model.Inline, marker string) []model.Inline {
	if len(buffer) == 0 {
		return buffer[:0]
	}
	snapshot := append([]model.Inline(nil), buffer...)
	text := strings.TrimSpace(RenderMarkdownInlines(buffer))
	if text != "" {
		piece := text
		if ShouldWrapStyledBuffer(snapshot) {
			piece = marker + text + marker
		}
		appendInlinePiece(builder, piece)
	}
	return buffer[:0]
}
