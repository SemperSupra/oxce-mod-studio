// Package adapters converts source-proven screen envelopes without inventing
// game-state, native acceptance, pixel provenance, or human approval.
package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"semper-supra/oxce-cockpit/contract"
	"strings"
)

type ScreenControl struct {
	Role  string `json:"role"`
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	W     int    `json:"w"`
	H     int    `json:"h"`
}
type ScreenClass struct {
	Event    string          `json:"event"`
	Class    string          `json:"class"`
	ScaleX   int             `json:"scaleX,omitempty"`
	ScaleY   int             `json:"scaleY,omitempty"`
	Controls []ScreenControl `json:"controls,omitempty"`
}
type Provenance struct{ EngineBuildSHA, ModManifestSHA, SessionID, ActorID string }

// ConvertLegacyScreen accepts either the older 'screen-class' schema or the
// real native PR13 2026-10-29 `screen` event with controls and scaling.
// The converted event remains an *unattested contract adapter record*.
func ConvertLegacyScreen(line []byte, p Provenance, seq uint64) (contract.Event, error) {
	var in ScreenClass
	if len(line) > contract.MaxRecordBytes {
		return contract.Event{}, errors.New("oversized native screen record")
	}
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return contract.Event{}, fmt.Errorf("decode native screen: %w", err)
	}
	var tail any
	if err := dec.Decode(&tail); err != io.EOF {
		return contract.Event{}, errors.New("trailing native screen JSON")
	}
	if strings.TrimSpace(in.Class) == "" || len(in.Class) > 256 {
		return contract.Event{}, errors.New("invalid native state class")
	}
	switch in.Event {
	case "screen-class":
		if in.ScaleX != 0 || in.ScaleY != 0 || len(in.Controls) > 0 {
			return contract.Event{}, errors.New("unexpected old screen-class metadata")
		}
	case "screen":
		if in.ScaleX < 1 || in.ScaleY < 1 || in.ScaleX > 8 || in.ScaleY > 8 || len(in.Controls) > 128 {
			return contract.Event{}, errors.New("invalid screen dimensions or control count")
		}
		for _, c := range in.Controls {
			if c.Role != "button" && c.Role != "text" {
				return contract.Event{}, errors.New("unsupported control role")
			}
			if len(c.Label) > 512 || strings.ContainsRune(c.Label, '\x00') || c.X < 0 || c.Y < 0 || c.W < 0 || c.H < 0 || c.X > 4096 || c.Y > 4096 || c.W > 4096 || c.H > 4096 {
				return contract.Event{}, errors.New("invalid screen control")
			}
		}
	default:
		return contract.Event{}, errors.New("unrecognized native screen envelope")
	}
	payload, _ := json.Marshal(struct {
		Class    string          `json:"class"`
		Source   string          `json:"source"`
		Limit    string          `json:"provenance_limit"`
		ScaleX   int             `json:"scale_x,omitempty"`
		ScaleY   int             `json:"scale_y,omitempty"`
		Controls []ScreenControl `json:"controls,omitempty"`
	}{in.Class, "actual-legacy-native-screen-envelope-adapter", "no tick/frame, asset identity, gameplay state, native acceptance or human approval", in.ScaleX, in.ScaleY, in.Controls})
	e := contract.Event{SchemaVersion: contract.Version, Type: "screen/view", EngineLineage: "oxce", EngineBuildSHA: p.EngineBuildSHA, ModManifestSHA: p.ModManifestSHA, SessionID: p.SessionID, ActorID: p.ActorID, AuthorityRole: "observer", Seq: seq, EvidenceGrade: "synthetic-contract", Payload: payload}
	if err := e.Validate(); err != nil {
		return e, fmt.Errorf("converted native screen event: %w", err)
	}
	return e, nil
}
