package render

import "strings"

const (
	defaultExampleIndent = "    "
)

// TerminalProfile configures terminal output capabilities such as ANSI support and indentation.
type TerminalProfile struct {
	ANSIEnabled   bool
	ExampleIndent string
}

// DefaultTerminalProfile returns a TerminalProfile with ANSI disabled and default indentation.
func DefaultTerminalProfile() TerminalProfile {
	return TerminalProfile{ANSIEnabled: false, ExampleIndent: defaultExampleIndent}
}

func normalizeTerminalProfile(profile TerminalProfile) TerminalProfile {
	if strings.TrimSpace(profile.ExampleIndent) == "" {
		profile.ExampleIndent = defaultExampleIndent
	}

	return profile
}
