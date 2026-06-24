// Package omaha implements the subset of the Chrome OS update_engine "Omaha"
// v3.0 protocol that reMarkable's legacy updater speaks. The tablet POSTs an
// update-check (or event) request and the server replies with a manifest that
// points at the signed payload.
package omaha

import (
	"encoding/xml"
	"fmt"
)

// AppID is the constant reMarkable application id the tablet reports.
const AppID = "{98DA7DF2-4E3E-4744-9DE6-EC931886ABAB}"

// EventTypeUpdateComplete is the event type the device reports once an update
// has been successfully applied.
const EventTypeUpdateComplete = 14

// Request is the subset of the update_engine request we care about: the device
// platform, and whether the body is an update-check or an event report.
type Request struct {
	XMLName xml.Name `xml:"request"`
	OS      struct {
		Platform string `xml:"platform,attr"`
		Version  string `xml:"version,attr"`
	} `xml:"os"`
	App struct {
		Version     string `xml:"version,attr"`
		UpdateCheck *node  `xml:"updatecheck"`
		Event       *Event `xml:"event"`
	} `xml:"app"`
}

type node struct{}

// Event is a status report the device sends during/after an update.
type Event struct {
	Type      int    `xml:"eventtype,attr"`
	Result    int    `xml:"eventresult,attr"`
	ErrorCode string `xml:"errorcode,attr"`
}

// ParseRequest decodes an Omaha request body.
func ParseRequest(body []byte) (*Request, error) {
	var r Request
	if err := xml.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("parse omaha request: %w", err)
	}
	return &r, nil
}

// IsUpdateCheck reports whether the request is an update-check (vs an event report).
func (r *Request) IsUpdateCheck() bool { return r.App.UpdateCheck != nil }

// Platform returns the device platform string ("reMarkable" / "reMarkable2").
func (r *Request) Platform() string { return r.OS.Platform }

// Update describes the payload offered in an update response. Hashes are the
// base64-encoded raw digests update_engine expects.
type Update struct {
	Version  string
	Name     string // package name (object key) appended to Codebase
	Size     int64
	SHA1     string // base64
	SHA256   string // base64
	Codebase string // download base URL, must end with "/"
}

// BuildUpdateResponse renders an Omaha response offering the given update. The
// XML mirrors codexctl's known-working server response so update_engine accepts
// it unchanged.
func BuildUpdateResponse(u Update) string {
	return fmt.Sprintf(updateTemplate, u.Codebase, u.Version, u.SHA1, u.Name, u.Size, u.SHA256)
}

// NoUpdateResponse renders the "no update available / OK" response.
func NoUpdateResponse() string { return noUpdateResponse }

const updateTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<response protocol="3.0" server="prod">
	<daystart elapsed_seconds="37145" elapsed_days="5179"/>
	<app status="ok" appid="{98DA7DF2-4E3E-4744-9DE6-EC931886ABAB}">
		<event status="ok"/>
		<updatecheck status="ok">
			<urls>
				<url codebase="%[1]s"/>
			</urls>
			<manifest version="%[2]s">
				<packages>
					<package required="true" hash="%[3]s" name="%[4]s" size="%[5]d"/>
				</packages>
				<actions>
					<action successaction="default" sha256="%[6]s" event="postinstall" DisablePayloadBackoff="true"/>
				</actions>
			</manifest>
		</updatecheck>
		<ping status="ok"/>
	</app>
</response>
`

const noUpdateResponse = `<?xml version="1.0" encoding="UTF-8"?>
<response protocol="3.0" server="prod">
	<daystart elapsed_seconds="42548" elapsed_days="5179"/>
	<app status="ok" appid="{98DA7DF2-4E3E-4744-9DE6-EC931886ABAB}">
		<event status="ok"/>
		<ping status="ok"/>
	</app>
</response>
`
