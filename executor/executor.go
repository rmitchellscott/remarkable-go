package executor

import (
	"context"
	"io"
)

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func (r *Result) Success() bool {
	return r.ExitCode == 0
}

type Executor interface {
	Run(ctx context.Context, cmd string, args ...string) (*Result, error)
	RunShell(ctx context.Context, script string) (*Result, error)
}

type StreamingExecutor interface {
	Executor
	RunStreaming(ctx context.Context, cmd string, stdout, stderr io.Writer, args ...string) error
}
