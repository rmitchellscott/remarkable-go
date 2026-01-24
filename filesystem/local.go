package filesystem

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Local struct{}

func NewLocal() *Local {
	return &Local{}
}

func (f *Local) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *Local) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (f *Local) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (f *Local) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *Local) Remove(path string) error {
	return os.Remove(path)
}

func (f *Local) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (f *Local) Mount(device, target string, readonly bool) error {
	args := []string{}
	if readonly {
		args = append(args, "-o", "ro")
	}
	args = append(args, device, target)

	return exec.CommandContext(context.Background(), "mount", args...).Run()
}

func (f *Local) Unmount(target string) error {
	return exec.CommandContext(context.Background(), "umount", target).Run()
}

func (f *Local) IsMounted(target string) (bool, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("failed to open /proc/mounts: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == target {
			return true, nil
		}
	}

	return false, scanner.Err()
}
