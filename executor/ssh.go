package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const sbinPath = "/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin"

// SSH is an Executor that runs commands on a device over an SSH connection.
type SSH struct {
	client *ssh.Client
	logger *slog.Logger
}

// Option configures an SSH executor.
type Option func(*SSH)

// WithLogger logs each command, its exit code, and stderr at slog Debug level.
// By default an SSH executor logs nothing.
func WithLogger(logger *slog.Logger) Option {
	return func(e *SSH) {
		if logger != nil {
			e.logger = logger
		}
	}
}

// NewSSH returns an Executor that runs commands over the given SSH client.
func NewSSH(client *ssh.Client, opts ...Option) *SSH {
	e := &SSH{client: client, logger: slog.New(slog.DiscardHandler)}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Run executes cmd (with args) in a new session under a PATH that includes the
// standard sbin directories. Per the Executor contract, a non-zero exit is
// returned via Result.ExitCode with a nil error; a non-nil error means the
// command could not be run or ctx was cancelled.
func (e *SSH) Run(ctx context.Context, cmd string, args ...string) (*Result, error) {
	session, err := e.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	go func() {
		<-ctx.Done()
		session.Close()
	}()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	displayCmd := cmd
	if len(args) > 0 {
		displayCmd = cmd + " " + shellJoin(args)
	}
	fullCmd := "PATH=" + sbinPath + ":$PATH " + displayCmd

	start := time.Now()
	err = session.Run(fullCmd)
	dur := time.Since(start)

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
			e.logger.Debug("exec", "cmd", displayCmd, "exit", result.ExitCode, "stderr", strings.TrimSpace(result.Stderr), "dur", dur)
			return result, nil
		}
		if ctx.Err() != nil {
			e.logger.Debug("exec cancelled", "cmd", displayCmd, "err", ctx.Err())
			return nil, ctx.Err()
		}
		e.logger.Debug("exec failed", "cmd", displayCmd, "err", err)
		return nil, err
	}

	e.logger.Debug("exec", "cmd", displayCmd, "exit", 0, "dur", dur)
	return result, nil
}

// RunShell runs script via sh -c.
func (e *SSH) RunShell(ctx context.Context, script string) (*Result, error) {
	return e.Run(ctx, "sh", "-c", script)
}

// RunStreaming runs cmd (with args) and streams stdout and stderr to the given writers.
func (e *SSH) RunStreaming(ctx context.Context, cmd string, stdout, stderr io.Writer, args ...string) error {
	session, err := e.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	go func() {
		<-ctx.Done()
		session.Close()
	}()

	session.Stdout = stdout
	session.Stderr = stderr

	displayCmd := cmd
	if len(args) > 0 {
		displayCmd = cmd + " " + shellJoin(args)
	}
	fullCmd := "PATH=" + sbinPath + ":$PATH " + displayCmd

	e.logger.Debug("exec stream", "cmd", displayCmd)
	err = session.Run(fullCmd)
	if err != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		e.logger.Debug("exec stream failed", "cmd", displayCmd, "err", err)
	}
	return err
}

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t\n\"'\\$`!#&|;(){}[]<>?*~") {
			quoted[i] = "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
		} else {
			quoted[i] = arg
		}
	}
	return strings.Join(quoted, " ")
}
