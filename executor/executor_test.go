package executor

import (
	"context"
	"testing"
)

func TestLocal_Run(t *testing.T) {
	e := NewLocal()
	result, err := e.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
}

func TestLocal_RunShell(t *testing.T) {
	e := NewLocal()
	result, err := e.RunShell(context.Background(), "echo hello && echo world")
	if err != nil {
		t.Fatalf("RunShell failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\nworld\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\nworld\n")
	}
}

func TestLocal_Run_ExitCode(t *testing.T) {
	e := NewLocal()
	result, err := e.Run(context.Background(), "sh", "-c", "exit 42")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

func TestDryRun_Run(t *testing.T) {
	e := NewDryRun()
	result, err := e.Run(context.Background(), "dangerous", "command")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}

	log := e.Log()
	if len(log) != 1 {
		t.Fatalf("Log length = %d, want 1", len(log))
	}
	if log[0] != "[DRY RUN] dangerous command" {
		t.Errorf("Log[0] = %q, want %q", log[0], "[DRY RUN] dangerous command")
	}
}

func TestDryRun_SetResponse(t *testing.T) {
	e := NewDryRun()
	e.SetResponse("rootdev", &Result{
		ExitCode: 0,
		Stdout:   "/dev/mmcblk0p3",
	})

	result, err := e.Run(context.Background(), "rootdev")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Stdout != "/dev/mmcblk0p3" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "/dev/mmcblk0p3")
	}
}

func TestResult_Success(t *testing.T) {
	r := &Result{ExitCode: 0}
	if !r.Success() {
		t.Error("ExitCode 0 should be success")
	}

	r = &Result{ExitCode: 1}
	if r.Success() {
		t.Error("ExitCode 1 should not be success")
	}
}
