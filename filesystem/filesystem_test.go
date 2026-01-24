package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocal_ReadWriteFile(t *testing.T) {
	f := NewLocal()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")

	if err := f.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := f.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("ReadFile = %q, want %q", data, content)
	}
}

func TestLocal_Exists(t *testing.T) {
	f := NewLocal()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "exists.txt")

	exists, err := f.Exists(path)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("File should not exist")
	}

	os.WriteFile(path, []byte("test"), 0644)

	exists, err = f.Exists(path)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("File should exist")
	}
}

func TestMock_ReadWriteFile(t *testing.T) {
	f := NewMock()

	_, err := f.ReadFile("/nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	f.SetFileString("/test.txt", "hello")

	data, err := f.ReadFile("/test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("ReadFile = %q, want %q", data, "hello")
	}

	if err := f.WriteFile("/new.txt", []byte("world"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err = f.ReadFile("/new.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("ReadFile = %q, want %q", data, "world")
	}
}

func TestMock_Exists(t *testing.T) {
	f := NewMock()

	exists, _ := f.Exists("/test")
	if exists {
		t.Error("Should not exist")
	}

	f.SetFileString("/test", "data")

	exists, _ = f.Exists("/test")
	if !exists {
		t.Error("Should exist")
	}
}

func TestMock_Mount(t *testing.T) {
	f := NewMock()

	mounted, _ := f.IsMounted("/mnt/test")
	if mounted {
		t.Error("Should not be mounted")
	}

	f.Mount("/dev/sda1", "/mnt/test", true)

	mounted, _ = f.IsMounted("/mnt/test")
	if !mounted {
		t.Error("Should be mounted")
	}

	f.Unmount("/mnt/test")

	mounted, _ = f.IsMounted("/mnt/test")
	if mounted {
		t.Error("Should not be mounted after unmount")
	}
}

func TestMock_SimulateMountedPartition(t *testing.T) {
	f := NewMock()
	f.SimulateMountedPartition(2, "3.20.0.92")

	mounted, _ := f.IsMounted("/tmp/mount_p2")
	if !mounted {
		t.Error("Partition should be mounted")
	}

	data, err := f.ReadFile("/tmp/mount_p2/usr/share/remarkable/update.conf")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "RELEASE_VERSION=3.20.0.92\n" {
		t.Errorf("Unexpected content: %q", data)
	}
}
