package platform

import "io"

type CLI interface {
	Args() []string
	Stdout() io.Writer
	Stderr() io.Writer
}
