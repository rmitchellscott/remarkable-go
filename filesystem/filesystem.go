package filesystem

import "os"

type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Exists(path string) (bool, error)
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	RemoveAll(path string) error
}

type MountableFS interface {
	FS
	Mount(device, target string, readonly bool) error
	Unmount(target string) error
	IsMounted(target string) (bool, error)
}
