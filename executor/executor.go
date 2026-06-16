// Package executor runs commands on a reMarkable device over SSH, on the host,
// or against an in-memory dry-run.
package executor

import (
	"context"
	"io"
)

// Result is the outcome of a command: its exit code and captured output.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Success reports whether the command exited zero.
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// Executor runs commands and returns their result. A non-zero exit code is
// reported in Result.ExitCode with a nil error; a non-nil error means the
// command could not be run at all (e.g. transport or session failure, or a
// cancelled context).
type Executor interface {
	Run(ctx context.Context, cmd string, args ...string) (*Result, error)
	RunShell(ctx context.Context, script string) (*Result, error)
}

// StreamingExecutor additionally streams command output to writers as it is produced.
type StreamingExecutor interface {
	Executor
	RunStreaming(ctx context.Context, cmd string, stdout, stderr io.Writer, args ...string) error
}
