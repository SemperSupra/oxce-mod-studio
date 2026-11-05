package main

import (
	"bytes"
	"os"
	"path/filepath"
	"semper-supra/oxce-cockpit/adapters"
	"semper-supra/oxce-cockpit/contract"
	"strings"
	"testing"
)

func TestRealGHAEnvelopeImport(t *testing.T) {
	raw, e := os.ReadFile("../../fixtures/native-gha/trace.jsonl")
	if e != nil {
		t.Fatal(e)
	}
	p := adapters.Provenance{EngineBuildSHA: strings.Repeat("a", 64), ModManifestSHA: strings.Repeat("b", 64), SessionID: "gha40899658518", ActorID: "observer1"}
	var buf bytes.Buffer
	n, e := convert(bytes.NewReader(raw), &buf, p)
	if e != nil {
		t.Fatal(e)
	}
	if n != 1 {
		t.Fatalf("got %d", n)
	}
	es, e := contract.ReadJSONL(bytes.NewReader(buf.Bytes()))
	if e != nil {
		t.Fatal(e)
	}
	if len(es) != 1 || es[0].EvidenceGrade != "synthetic-contract" || !bytes.Contains(es[0].Payload, []byte("New Game")) {
		t.Fatal("native screen controls or grade lost")
	}
}
func TestImportNegative(t *testing.T) {
	p := adapters.Provenance{EngineBuildSHA: strings.Repeat("a", 64), ModManifestSHA: strings.Repeat("b", 64), SessionID: "gha40899658518", ActorID: "observer1"}
	for _, s := range []string{"", `{"event":"screen-class","class":""}`, `{"event":"screen","class":"X","scaleX":0}`} {
		if _, e := convert(strings.NewReader(s), &bytes.Buffer{}, p); e == nil {
			t.Errorf("accepted invalid trace %q", s)
		}
	}
	p.EngineBuildSHA = "not-a-real-hash"
	if _, e := convert(strings.NewReader(`{"event":"screen-class","class":"X"}`), &bytes.Buffer{}, p); e == nil {
		t.Fatal("accepted invalid provenance")
	}
}

// A late corrupt record used to leave a valid-looking partial output because
// main called os.Exit before its deferred removal. This is the regression guard.
func TestNoPartialOutputOnRejectedNativeTrace(t *testing.T) {
	p := adapters.Provenance{EngineBuildSHA: strings.Repeat("a", 64), ModManifestSHA: strings.Repeat("b", 64), SessionID: "gha40899658518", ActorID: "observer1"}
	input := filepath.Join(t.TempDir(), "trace.jsonl")
	output := filepath.Join(t.TempDir(), "adapted.jsonl")
	rows := `{"event":"screen-class","class":"MainMenuState"}` + "\n" + `{"event":"screen","class":"MainMenuState","scaleX":0}` + "\n"
	if err := os.WriteFile(input, []byte(rows), 0600); err != nil {
		t.Fatal(err)
	}
	if n, err := convertFile(input, output, p); err == nil || n != 1 {
		t.Fatalf("unexpected count=%d err=%v", n, err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("left partial output: %v", err)
	}
	valid := filepath.Join(t.TempDir(), "good.jsonl")
	if err := os.WriteFile(valid, []byte(`{"event":"screen-class","class":"MainMenuState"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if n, err := convertFile(valid, output, p); err != nil || n != 1 {
		t.Fatalf("valid count=%d err=%v", n, err)
	}
	if _, err := convertFile(valid, output, p); err == nil {
		t.Fatal("overwrote a prior completed output")
	}
}
