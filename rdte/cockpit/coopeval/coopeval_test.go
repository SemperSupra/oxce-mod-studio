package coopeval

import (
	"encoding/json"
	"semper-supra/oxce-cockpit/contract"
	"strings"
	"testing"
)

const sha = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func pair(role string, seq uint64, actionID string, digest string) contract.Event {
	b, _ := json.Marshal(State{ActionID: actionID, AuthoritySeq: seq, Digest: digest})
	return contract.Event{SchemaVersion: contract.Version, Type: "game/state-digest", EngineLineage: "oxce-coop", EngineBuildSHA: sha, ModManifestSHA: sha, SessionID: "session-a", ActorID: role + "-actor", AuthorityRole: role, Seq: seq, EvidenceGrade: "synthetic-contract", Payload: b}
}
func TestPositiveAndFailureModes(t *testing.T) {
	h := []contract.Event{pair("host", 1, "action1", sha), pair("host", 2, "action2", sha)}
	c := []contract.Event{pair("client", 1, "action1", sha), pair("client", 2, "action2", sha)}
	if err := Correlate(h, c); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		run  func() error
	}{
		{"missing-peer", func() error { return Correlate(h, nil) }},
		{"missing-action", func() error { return Correlate(h, c[:1]) }},
		{"wrong-manifest", func() error {
			x := append([]contract.Event{}, c...)
			x[0].ModManifestSHA = strings.Repeat("b", 64)
			return Correlate(h, x)
		}},
		{"desync", func() error {
			x := append([]contract.Event{}, c...)
			x[1] = pair("client", 2, "action2", strings.Repeat("b", 64))
			return Correlate(h, x)
		}},
		{"wrong-role", func() error {
			x := append([]contract.Event{}, c...)
			x[0].AuthorityRole = "host"
			return Correlate(h, x)
		}},
		{"duplicate-authority", func() error {
			x := append([]contract.Event{}, c...)
			x[1] = pair("client", 2, "action2", sha)
			x[1].Payload = x[0].Payload
			return Correlate(h, x)
		}},
		{"bad-payload", func() error {
			x := append([]contract.Event{}, c...)
			x[0].Payload = []byte(`{}`)
			return Correlate(h, x)
		}},
		{"wrong-lineage", func() error {
			x := append([]contract.Event{}, c...)
			x[0].EngineLineage = "oxce"
			return Correlate(h, x)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err == nil {
				t.Fatal("accepted invalid pair")
			}
		})
	}
}
