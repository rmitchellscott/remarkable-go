package executor

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type DryRun struct {
	responses map[string]*Result
	log       []string
}

func NewDryRun() *DryRun {
	return &DryRun{
		responses: make(map[string]*Result),
	}
}

func (e *DryRun) SetResponse(cmdPrefix string, result *Result) {
	e.responses[cmdPrefix] = result
}

func (e *DryRun) Log() []string {
	return e.log
}

func (e *DryRun) Run(ctx context.Context, cmd string, args ...string) (*Result, error) {
	fullCmd := cmd
	if len(args) > 0 {
		fullCmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	e.log = append(e.log, fmt.Sprintf("[DRY RUN] %s", fullCmd))

	for prefix, result := range e.responses {
		if strings.HasPrefix(fullCmd, prefix) {
			return result, nil
		}
	}

	return &Result{
		ExitCode: 0,
		Stdout:   fmt.Sprintf("[DRY RUN] %s", fullCmd),
	}, nil
}

func (e *DryRun) RunShell(ctx context.Context, script string) (*Result, error) {
	e.log = append(e.log, fmt.Sprintf("[DRY RUN] sh -c %q", script))

	for prefix, result := range e.responses {
		if strings.Contains(script, prefix) {
			return result, nil
		}
	}

	return &Result{
		ExitCode: 0,
		Stdout:   fmt.Sprintf("[DRY RUN] %s", script),
	}, nil
}

func (e *DryRun) RunStreaming(ctx context.Context, cmd string, stdout, stderr io.Writer, args ...string) error {
	result, err := e.Run(ctx, cmd, args...)
	if err != nil {
		return err
	}
	if stdout != nil {
		stdout.Write([]byte(result.Stdout))
	}
	if stderr != nil {
		stderr.Write([]byte(result.Stderr))
	}
	return nil
}
