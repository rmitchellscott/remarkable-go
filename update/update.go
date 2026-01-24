package update

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/rmitchellscott/remarkable-go/executor"
)

type InstallResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Manager interface {
	Install(ctx context.Context, swuPath string, output io.Writer) (*InstallResult, error)
}

type manager struct {
	exec executor.Executor
}

func NewManager(exec executor.Executor) Manager {
	return &manager{exec: exec}
}

func (m *manager) Install(ctx context.Context, swuPath string, output io.Writer) (*InstallResult, error) {
	streamExec, ok := m.exec.(executor.StreamingExecutor)
	if !ok {
		result, err := m.exec.Run(ctx, "swupdate-from-image-file", swuPath)
		if err != nil {
			return nil, fmt.Errorf("failed to run swupdate-from-image-file: %w", err)
		}
		if output != nil {
			io.WriteString(output, result.Stdout)
			io.WriteString(output, result.Stderr)
		}
		if !result.Success() {
			return &InstallResult{
				Success: false,
				Message: strings.TrimSpace(result.Stderr),
			}, nil
		}
		return &InstallResult{
			Success: true,
			Message: "Update installed successfully",
		}, nil
	}

	err := streamExec.RunStreaming(ctx, "swupdate-from-image-file", output, output, swuPath)
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &InstallResult{
		Success: true,
		Message: "Update installed successfully",
	}, nil
}
