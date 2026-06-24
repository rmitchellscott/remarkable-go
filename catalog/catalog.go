// Package catalog loads the precomputed image manifest (images.json) and
// indexes it by (platform, version), tracking the latest version per platform.
package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rmitchellscott/remarkable-go/version"
)

// Entry is one image in the manifest. Hashes are base64-encoded raw digests,
// the form update_engine expects.
type Entry struct {
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	SHA1     string `json:"sha1"`
	SHA256   string `json:"sha256"`
}

// Catalog is an indexed, read-only view of the manifest.
type Catalog struct {
	byVersion map[string]Entry // platform + "\x00" + version
	latest    map[string]Entry // platform -> highest version
}

// Load reads a manifest from a local file path or an http(s) URL.
func Load(source string) (*Catalog, error) {
	data, err := readSource(source)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", source, err)
	}
	return New(entries)
}

// New indexes a set of entries.
func New(entries []Entry) (*Catalog, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("manifest is empty")
	}
	c := &Catalog{
		byVersion: make(map[string]Entry, len(entries)),
		latest:    make(map[string]Entry),
	}
	for _, e := range entries {
		if e.Platform == "" || e.Version == "" || e.Key == "" {
			return nil, fmt.Errorf("manifest entry missing platform/version/key: %+v", e)
		}
		c.byVersion[key(e.Platform, e.Version)] = e
		if cur, ok := c.latest[e.Platform]; !ok || version.Compare(e.Version, cur.Version) > 0 {
			c.latest[e.Platform] = e
		}
	}
	return c, nil
}

// Lookup returns the image to offer a device. An empty or "latest" version
// selects the highest version available for the platform.
func (c *Catalog) Lookup(platform, ver string) (Entry, bool) {
	if ver == "" || ver == "latest" {
		e, ok := c.latest[platform]
		return e, ok
	}
	e, ok := c.byVersion[key(platform, ver)]
	return e, ok
}

// Entries returns all images sorted by platform then descending version.
func (c *Catalog) Entries() []Entry {
	out := make([]Entry, 0, len(c.byVersion))
	for _, e := range c.byVersion {
		out = append(out, e)
	}
	sortEntries(out)
	return out
}

// Versions returns the images for a single platform, sorted descending by
// version (newest first). Useful for populating a UI list.
func (c *Catalog) Versions(platform string) []Entry {
	var out []Entry
	for _, e := range c.byVersion {
		if e.Platform == platform {
			out = append(out, e)
		}
	}
	sortEntries(out)
	return out
}

func key(platform, ver string) string { return platform + "\x00" + ver }

func readSource(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(source)
		if err != nil {
			return nil, fmt.Errorf("fetch manifest %s: %w", source, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch manifest %s: status %d", source, resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}

func sortEntries(e []Entry) {
	for i := 1; i < len(e); i++ {
		for j := i; j > 0 && less(e[j], e[j-1]); j-- {
			e[j], e[j-1] = e[j-1], e[j]
		}
	}
}

func less(a, b Entry) bool {
	if a.Platform != b.Platform {
		return a.Platform < b.Platform
	}
	return version.Compare(a.Version, b.Version) > 0
}
