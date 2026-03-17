// Script to fetch DPD articles and analyze sign usage
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Disble/dlexa/internal/fetch"
	"github.com/Disble/dlexa/internal/model"
)

func main() {
	words := []string{
		"abrogar",    // []
		"abstraer",   // [...]
		"abuelo",     // /
		"acertar",    // +
		"adherir",    // ―
		"alférez",    // ^2
		"alícuota",   // otros signos
		"androfobia", // otros signos
	}

	baseURL := "https://www.rae.es/dpd"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	timeout := 15 * time.Second

	fetcher := fetch.NewDPDFetcher(baseURL, timeout, userAgent)
	ctx := context.Background()

	outputDir := filepath.Join("testdata", "dpd-signs-analysis")
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	for _, word := range words {
		fmt.Printf("Fetching: %s...\n", word)

		req := fetch.Request{
			Query: word,
			Source: model.SourceDescriptor{
				Name: "dpd",
			},
		}

		doc, err := fetcher.Fetch(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", word, err)
			continue
		}

		// Save HTML
		htmlPath := filepath.Join(outputDir, word+".html")
		if err := os.WriteFile(htmlPath, doc.Body, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", htmlPath, err)
			continue
		}

		// Extract article content only (within <entry> tags)
		body := string(doc.Body)
		articleContent := extractArticleContent(body)

		// Analyze signs in article content only
		signs := analyzeSignsInHTML(articleContent, word)

		// Print analysis
		fmt.Printf("  URL: %s\n", doc.URL)
		fmt.Printf("  Status: %d\n", doc.StatusCode)
		if len(signs) > 0 {
			fmt.Printf("  Signs found:\n")
			for _, s := range signs {
				fmt.Printf("    %s\n", s)
			}
		} else {
			fmt.Printf("  No special signs detected\n")
		}
		fmt.Println()

		// Be nice to the server
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("HTML files saved to: %s\n", outputDir)
}

func extractArticleContent(html string) string {
	// Find content within <entry> tags
	startTag := "<entry"
	endTag := "</entry>"

	startIdx := strings.Index(html, startTag)
	if startIdx == -1 {
		return html // No entry tag, return full HTML
	}

	endIdx := strings.Index(html[startIdx:], endTag)
	if endIdx == -1 {
		return html // No closing tag, return full HTML
	}

	// Extract from opening tag to closing tag (inclusive)
	return html[startIdx : startIdx+endIdx+len(endTag)]
}

func analyzeSignsInHTML(html, _ string) []string {
	var findings []string

	// Check for specific signs
	signs := map[string]string{
		"⊗":   "exclusion marker",
		"*":   "asterisk (agrammatical)",
		"‖":   "double bar (hypothetical)",
		">":   "greater than (transformation)",
		"<":   "less than (etymology)",
		"→":   "arrow (cross-reference)",
		"//":  "double slash (phoneme)",
		"[…]": "ellipsis brackets",
		"[":   "left bracket",
		"]":   "right bracket",
		"/":   "slash",
		"@":   "at sign (digital edition)",
		"^":   "caret (superscript indicator)",
		"―":   "em dash",
		"+":   "plus sign",
	}

	for sign, desc := range signs {
		if strings.Contains(html, sign) {
			// Try to find context
			context := extractSignContext(html, sign)
			findings = append(findings, fmt.Sprintf("%s (%s): %s", sign, desc, context))
		}
	}

	return findings
}

func extractSignContext(html, sign string) string {
	idx := strings.Index(html, sign)
	if idx == -1 {
		return "not found"
	}

	// Extract ~100 chars before and after
	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(sign) + 50
	if end > len(html) {
		end = len(html)
	}

	snippet := html[start:end]
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	snippet = strings.ReplaceAll(snippet, "\t", " ")

	// Find the containing tag
	tagStart := strings.LastIndex(snippet[:50], "<")
	tagEnd := strings.Index(snippet[50:], ">")
	if tagStart >= 0 && tagEnd >= 0 {
		tag := snippet[tagStart : 50+tagEnd+1]
		return fmt.Sprintf("in tag: %s", tag)
	}

	return snippet
}
