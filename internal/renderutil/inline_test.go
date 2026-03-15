package renderutil

import (
	"testing"

	"github.com/Disble/dlexa/internal/model"
)

const (
	testSingleText = "single text"
	testBienRef1   = "bien#ref1"
)

func TestNeedsInlineSpace(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		{"empty current", "", "word", false},
		{"empty next", "word", "", false},
		{"both empty", "", "", false},
		{"word to word", "hello", "world", true},
		{"trailing space", "hello ", "world", false},
		{"leading space", "hello", " world", false},
		{"next starts with open bracket", "word", "[ref]", false},
		{"next starts with open brace", "word", "{foo}", false},
		{"next starts with guillemet", "word", "«quote»", false},
		{"next starts with close paren", "word", ")", false},
		{"next starts with close bracket", "word", "]", false},
		{"next starts with period", "word", ".", false},
		{"next starts with semicolon", "word", ";", false},
		{"next starts with comma", "word", ",", false},
		{"next starts with colon", "word", ":", false},
		{"next starts with exclamation", "word", "!", false},
		{"next starts with question", "word", "?", false},
		{"next starts with close guillemet", "word", "»", false},
		{"current ends with open paren", "hello(", "word", false},
		{"current ends with open bracket", "hello[", "word", false},
		{"current ends with open brace", "hello{", "word", false},
		{"current ends with open guillemet", "hello«", "word", false},
		{"current ends with space char", "hello ", "world", false},
		{"letter to letter", "a", "b", true},
		{"unicode letters", "café", "latte", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsInlineSpace(tt.current, tt.next)
			if got != tt.want {
				t.Fatalf("NeedsInlineSpace(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.want)
			}
		})
	}
}

func TestShouldGlueInlineWordBoundary(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		{"letter to letter", "hello", "world", true},
		{"letter to marker then letter", "hello*", "*world", true},
		{"letter to punctuation", "hello", ".", false},
		{"punctuation to letter", "hello.", "world", false},
		{"bracket to letter", "hello]", "world", false},
		{"letter to bracket", "hello", "[world", false},
		{"empty current", "", "world", false},
		{"empty next", "hello", "", false},
		{"markdown marker wrapping", "*word*", "*next*", true},
		{"spaces then marker", "word  ", "*next", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldGlueInlineWordBoundary(tt.current, tt.next)
			if got != tt.want {
				t.Fatalf("ShouldGlueInlineWordBoundary(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.want)
			}
		})
	}
}

func TestShouldWrapStyledBuffer(t *testing.T) {
	tests := []struct {
		name   string
		buffer []model.Inline
		want   bool
	}{
		{"empty buffer", nil, true},
		{"multiple items", []model.Inline{{Kind: model.InlineKindText}, {Kind: model.InlineKindText}}, true},
		{testSingleText, []model.Inline{{Kind: model.InlineKindText}}, true},
		{"single mention", []model.Inline{{Kind: model.InlineKindMention}}, false},
		{"single emphasis", []model.Inline{{Kind: model.InlineKindEmphasis}}, false},
		{"single work title", []model.Inline{{Kind: model.InlineKindWorkTitle}}, false},
		{"single correction", []model.Inline{{Kind: model.InlineKindCorrection}}, false},
		{"single example", []model.Inline{{Kind: model.InlineKindExample}}, false},
		{"single scaffold", []model.Inline{{Kind: model.InlineKindScaffold}}, true},
		{"single reference", []model.Inline{{Kind: model.InlineKindReference}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldWrapStyledBuffer(tt.buffer)
			if got != tt.want {
				t.Fatalf("ShouldWrapStyledBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastInlineWordRune(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantRune rune
		wantOK   bool
	}{
		{"empty", "", 0, false},
		{"simple letter", "abc", 'c', true},
		{"trailing space", "abc  ", 'c', true},
		{"trailing markdown markers", "abc*", 'c', true},
		{"only markers", "***", 0, false},
		{"ends with close bracket", "abc]", 0, false},
		{"ends with close paren", "abc)", 0, false},
		{"ends with close brace", "abc}", 0, false},
		{"ends with close angle", "abc>", 0, false},
		{"unicode letter", "café", 'é', true},
		{"trailing tilde marker", "abc~", 'c', true},
		{"trailing backtick marker", "abc`", 'c', true},
		{"trailing underscore marker", "abc_", 'c', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRune, gotOK := LastInlineWordRune(tt.raw)
			if gotOK != tt.wantOK || (gotOK && gotRune != tt.wantRune) {
				t.Fatalf("LastInlineWordRune(%q) = (%q, %v), want (%q, %v)", tt.raw, gotRune, gotOK, tt.wantRune, tt.wantOK)
			}
		})
	}
}

func TestFirstInlineWordRune(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantRune rune
		wantOK   bool
	}{
		{"empty", "", 0, false},
		{"simple letter", "abc", 'a', true},
		{"leading space", "  abc", 'a', true},
		{"leading markdown markers", "*abc", 'a', true},
		{"only markers", "***", 0, false},
		{"starts with open bracket", "[abc", 0, false},
		{"starts with open paren", "(abc", 0, false},
		{"starts with open brace", "{abc", 0, false},
		{"starts with open angle", "<abc", 0, false},
		{"starts with arrow", "→abc", 0, false},
		{"unicode letter", "café", 'c', true},
		{"leading tilde marker", "~abc", 'a', true},
		{"leading backtick marker", "`abc", 'a', true},
		{"leading underscore marker", "_abc", 'a', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRune, gotOK := FirstInlineWordRune(tt.raw)
			if gotOK != tt.wantOK || (gotOK && gotRune != tt.wantRune) {
				t.Fatalf("FirstInlineWordRune(%q) = (%q, %v), want (%q, %v)", tt.raw, gotRune, gotOK, tt.wantRune, tt.wantOK)
			}
		})
	}
}

func TestRenderStyledInlineMarkdownVariants(t *testing.T) {
	tests := []struct {
		name    string
		inlines []model.Inline
		want    string
	}{
		{
			"emphasis with text children wraps",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "hello world"},
				},
			}},
			"*hello world*",
		},
		{
			"mention with multiple text children",
			[]model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "first"},
					{Kind: model.InlineKindText, Text: " second"},
				},
			}},
			"*first second*",
		},
		{
			"emphasis with empty scaffold child",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "abc"},
					{Kind: model.InlineKindScaffold, Text: ""},
				},
			}},
			"*abc*",
		},
		{
			"mention with scaffold that has children",
			[]model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "abc"},
					{Kind: model.InlineKindScaffold, Children: []model.Inline{
						{Kind: model.InlineKindText, Text: "def"},
					}},
				},
			}},
			"*abc*def",
		},
		{
			"work title with children",
			[]model.Inline{{
				Kind: model.InlineKindWorkTitle,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "Don Quijote"},
				},
			}},
			"*Don Quijote*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderInlineMarkdown(tt.inlines)
			if got != tt.want {
				t.Fatalf("RenderInlineMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderStyledMarkdownInlineVariants(t *testing.T) {
	tests := []struct {
		name    string
		inlines []model.Inline
		want    string
	}{
		{
			"emphasis with text children wraps",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "hello world"},
				},
			}},
			"*hello world*",
		},
		{
			"mention with multiple text children",
			[]model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "first"},
					{Kind: model.InlineKindText, Text: " second"},
				},
			}},
			"*first second*",
		},
		{
			"emphasis with empty scaffold child",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "abc"},
					{Kind: model.InlineKindScaffold, Text: ""},
				},
			}},
			"*abc*",
		},
		{
			"mention with scaffold that has children",
			[]model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "abc"},
					{Kind: model.InlineKindScaffold, Children: []model.Inline{
						{Kind: model.InlineKindText, Text: "def"},
					}},
				},
			}},
			"*abc*def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderMarkdownInlines(tt.inlines)
			if got != tt.want {
				t.Fatalf("RenderMarkdownInlines() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderInlineMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		inlines []model.Inline
		want    string
	}{
		{"nil inlines", nil, ""},
		{"empty inlines", []model.Inline{}, ""},
		{testSingleText, []model.Inline{{Kind: model.InlineKindText, Text: "hello"}}, "hello"},
		{"text with spaces", []model.Inline{{Kind: model.InlineKindText, Text: "  hello  "}}, "hello"},
		{
			"example wraps with angle quotes",
			[]model.Inline{{Kind: model.InlineKindExample, Text: "Cierra bien"}},
			"‹Cierra bien›",
		},
		{
			"mention wraps with asterisks",
			[]model.Inline{{Kind: model.InlineKindMention, Text: "mejor"}},
			"*mejor*",
		},
		{
			"emphasis wraps with asterisks",
			[]model.Inline{{Kind: model.InlineKindEmphasis, Text: "muy"}},
			"*muy*",
		},
		{
			"reference renders arrow link",
			[]model.Inline{{Kind: model.InlineKindReference, Text: "6", Target: testBienRef1}},
			"→ [6](bien#ref1)",
		},
		{
			"empty reference returns empty",
			[]model.Inline{{Kind: model.InlineKindReference, Text: "", Target: testBienRef1}},
			"",
		},
		{
			"citation quote wraps with guillemets",
			[]model.Inline{{Kind: model.InlineKindCitationQuote, Text: "quote text"}},
			"«quote text»",
		},
		{
			"scaffold returns plain text",
			[]model.Inline{{Kind: model.InlineKindScaffold, Text: "plain"}},
			"plain",
		},
		{
			"mixed text and example",
			[]model.Inline{
				{Kind: model.InlineKindText, Text: "El comparativo es "},
				{Kind: model.InlineKindMention, Text: "mejor"},
				{Kind: model.InlineKindText, Text: ". "},
				{Kind: model.InlineKindExample, Text: "Cierra bien la ventana"},
				{Kind: model.InlineKindText, Text: "."},
			},
			"El comparativo es *mejor*. ‹Cierra bien la ventana›.",
		},
		{
			"emphasis with mention child unwraps",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{{
					Kind: model.InlineKindMention,
					Text: "tilde",
				}},
			}},
			"*tilde*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderInlineMarkdown(tt.inlines)
			if got != tt.want {
				t.Fatalf("RenderInlineMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderMarkdownInlines(t *testing.T) {
	tests := []struct {
		name    string
		inlines []model.Inline
		want    string
	}{
		{"nil inlines", nil, ""},
		{"empty inlines", []model.Inline{}, ""},
		{testSingleText, []model.Inline{{Kind: model.InlineKindText, Text: "hello"}}, "hello"},
		{
			"example wraps with asterisks (not angle quotes)",
			[]model.Inline{{Kind: model.InlineKindExample, Text: "Cierra bien"}},
			"*Cierra bien*",
		},
		{
			"mention wraps with asterisks",
			[]model.Inline{{Kind: model.InlineKindMention, Text: "mejor"}},
			"*mejor*",
		},
		{
			"reference renders arrow link",
			[]model.Inline{{Kind: model.InlineKindReference, Text: "6", Target: testBienRef1}},
			"→ [6](bien#ref1)",
		},
		{
			"mixed text and example uses asterisks",
			[]model.Inline{
				{Kind: model.InlineKindText, Text: "El comparativo es "},
				{Kind: model.InlineKindMention, Text: "mejor"},
				{Kind: model.InlineKindText, Text: ". "},
				{Kind: model.InlineKindExample, Text: "Cierra bien la ventana"},
				{Kind: model.InlineKindText, Text: "."},
			},
			"El comparativo es *mejor*. *Cierra bien la ventana*.",
		},
		{
			"emphasis with mention child unwraps",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{{
					Kind: model.InlineKindMention,
					Text: "tilde",
				}},
			}},
			"*tilde*",
		},
		{
			"scaffold children with mention word fragments glue",
			[]model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "gr"},
					{Kind: model.InlineKindScaffold, Text: "úa"},
				},
			}},
			"*gr*úa",
		},
		{
			"nested plain and styled suffix stays glued",
			[]model.Inline{{
				Kind: model.InlineKindMention,
				Children: []model.Inline{
					{Kind: model.InlineKindText, Text: "anch"},
					{Kind: model.InlineKindScaffold, Text: "oa"},
					{Kind: model.InlineKindCorrection, Text: "s"},
				},
			}},
			"*anch*oa*s*",
		},
		{
			"emphasis with scaffold child override",
			[]model.Inline{{
				Kind: model.InlineKindEmphasis,
				Children: []model.Inline{{
					Kind: model.InlineKindMention,
					Children: []model.Inline{
						{Kind: model.InlineKindText, Text: "tilde"},
						{Kind: model.InlineKindScaffold, Text: "2"},
					},
				}},
			}},
			"*tilde* 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderMarkdownInlines(tt.inlines)
			if got != tt.want {
				t.Fatalf("RenderMarkdownInlines() = %q, want %q", got, tt.want)
			}
		})
	}
}
