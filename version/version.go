package version

import (
	"strconv"
	"strings"
)

type Version string

const (
	V3_18 Version = "3.18"
	V3_22 Version = "3.22"
)

func (v Version) String() string {
	return string(v)
}

func (v Version) Compare(other Version) int {
	return Compare(string(v), string(other))
}

func (v Version) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

func (v Version) GreaterOrEqual(other Version) bool {
	return v.Compare(other) >= 0
}

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
