package render

import "strings"

const (
	defaultExampleIndent = "    "
	ansiUnderlineStart   = "\x1b[4m"
	ansiReset            = "\x1b[0m"
)

type TerminalProfile struct {
	ANSIEnabled   bool
	ExampleIndent string
}

func DefaultTerminalProfile() TerminalProfile {
	return TerminalProfile{ANSIEnabled: false, ExampleIndent: defaultExampleIndent}
}

func normalizeTerminalProfile(profile TerminalProfile) TerminalProfile {
	if strings.TrimSpace(profile.ExampleIndent) == "" {
		profile.ExampleIndent = defaultExampleIndent
	}

	return profile
}

func (p TerminalProfile) formatRun(run TerminalRun) string {
	text := run.Text
	switch run.Role {
	case TerminalRunRoleEmphasis:
		text = compactTerminalWhitespace(text)
		if text == "" {
			return ""
		}
		if p.ANSIEnabled {
			return ansiUnderlineStart + text + ansiReset
		}
		return text
	case TerminalRunRoleReference:
		text = compactTerminalWhitespace(text)
		if text == "" {
			return ""
		}
		return "→ " + text
	default:
		return text
	}
}

func (p TerminalProfile) formatExample(text string) string {
	text = compactTerminalWhitespace(text)
	if text == "" {
		return ""
	}
	return text
}
