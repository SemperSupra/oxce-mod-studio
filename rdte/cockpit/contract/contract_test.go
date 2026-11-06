package contract

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const sha = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func base() Event {
	return Event{SchemaVersion: Version, Type: "screen/view", EngineLineage: "oxce", EngineBuildSHA: sha, ModManifestSHA: sha, SessionID: "session-01", ActorID: "actor-01", AuthorityRole: "observer", Seq: 1, EvidenceGrade: "synthetic-contract", Payload: json.RawMessage(`{"class":"MainMenuState"}`)}
}
func TestValidAndNegativeControls(t *testing.T) {
	if err := base().Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		mut  func(*Event)
	}{
		{"schema", func(e *Event) { e.SchemaVersion = "99" }},
		{"unknown-kind", func(e *Event) { e.Type = "unknown/event" }},
		{"lineage", func(e *Event) { e.EngineLineage = "gamulator" }},
		{"build-sha", func(e *Event) { e.EngineBuildSHA = "not-a-hash" }},
		{"manifest-sha", func(e *Event) { e.ModManifestSHA = "b" }},
		{"id", func(e *Event) { e.SessionID = "../danger" }},
		{"actor", func(e *Event) { e.ActorID = "" }},
		{"role", func(e *Event) { e.AuthorityRole = "superuser" }},
		{"zero-seq", func(e *Event) { e.Seq = 0 }},
		{"asset-id", func(e *Event) { e.AssetID = "a path/asset" }},
		{"grade", func(e *Event) { e.EvidenceGrade = "Q5_APPROVED" }},
		{"payload-null", func(e *Event) { e.Payload = []byte("null") }},
		{"payload-array", func(e *Event) { e.Payload = []byte("[]") }},
		{"payload-invalid", func(e *Event) { e.Payload = []byte("{bad") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := base()
			tt.mut(&e)
			if e.Validate() == nil {
				t.Fatal("accepted invalid event")
			}
		})
	}
}
func TestStrictDecodeAndStream(t *testing.T) {
	e := base()
	buf, _ := json.Marshal(e)
	if _, err := Parse(buf); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{append(append([]byte{}, buf...), []byte(` {}`)...), []byte(`{"unrecognized":true}`), bytes.Repeat([]byte("x"), MaxRecordBytes+1), []byte(`{"event_type":`)} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("invalid record passed: %.25s", bad)
		}
	}
	v := NewValidator()
	if err := v.Accept(e); err != nil {
		t.Fatal(err)
	}
	if err := v.Accept(e); err == nil {
		t.Fatal("duplicate seq accepted")
	}
	e.Seq = 2
	if err := v.Accept(e); err != nil {
		t.Fatal(err)
	}
	e.ModManifestSHA = strings.Repeat("b", 64)
	e.Seq = 3
	if err := v.Accept(e); err == nil {
		t.Fatal("manifest mutation accepted")
	}
}
func TestJSONLFailureAndPositive(t *testing.T) {
	e := base()
	a, _ := json.Marshal(e)
	e.Seq = 2
	b, _ := json.Marshal(e)
	if got, err := ReadJSONL(bytes.NewReader(append(append(append([]byte{}, a...), '\n'), append(b, '\n')...))); err != nil || len(got) != 2 {
		t.Fatalf("%v %d", err, len(got))
	}
	for _, s := range []string{"", string(a) + "\n" + string(a) + "\n", string(a) + "\n{"} {
		if _, err := ReadJSONL(strings.NewReader(s)); err == nil {
			t.Fatalf("accepted bad stream: %.20s", s)
		}
	}
}

// A bounded fixture loader prevents a long-running trace from silently exhausting memory.
// Long native captures will use a separately qualified streaming sink.
func TestFixtureEventAndByteBudgets(t *testing.T) {
	e := base()
	var b bytes.Buffer
	for n := uint64(1); n <= MaxFixtureEvents; n++ {
		e.Seq = n
		raw, _ := json.Marshal(e)
		b.Write(raw)
		b.WriteByte('\n')
	}
	if got, err := ReadJSONL(bytes.NewReader(b.Bytes())); err != nil || len(got) != MaxFixtureEvents {
		t.Fatalf("boundary: %d %v", len(got), err)
	}
	e.Seq = MaxFixtureEvents + 1
	raw, _ := json.Marshal(e)
	b.Write(raw)
	b.WriteByte('\n')
	if _, err := ReadJSONL(bytes.NewReader(b.Bytes())); err == nil {
		t.Fatal("accepted too many events")
	}
	// Empty lines also cost bytes; do not permit a blank-line denial of service.
	huge := bytes.Repeat([]byte("\n"), MaxFixtureBytes+1)
	if _, err := ReadJSONL(bytes.NewReader(huge)); err == nil {
		t.Fatal("accepted oversized blank input")
	}
}
