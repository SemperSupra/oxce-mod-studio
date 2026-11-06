package adapters

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestActualGHARunScreenEnvelope(t *testing.T) {
	line, e := os.ReadFile("../fixtures/native-gha/trace.jsonl")
	if e != nil {
		t.Fatal(e)
	}
	p := Provenance{strings.Repeat("a", 64), strings.Repeat("b", 64), "gha40899658518", "observer1"}
	event, e := ConvertLegacyScreen(line, p, 1)
	if e != nil {
		t.Fatal(e)
	}
	if event.EvidenceGrade != "synthetic-contract" || event.Frame != nil || event.Tick != nil || event.AssetID != "" {
		t.Fatal("unearned native/gameplay acceptance")
	}
	var details struct {
		Class    string          `json:"class"`
		ScaleX   int             `json:"scale_x"`
		Controls []ScreenControl `json:"controls"`
	}
	if e = json.Unmarshal(event.Payload, &details); e != nil {
		t.Fatal(e)
	}
	if details.Class != "N8OpenXcom13MainMenuStateE" || details.ScaleX != 2 || len(details.Controls) != 7 || details.Controls[0].Label != "New Game" {
		t.Fatalf("native source data not preserved: %+v", details)
	}
}
func TestNativeScreenNegatives(t *testing.T) {
	p := Provenance{strings.Repeat("a", 64), strings.Repeat("b", 64), "session1", "observer1"}
	bad := []string{
		`{"event":"screen","class":"MainMenu","scaleX":0,"scaleY":2}`,
		`{"event":"screen","class":"MainMenu","scaleX":2,"scaleY":2,"controls":[{"role":"script","label":"unsafe","x":0,"y":0,"w":2,"h":3}]}`,
		`{"event":"screen","class":"MainMenu","scaleX":2,"scaleY":2,"extra":1}`,
		`{"event":"screen","class":"MainMenu","scaleX":2,"scaleY":2}{"event":"other"}`,
		`{"event":"screen-class","class":"MainMenu","scaleX":2}`,
	}
	for _, s := range bad {
		if _, e := ConvertLegacyScreen([]byte(s), p, 1); e == nil {
			t.Fatalf("accepted bad native record: %s", s)
		}
	}
}
