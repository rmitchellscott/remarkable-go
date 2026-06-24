// Package updateconf rewrites the SERVER= line in a reMarkable
// /usr/share/remarkable/update.conf so the legacy update_engine checks in with a
// chosen Omaha server. It operates on in-memory bytes (read/write the file over
// SFTP around it) and has no SSH dependency.
package updateconf

import "strings"

// SetServer returns conf with SERVER= pointed at serverURL: any existing
// SERVER= line is commented out and a new one is inserted immediately after the
// first [General] section header (appended at the end if there is no [General]).
// This mirrors codexctl's set_server_config and scripts/rm-legacy-update.sh.
func SetServer(conf []byte, serverURL string) []byte {
	serverLine := "SERVER=" + serverURL
	if len(conf) == 0 {
		return []byte(serverLine)
	}

	hadTrailingNewline := conf[len(conf)-1] == '\n'
	lines := strings.Split(string(conf), "\n")
	if hadTrailingNewline && len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	out := make([]string, 0, len(lines)+1)
	inserted := false
	for _, line := range lines {
		if strings.HasPrefix(line, "SERVER=") {
			out = append(out, "#"+line)
			continue
		}
		out = append(out, line)
		if !inserted && strings.HasPrefix(line, "[General]") {
			out = append(out, serverLine)
			inserted = true
		}
	}
	if !inserted {
		out = append(out, serverLine)
	}

	result := strings.Join(out, "\n")
	if hadTrailingNewline {
		result += "\n"
	}
	return []byte(result)
}

// CurrentServer returns the value of the first uncommented SERVER= line, or ""
// if none is present.
func CurrentServer(conf []byte) string {
	for _, line := range strings.Split(string(conf), "\n") {
		if strings.HasPrefix(line, "SERVER=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "SERVER="))
		}
	}
	return ""
}
