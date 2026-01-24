package filesystem

import (
	"fmt"
	"os"
	"strings"
)

type Mock struct {
	files   map[string][]byte
	mounts  map[string]string
	removed []string
}

func NewMock() *Mock {
	return &Mock{
		files:  make(map[string][]byte),
		mounts: make(map[string]string),
	}
}

func (f *Mock) SetFile(path string, content []byte) {
	f.files[path] = content
}

func (f *Mock) SetFileString(path, content string) {
	f.files[path] = []byte(content)
}

func (f *Mock) GetFile(path string) ([]byte, bool) {
	data, ok := f.files[path]
	return data, ok
}

func (f *Mock) Removed() []string {
	return f.removed
}

func (f *Mock) ReadFile(path string) ([]byte, error) {
	if data, ok := f.files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (f *Mock) WriteFile(path string, data []byte, perm os.FileMode) error {
	f.files[path] = data
	return nil
}

func (f *Mock) Exists(path string) (bool, error) {
	_, ok := f.files[path]
	if ok {
		return true, nil
	}
	for p := range f.files {
		if strings.HasPrefix(p, path+"/") {
			return true, nil
		}
	}
	return false, nil
}

func (f *Mock) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (f *Mock) Remove(path string) error {
	f.removed = append(f.removed, path)
	delete(f.files, path)
	return nil
}

func (f *Mock) RemoveAll(path string) error {
	f.removed = append(f.removed, path)
	for p := range f.files {
		if strings.HasPrefix(p, path) {
			delete(f.files, p)
		}
	}
	return nil
}

func (f *Mock) Mount(device, target string, readonly bool) error {
	f.mounts[target] = device
	return nil
}

func (f *Mock) Unmount(target string) error {
	delete(f.mounts, target)
	return nil
}

func (f *Mock) IsMounted(target string) (bool, error) {
	_, ok := f.mounts[target]
	return ok, nil
}

func (f *Mock) SimulateMountedPartition(partNum int, version string) {
	mountPoint := fmt.Sprintf("/tmp/mount_p%d", partNum)
	f.mounts[mountPoint] = fmt.Sprintf("/dev/mmcblk0p%d", partNum)

	f.SetFileString(
		fmt.Sprintf("%s/usr/share/remarkable/update.conf", mountPoint),
		fmt.Sprintf("RELEASE_VERSION=%s\n", version),
	)
}
