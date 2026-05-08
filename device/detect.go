package device

import (
	"strings"

	"github.com/rmitchellscott/remarkable-go/filesystem"
)

func Detect(fs filesystem.FS) (Type, error) {
	data, err := fs.ReadFile("/proc/device-tree/model")
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
