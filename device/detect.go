package device

import (
	"context"
	"fmt"
	"strings"

	"github.com/rmitchellscott/remarkable-go/executor"
	"github.com/rmitchellscott/remarkable-go/filesystem"
	"github.com/rmitchellscott/remarkable-go/version"
)

// Detect identifies the device model from /sys/devices/soc0/machine. It returns
// Unknown with a nil error when the model is unrecognized, and a non-nil error
// only when the model file cannot be read.
func Detect(fs filesystem.FS) (Type, error) {
	data, err := fs.ReadFile("/sys/devices/soc0/machine")
	if err != nil {
		return Unknown, err
	}

	model := string(data)

	if strings.Contains(model, "Ferrari") {
		return RMPP, nil
	}
	if strings.Contains(model, "Chiappa") {
		return RMPPMove, nil
	}
	if strings.Contains(model, "Tatsu") {
		return RMPPure, nil
	}

	if strings.Contains(model, "reMarkable 2") {
		return RM2, nil
	}
	if strings.Contains(model, "reMarkable") {
		return RM1, nil
	}

	return Unknown, nil
}

// DetectVersion returns the running OS version read from the device's release files.
func DetectVersion(fs filesystem.FS) (string, error) {
	return version.ReadFromFS(fs, "")
}

// DetectArchitecture returns the device CPU architecture via uname -m.
func DetectArchitecture(exec executor.Executor) (Architecture, error) {
	result, err := exec.Run(context.Background(), "uname", "-m")
	if err != nil {
		return "", err
	}
	if !result.Success() {
		return "", fmt.Errorf("uname -m failed: %s", result.Stderr)
	}

	machine := strings.TrimSpace(result.Stdout)

	switch machine {
	case "aarch64", "arm64":
		return Aarch64, nil
	case "armv7l", "armv6l", "arm":
		return Arm32, nil
	default:
		return "", fmt.Errorf("unknown architecture: %s", machine)
	}
}
