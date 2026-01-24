package filesystem

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSH struct {
	client     *ssh.Client
	sftpClient *sftp.Client
}

func NewSSH(client *ssh.Client) (*SSH, error) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	return &SSH{client: client, sftpClient: sftpClient}, nil
}

func (f *SSH) Close() error {
	return f.sftpClient.Close()
}

func (f *SSH) ReadFile(path string) ([]byte, error) {
	file, err := f.sftpClient.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func (f *SSH) WriteFile(path string, data []byte, perm os.FileMode) error {
	file, err := f.sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Chmod(perm)
}

func (f *SSH) Exists(path string) (bool, error) {
	_, err := f.sftpClient.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (f *SSH) MkdirAll(path string, perm os.FileMode) error {
	return f.sftpClient.MkdirAll(path)
}

func (f *SSH) Remove(path string) error {
	return f.sftpClient.Remove(path)
}

func (f *SSH) RemoveAll(path string) error {
	return f.removeRecursive(path)
}

func (f *SSH) removeRecursive(path string) error {
	info, err := f.sftpClient.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return f.sftpClient.Remove(path)
	}

	entries, err := f.sftpClient.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		childPath := path + "/" + entry.Name()
		if err := f.removeRecursive(childPath); err != nil {
			return err
		}
	}

	return f.sftpClient.RemoveDirectory(path)
}

func (f *SSH) Mount(device, target string, readonly bool) error {
	cmd := "mount"
	if readonly {
		cmd += " -o ro"
	}
	cmd += " " + device + " " + target
	return f.runSSHCommand(cmd)
}

func (f *SSH) Unmount(target string) error {
	return f.runSSHCommand("umount " + target)
}

func (f *SSH) IsMounted(target string) (bool, error) {
	data, err := f.ReadFile("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("failed to read /proc/mounts: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == target {
			return true, nil
		}
	}
	return false, nil
}

func (f *SSH) runSSHCommand(cmd string) error {
	session, err := f.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()
	return session.Run(cmd)
}
