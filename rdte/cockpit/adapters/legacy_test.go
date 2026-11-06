package adapters

import (
	"strings"
	"testing"
)

func TestConversionBoundaries(t *testing.T) {
	p := Provenance{strings.Repeat("a", 64), strings.Repeat("b", 64), "session1", "observer1"}
	e, err := ConvertLegacyScreen([]byte(`{"event":"screen-class","class":"MainMenuState"}`), p, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Tick != nil || e.Frame != nil || e.AssetID != "" || e.EvidenceGrade != "synthetic-contract" {
		t.Fatal("fabricated native facts")
	}
	for _, s := range []string{`{"event":"screen-class","class":""}`, `{"event":"wrong","class":"State"}`, `{"event":"screen-class","class":"S","injected":1}`, `broken`} {
		if _, err := ConvertLegacyScreen([]byte(s), p, 1); err == nil {
			t.Fatalf("accepted invalid legacy %q", s)
		}
	}
}
