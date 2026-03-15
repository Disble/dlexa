package render

import (
	"strings"

	"github.com/Disble/dlexa/internal/model"
	"github.com/Disble/dlexa/internal/renderutil"
)

// TerminalBlockKind classifies a terminal output block.
type TerminalBlockKind string

// Terminal block kinds.
const (
	TerminalBlockKindProse   TerminalBlockKind = "prose"
	TerminalBlockKindExample TerminalBlockKind = "example"
	TerminalBlockKindTable   TerminalBlockKind = "table"
)

// TerminalRunRole classifies the semantic role of a text run.
type TerminalRunRole string

// Terminal run roles.
const (
	TerminalRunRoleProse     TerminalRunRole = "prose"
	TerminalRunRoleEmphasis  TerminalRunRole = "emphasis"
	TerminalRunRoleReference TerminalRunRole = "reference"
)

// TerminalParagraph is a sequence of terminal blocks produced from a model paragraph.
type TerminalParagraph struct {
	Blocks []TerminalBlock
}

// TerminalBlock is a single output block with a kind, optional runs, and optional lines.
type TerminalBlock struct {
	Kind  TerminalBlockKind
	Runs  []TerminalRun
	Lines []string
}

// TerminalRun is a span of text with a semantic role and optional link target.
type TerminalRun struct {
	Role   TerminalRunRole
	Text   string
	Target string
}

func planTerminalParagraph(paragraph model.Paragraph) TerminalParagraph {
	if len(paragraph.Inlines) == 0 {
		text := cleanTerminalProjection(paragraph.Markdown)
		if text == "" {
			return TerminalParagraph{}
		}
		return TerminalParagraph{Blocks: []TerminalBlock{{
			Kind: TerminalBlockKindProse,
			Runs: []TerminalRun{{Role: TerminalRunRoleProse, Text: text}},
		}}}
	}

	planner := terminalParagraphPlanner{}
	for _, inline := range paragraph.Inlines {
		planner.consume(inline)
	}
	planner.flushProse()

	return TerminalParagraph{Blocks: planner.blocks}
}

func planTerminalBlocks(blocks []model.Block) []TerminalBlock {
	planned := make([]TerminalBlock, 0, len(blocks))
	for _, block := range blocks {
		switch block.Kind {
		case model.ArticleBlockKindParagraph:
			if block.Paragraph == nil {
				continue
			}
			planned = append(planned, planTerminalParagraph(*block.Paragraph).Blocks...)
		case model.ArticleBlockKindTable:
			if block.Table == nil {
				continue
			}
			text := renderutil.RenderTableMarkdown(*block.Table, "")
			if strings.TrimSpace(text) == "" {
				continue
			}
			planned = append(planned, TerminalBlock{Kind: TerminalBlockKindTable, Lines: strings.Split(text, "\n")})
		}
	}
	return planned
}

type terminalParagraphPlanner struct {
	blocks   []TerminalBlock
	proseRun []TerminalRun
}

func (p *terminalParagraphPlanner) consume(inline model.Inline) {
	switch inline.Kind {
	case model.InlineKindExample:
		p.flushProse()
		text := flattenInlineText(inline)
		if text == "" {
			return
		}
		p.blocks = append(p.blocks, TerminalBlock{
			Kind: TerminalBlockKindExample,
			Runs: []TerminalRun{{Role: TerminalRunRoleProse, Text: text, Target: inline.Target}},
		})
	case model.InlineKindMention, model.InlineKindEmphasis:
		p.appendRun(TerminalRun{Role: TerminalRunRoleEmphasis, Text: flattenInlineText(inline), Target: inline.Target})
	case model.InlineKindReference:
		p.appendRun(TerminalRun{Role: TerminalRunRoleReference, Text: flattenInlineText(inline), Target: inline.Target})
	default:
		if inline.Text == "" && len(inline.Children) > 0 {
			for _, child := range inline.Children {
				p.consume(child)
			}
			return
		}
		p.appendRun(TerminalRun{Role: TerminalRunRoleProse, Text: flattenInlineText(inline), Target: inline.Target})
	}
}

func (p *terminalParagraphPlanner) appendRun(run TerminalRun) {
	if compactTerminalWhitespace(run.Text) == "" {
		return
	}
	if len(p.proseRun) > 0 {
		last := &p.proseRun[len(p.proseRun)-1]
		if last.Role == run.Role && last.Target == run.Target {
			last.Text += run.Text
			return
		}
	}
	p.proseRun = append(p.proseRun, run)
}

func (p *terminalParagraphPlanner) flushProse() {
	if len(p.proseRun) == 0 {
		return
	}
	p.blocks = append(p.blocks, TerminalBlock{Kind: TerminalBlockKindProse, Runs: append([]TerminalRun(nil), p.proseRun...)})
	p.proseRun = nil
}

func flattenInlineText(inline model.Inline) string {
	if len(inline.Children) == 0 {
		return inline.Text
	}

	var builder strings.Builder
	for _, child := range inline.Children {
		builder.WriteString(flattenInlineText(child))
	}
	if builder.Len() > 0 {
		return builder.String()
	}
	return inline.Text
}

func compactTerminalWhitespace(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}
