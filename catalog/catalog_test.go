package catalog

import "testing"

func sample() *Catalog {
	c, err := New([]Entry{
		{Version: "3.11.2.5", Platform: "reMarkable2", Key: "3.11.2.5_reMarkable2-x.signed", Size: 1, SHA1: "a", SHA256: "b"},
		{Version: "3.10.2.2063", Platform: "reMarkable2", Key: "3.10.2.2063_reMarkable2-y.signed", Size: 2, SHA1: "c", SHA256: "d"},
		{Version: "3.11.2.5", Platform: "reMarkable", Key: "3.11.2.5_reMarkable-z.signed", Size: 3, SHA1: "e", SHA256: "f"},
	})
	if err != nil {
		panic(err)
	}
	return c
}

func TestLookupExact(t *testing.T) {
	c := sample()
	e, ok := c.Lookup("reMarkable2", "3.10.2.2063")
	if !ok || e.Key != "3.10.2.2063_reMarkable2-y.signed" {
		t.Fatalf("exact lookup failed: %+v ok=%v", e, ok)
	}
}

func TestLookupLatest(t *testing.T) {
	c := sample()
	for _, v := range []string{"", "latest"} {
		e, ok := c.Lookup("reMarkable2", v)
		if !ok || e.Version != "3.11.2.5" {
			t.Fatalf("latest(%q) = %+v ok=%v, want 3.11.2.5", v, e, ok)
		}
	}
}

func TestLookupUnknown(t *testing.T) {
	c := sample()
	if _, ok := c.Lookup("reMarkable2", "9.9.9.9"); ok {
		t.Fatal("expected miss for unknown version")
	}
	if _, ok := c.Lookup("rmpp", ""); ok {
		t.Fatal("expected miss for unknown platform")
	}
}

func TestVersionsSortedDescending(t *testing.T) {
	c := sample()
	got := c.Versions("reMarkable2")
	if len(got) != 2 {
		t.Fatalf("Versions(reMarkable2) returned %d entries, want 2", len(got))
	}
	if got[0].Version != "3.11.2.5" || got[1].Version != "3.10.2.2063" {
		t.Fatalf("Versions not sorted newest-first: %+v", got)
	}
	if len(c.Versions("rmpp")) != 0 {
		t.Fatal("expected no entries for unknown platform")
	}
}

func TestNewRejectsEmpty(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("expected error for empty manifest")
	}
}
