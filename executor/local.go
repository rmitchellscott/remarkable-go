package executor

import (
	"bytes"
	"context"
	"io"
	"os/exec"
)

type Local struct{}

func NewLocal() *Local {
	return &Local{}
}

func (e *Local) Run(ctx context.Context, cmd string, args ...string) (*Result, error) {
	c := exec.CommandContext(ctx, cmd, args...)

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return nil, err
	}

	return result, nil
}

func (e *Local) RunShell(ctx context.Context, script string) (*Result, error) {
	return e.Run(ctx, "sh", "-c", script)
}

func (e *Local) RunStreaming(ctx context.Context, cmd string, stdout, stderr io.Writer, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}
