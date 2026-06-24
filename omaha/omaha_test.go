package omaha

import (
	"strings"
	"testing"
)

const updateCheckBody = `<?xml version="1.0" encoding="UTF-8"?>
<request protocol="3.0" version="3.2.3.1595" requestid="{req}" sessionid="{ses}" updaterversion="0.4.2" installsource="scheduler" ismachine="1">
	<os version="zg" platform="reMarkable2" sp="3.2.3.1595_armv7l" arch="armv7l"/>
	<app appid="{98DA7DF2-4E3E-4744-9DE6-EC931886ABAB}" version="3.2.3.1595" track="Prod" ap="Prod" bootid="{boot}" oem="RM100-753-12345" lang="en-US" delta_okay="false">
		<updatecheck/>
	</app>
</request>`

const eventBody = `<?xml version="1.0" encoding="UTF-8"?>
<request protocol="3.0" version="3.2.3.1595">
	<os version="zg" platform="reMarkable" arch="armv7l"/>
	<app appid="{98DA7DF2-4E3E-4744-9DE6-EC931886ABAB}" version="3.2.3.1595">
		<event eventtype="14" eventresult="1" errorcode="0"/>
	</app>
</request>`

func TestParseUpdateCheck(t *testing.T) {
	r, err := ParseRequest([]byte(updateCheckBody))
	if err != nil {
		t.Fatalf("ParseRequest: %v", err)
	}
	if !r.IsUpdateCheck() {
		t.Fatal("expected update-check request")
	}
	if r.App.Event != nil {
		t.Fatal("did not expect an event node")
	}
	if got := r.Platform(); got != "reMarkable2" {
		t.Fatalf("platform = %q, want reMarkable2", got)
	}
	if got := r.App.Version; got != "3.2.3.1595" {
		t.Fatalf("app version = %q, want 3.2.3.1595", got)
	}
}

func TestParseEvent(t *testing.T) {
	r, err := ParseRequest([]byte(eventBody))
	if err != nil {
		t.Fatalf("ParseRequest: %v", err)
	}
	if r.IsUpdateCheck() {
		t.Fatal("did not expect update-check")
	}
	if r.App.Event == nil {
		t.Fatal("expected an event node")
	}
	if r.App.Event.Type != EventTypeUpdateComplete {
		t.Fatalf("event type = %d, want %d", r.App.Event.Type, EventTypeUpdateComplete)
	}
	if r.Platform() != "reMarkable" {
		t.Fatalf("platform = %q, want reMarkable", r.Platform())
	}
}

func TestBuildUpdateResponse(t *testing.T) {
	resp := BuildUpdateResponse(Update{
		Version:  "3.11.2.5",
		Name:     "3.11.2.5_reMarkable2-qLFGoqPtPL.signed",
		Size:     87822626,
		SHA1:     "c2hhMQ==",
		SHA256:   "c2hhMjU2",
		Codebase: "https://bucket.example.com/legacy/",
	})
	for _, want := range []string{
		`codebase="https://bucket.example.com/legacy/"`,
		`<manifest version="3.11.2.5">`,
		`hash="c2hhMQ=="`,
		`name="3.11.2.5_reMarkable2-qLFGoqPtPL.signed"`,
		`size="87822626"`,
		`sha256="c2hhMjU2"`,
		`event="postinstall"`,
	} {
		if !strings.Contains(resp, want) {
			t.Errorf("response missing %q\n%s", want, resp)
		}
	}
}

func TestNoUpdateResponse(t *testing.T) {
	resp := NoUpdateResponse()
	if strings.Contains(resp, "updatecheck") {
		t.Errorf("no-update response should not contain an updatecheck node:\n%s", resp)
	}
	if !strings.Contains(resp, AppID) {
		t.Errorf("no-update response missing appid:\n%s", resp)
	}
}
