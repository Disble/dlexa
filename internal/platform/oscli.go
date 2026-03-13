package platform

import "io"

type OSCLI struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

func NewOSCLI(args []string, stdout io.Writer, stderr io.Writer) *OSCLI {
	return &OSCLI{
		args:   args,
		stdout: stdout,
		stderr: stderr,
	}
}

func (c *OSCLI) Args() []string {
	return c.args
}

func (c *OSCLI) Stdout() io.Writer {
	return c.stdout
}

func (c *OSCLI) Stderr() io.Writer {
	return c.stderr
}
