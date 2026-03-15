package platform

import "io"

// OSCLI is the production implementation of CLI backed by real OS streams.
type OSCLI struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

// NewOSCLI constructs an OSCLI with the given arguments and output writers.
func NewOSCLI(args []string, stdout io.Writer, stderr io.Writer) *OSCLI {
	return &OSCLI{
		args:   args,
		stdout: stdout,
		stderr: stderr,
	}
}

// Args returns the command-line arguments passed to the process.
func (c *OSCLI) Args() []string {
	return c.args
}

// Stdout returns the writer for standard output.
func (c *OSCLI) Stdout() io.Writer {
	return c.stdout
}

// Stderr returns the writer for standard error.
func (c *OSCLI) Stderr() io.Writer {
	return c.stderr
}
