package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"semper-supra/oxce-cockpit/contract"
	"strings"
	"testing"
)

const h = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func fixture() contract.Event {
	return contract.Event{SchemaVersion: contract.Version, Type: "screen/view", EngineLineage: "oxce", EngineBuildSHA: h, ModManifestSHA: h, SessionID: "session1", ActorID: "testactor", AuthorityRole: "observer", Seq: 1, EvidenceGrade: "synthetic-contract", Payload: []byte(`{"class":"MainMenuState"}`)}
}
func TestSecurityAndFeedback(t *testing.T) {
	p := filepath.Join(t.TempDir(), "records.jsonl")
	s, err := NewServer([]contract.Event{fixture()}, strings.Repeat("z", 48), p)
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, token, origin, ctype, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		if token != "" {
			req.Header.Set("X-Cockpit-Token", token)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if ctype != "" {
			req.Header.Set("Content-Type", ctype)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, req)
		return w
	}
	token := strings.Repeat("z", 48)
	if r := call("GET", "/api/v0/events", "", "", "", ""); r.Code != 401 {
		t.Fatalf("missing token status %d", r.Code)
	}
	if r := call("GET", "/api/v0/events", token, "https://evil.test", "", ""); r.Code != 403 {
		t.Fatalf("foreign origin status %d", r.Code)
	}
	if r := call("GET", "/api/v0/capabilities", token, "", "", ""); r.Code != 200 || !strings.Contains(r.Body.String(), `"apply"`) {
		t.Fatal("capabilities incomplete")
	}
	if r := call("GET", "/api/v0/events", token, "", "", ""); r.Code != 200 || !strings.Contains(r.Body.String(), "synthetic") {
		t.Fatal("events unavailable")
	}
	mk := func(seq int, cat, note string) string {
		b, _ := json.Marshal(Feedback{SessionID: "session1", ActorID: "testactor", EventSeq: uint64(seq), Category: cat, Rating: 3, Note: note})
		return string(b)
	}
	for _, tt := range []struct {
		name  string
		body  string
		ctype string
		want  int
	}{
		{"unknown-asset-reference", mk(999, "visual", "text"), "application/json", 409},
		{"invalid-category", mk(1, "exec", "text"), "application/json", 400},
		{"empty-note", mk(1, "visual", ""), "application/json", 400},
		{"wrong-content-type", mk(1, "visual", "text"), "text/plain", 415},
		{"unknown-json-field", `{"session_id":"session1","actor_id":"testactor","event_seq":1,"category":"audio","rating":1,"note":"test","apply":true}`, "application/json", 400},
		{"oversize", strings.Repeat("X", 5000), "application/json", 400},
		{"valid", mk(1, "visual", "Sprite edge does not contrast"), "application/json", 201},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := call("POST", "/api/v0/feedback", token, "", tt.ctype, tt.body)
			if r.Code != tt.want {
				t.Fatalf("status=%d want=%d body=%s", r.Code, tt.want, r.Body.String())
			}
		})
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(b, []byte("\n")) != 1 {
		t.Fatal("expected one persisted feedback record")
	}
	if r := call("POST", "/api/v0/feedback", token, "https://attacker.invalid", "application/json", mk(1, "visual", "text")); r.Code != 403 {
		t.Fatal("cross origin accepted")
	}
}

func TestFixtureCannotAttestNativeOrHuman(t *testing.T) {
	raw, err := os.ReadFile("../../fixtures/single.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	events, err := contract.ReadJSONL(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, grade := range []string{"native-stock", "native-instrumented", "human-reviewed"} {
		modified := append([]contract.Event(nil), events...)
		modified[0].EvidenceGrade = grade
		if _, err := NewServer(modified, strings.Repeat("x", 32), filepath.Join(t.TempDir(), "feedback.jsonl")); err == nil {
			t.Errorf("trusted unverified %s", grade)
		}
	}
}
