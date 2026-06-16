// Package filesystem abstracts file access on a reMarkable device (over SFTP),
// on the host, or against an in-memory mock.
package filesystem

import "os"

// FS provides file operations on a target filesystem.
type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Exists(path string) (bool, error)
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	RemoveAll(path string) error
}

// MountableFS is an FS that can also mount and unmount block devices.
type MountableFS interface {
	FS
	Mount(device, target string, readonly bool) error
	Unmount(target string) error
	IsMounted(target string) (bool, error)
}
