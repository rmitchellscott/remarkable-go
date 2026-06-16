// Package version compares reMarkable OS version strings and reads them from a device.
package version

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/rmitchellscott/remarkable-go/filesystem"
)

// Version is a dotted reMarkable OS version string, e.g. "3.11.2.5".
type Version string

// Notable version thresholds.
const (
	V3_18 Version = "3.18"
	V3_22 Version = "3.22"
)

func (v Version) String() string {
	return string(v)
}

// Compare reports whether v is less than (-1), equal to (0), or greater than (1) other.
func (v Version) Compare(other Version) int {
	return Compare(string(v), string(other))
}

func (v Version) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

func (v Version) GreaterOrEqual(other Version) bool {
	return v.Compare(other) >= 0
}

// Compare numerically compares two dotted version strings, returning -1, 0, or 1.
// Missing trailing components are treated as zero, so "3.11" equals "3.11.0.0".
func Compare(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}
	return 0
}

func Parse(s string) Version {
	return Version(s)
}

var versionFiles = []struct {
	path   string
	prefix string
}{
	{"/usr/share/remarkable/update.conf", "RELEASE_VERSION="},
	{"/etc/os-release", "IMG_VERSION="},
}

// ReadFromFS reads the OS version from the standard release files under root
// (pass "" for the live root). It returns an error if no version file is found.
func ReadFromFS(fs filesystem.FS, root string) (string, error) {
	for _, vf := range versionFiles {
		data, err := fs.ReadFile(root + vf.path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if idx := strings.Index(line, vf.prefix); idx != -1 {
				return strings.Trim(line[idx+len(vf.prefix):], `"`), nil
			}
		}
	}
	return "", fmt.Errorf("unable to detect OS version")
}
