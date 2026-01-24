package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SSH struct {
	client *ssh.Client
}

func NewSSH(client *ssh.Client) *SSH {
	return &SSH{client: client}
}

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

	fullCmd := cmd
	if len(args) > 0 {
		fullCmd = cmd + " " + shellJoin(args)
	}

	err = session.Run(fullCmd)

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
			return result, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return result, nil
}

func (e *SSH) RunShell(ctx context.Context, script string) (*Result, error) {
	return e.Run(ctx, "sh", "-c", script)
}

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

	fullCmd := cmd
	if len(args) > 0 {
		fullCmd = cmd + " " + shellJoin(args)
	}

	err = session.Run(fullCmd)
	if err != nil && ctx.Err() != nil {
		return ctx.Err()
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
