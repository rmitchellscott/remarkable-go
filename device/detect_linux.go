package device

import (
	"fmt"
	"syscall"
)

func DetectArchitecture() (Architecture, error) {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return "", err
	}

	machine := int8ToString(uname.Machine[:])

	switch machine {
	case "aarch64", "arm64":
		return Aarch64, nil
	case "armv7l", "armv6l", "arm":
		return Arm32, nil
	default:
		return "", fmt.Errorf("unknown architecture: %s", machine)
	}
}

func int8ToString(arr []int8) string {
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}
