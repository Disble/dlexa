package render

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/dlexa/internal/model"
)

func TestPlanTerminalParagraphUsesTypedInlinesAsSemanticBoundary(t *testing.T) {
	paragraph := model.Paragraph{
		Markdown: "[ej.: roto] ‹todavia peor› *colapsado*",
		Inlines: []model.Inline{
			{Kind: model.InlineKindText, Text: "Arranca con "},
			{Kind: model.InlineKindMention, Text: "foco"},
			{Kind: model.InlineKindText, Text: ": "},
			{Kind: model.InlineKindExample, Text: "ejemplo real"},
		},
	}

	plan := planTerminalParagraph(paragraph)

	if got, want := blockKinds(plan), []TerminalBlockKind{TerminalBlockKindProse, TerminalBlockKindExample}; !reflect.DeepEqual(got, want) {
		t.Fatalf("blockKinds = %v, want %v", got, want)
	}
	if got, want := runRoles(plan.Blocks[0]), []TerminalRunRole{TerminalRunRoleProse, TerminalRunRoleEmphasis, TerminalRunRoleProse}; !reflect.DeepEqual(got, want) {
		t.Fatalf("runRoles(first prose block) = %v, want %v", got, want)
	}
	if got := compactTerminalWhitespace(plan.Blocks[1].Runs[0].Text); got != "ejemplo real" {
		t.Fatalf("example text = %q, want %q", got, "ejemplo real")
	}
	if got := compactTerminalWhitespace(plan.Blocks[0].Runs[0].Text + plan.Blocks[0].Runs[1].Text + plan.Blocks[0].Runs[2].Text); got == compactTerminalWhitespace(paragraph.Markdown) {
		t.Fatal("planner fell back to markdown projection instead of typed inlines")
	}
}

func TestPlanTerminalParagraphKeepsProseEmphasisAndExampleDistinct(t *testing.T) {
	tests := []struct {
		name          string
		paragraph     model.Paragraph
		wantBlockKind []TerminalBlockKind
		wantRunRoles  [][]TerminalRunRole
	}{
		{
			name: "mixed paragraph yields prose example prose blocks",
			paragraph: model.Paragraph{Inlines: []model.Inline{
				{Kind: model.InlineKindText, Text: "El comparativo es "},
				{Kind: model.InlineKindMention, Text: "mejor"},
				{Kind: model.InlineKindText, Text: ". "},
				{Kind: model.InlineKindExample, Text: "Cierra bien la ventana"},
				{Kind: model.InlineKindText, Text: " Ver "},
				{Kind: model.InlineKindReference, Text: "6", Target: "bien#6"},
				{Kind: model.InlineKindText, Text: "."},
			}},
			wantBlockKind: []TerminalBlockKind{TerminalBlockKindProse, TerminalBlockKindExample, TerminalBlockKindProse},
			wantRunRoles:  [][]TerminalRunRole{{TerminalRunRoleProse, TerminalRunRoleEmphasis, TerminalRunRoleProse}, {TerminalRunRoleProse}, {TerminalRunRoleProse, TerminalRunRoleReference, TerminalRunRoleProse}},
		},
		{
			name:          "plain markdown fallback stays prose only",
			paragraph:     model.Paragraph{Markdown: "Texto *markdown* sin inlines"},
			wantBlockKind: []TerminalBlockKind{TerminalBlockKindProse},
			wantRunRoles:  [][]TerminalRunRole{{TerminalRunRoleProse}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := planTerminalParagraph(tt.paragraph)
			if got := blockKinds(plan); !reflect.DeepEqual(got, tt.wantBlockKind) {
				t.Fatalf("blockKinds = %v, want %v", got, tt.wantBlockKind)
			}
			if len(plan.Blocks) != len(tt.wantRunRoles) {
				t.Fatalf("blocks = %d, want %d", len(plan.Blocks), len(tt.wantRunRoles))
			}
			for i, block := range plan.Blocks {
				if got := runRoles(block); !reflect.DeepEqual(got, tt.wantRunRoles[i]) {
					t.Fatalf("runRoles(block %d) = %v, want %v", i, got, tt.wantRunRoles[i])
				}
			}
		})
	}
}

func TestPlanTerminalBlocksIncludesTableBlocks(t *testing.T) {
	blocks := planTerminalBlocks([]model.Block{
		{Kind: model.ArticleBlockKindParagraph, Paragraph: &model.Paragraph{Inlines: []model.Inline{{Kind: model.InlineKindText, Text: "Antes"}}}},
		{Kind: model.ArticleBlockKindTable, Table: &model.Table{Headers: []model.TableRow{{Cells: []model.TableCell{{Text: "A"}, {Text: "B"}}}}, Rows: []model.TableRow{{Cells: []model.TableCell{{Text: "1"}, {Text: "2"}}}}}},
	})

	if got, want := []TerminalBlockKind{blocks[0].Kind, blocks[1].Kind}, []TerminalBlockKind{TerminalBlockKindProse, TerminalBlockKindTable}; !reflect.DeepEqual(got, want) {
		t.Fatalf("block kinds = %v, want %v", got, want)
	}
	if len(blocks[1].Lines) == 0 || blocks[1].Lines[0] != "| A | B |" {
		t.Fatalf("table lines = %#v", blocks[1].Lines)
	}
}

func blockKinds(paragraph TerminalParagraph) []TerminalBlockKind {
	result := make([]TerminalBlockKind, 0, len(paragraph.Blocks))
	for _, block := range paragraph.Blocks {
		result = append(result, block.Kind)
	}
	return result
}

func runRoles(block TerminalBlock) []TerminalRunRole {
	result := make([]TerminalRunRole, 0, len(block.Runs))
	for _, run := range block.Runs {
		result = append(result, run.Role)
	}
	return result
}
