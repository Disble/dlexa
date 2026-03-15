// Package platform provides OS-level abstractions for the CLI.
package platform

import "io"

// CLI defines the minimal interface for CLI environment access.
type CLI interface {
	Args() []string
	Stdout() io.Writer
	Stderr() io.Writer
}
