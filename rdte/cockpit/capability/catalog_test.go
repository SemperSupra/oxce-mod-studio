package capability

import (
	"strings"
	"testing"
)

func TestCatalogSourceAndSemantics(t *testing.T) {
	c, e := Load()
	if e != nil {
		t.Fatal(e)
	}
	if len(c.Capabilities) != 10 {
		t.Fatalf("unexpected prototype count %d", len(c.Capabilities))
	}
	if c.EvidenceClass != "SOURCE_VERIFIED_COMMAND_PRESENCE_NOT_NATIVE_ACCEPTANCE" {
		t.Fatal("misleading grade")
	}
	for _, name := range []string{"game.inspect_state", "session.join", "coop.force_resync"} {
		if _, ok := c.Lookup(name); !ok {
			t.Fatal(name)
		}
	}
}
func TestPrepareAllowedAndDenied(t *testing.T) {
	c, e := Load()
	if e != nil {
		t.Fatal(e)
	}
	r := uint64(5)
	human := Grant{Actor: "human", ControlLease: "lease-01", ExpectedRevision: &r}
	for _, tc := range []struct {
		id, args string
		grant    Grant
		allow    bool
	}{
		{"game.inspect_state", "{}", Grant{Actor: "agent"}, true},
		{"game.inspect_state", "{}", Grant{}, false},
		{"session.open_new_game", `{"mode":"coop_shared"}`, human, true},
		{"session.open_new_game", `{"mode":"bogus"}`, human, false},
		{"session.open_new_game", `{"mode":"coop_shared"}`, Grant{Actor: "agent"}, false},
		{"session.open_new_game", `{"mode":"coop_shared"}`, Grant{Actor: "human", ControlLease: "lease-01"}, false},
		{"session.join", `{"ip":"127.0.0.1","port":"47900","player":"client"}`, human, true},
		{"session.join", `{"ip":"0.0.0.0","port":"47900","player":"client"}`, human, false},
		{"session.join", `{"ip":"8.8.8.8","port":"47900","player":"client"}`, human, false},
		{"session.join", `{"ip":"127.0.0.1","port":"0","player":"client"}`, human, false},
		{"session.join", `{"ip":"127.0.0.1","port":"47900","player":"client","rm":"/"}`, human, false},
		{"session.host", `{"server":"lab","port":"47900","player":"host"}`, human, true},
		{"coop.force_resync", "{}", human, false},
		{"coop.force_resync", "{}", Grant{Actor: "automation", ControlLease: "lease", ExpectedRevision: &r}, false},
		{"coop.force_resync", "{}", Grant{Actor: "automation", ControlLease: "lease", ExpectedRevision: &r, SealedDiagnostic: true}, true},
		{"game.exec_arbitrary", "{}", human, false},
	} {
		out, err := c.Prepare(tc.id, []byte(tc.args), tc.grant)
		if (err == nil) != tc.allow {
			t.Errorf("%s allow=%v err=%v", tc.id, tc.allow, err)
		}
		if err == nil && !strings.Contains(string(out), `"cmd"`) {
			t.Errorf("missing native cmd: %s", out)
		}
	}
}

func TestSourceIntegrityAndLineAnchors(t *testing.T) {
	src := []byte("// header\n if (cmd == \"get_state\") {return;}\n")
	c := Catalog{NativeBlobSHA: GitBlobSHA(src), Capabilities: []Descriptor{{ID: "game.inspect_state", NativeCommand: "get_state", SourceLine: 2}}}
	if err := c.VerifySource(src); err != nil {
		t.Fatal(err)
	}
	if err := c.VerifySource(append([]byte{}, src[:len(src)-1]...)); err == nil {
		t.Fatal("accepted modified source")
	}
	c.Capabilities[0].SourceLine = 1
	if err := c.VerifySource(src); err == nil {
		t.Fatal("accepted wrong pinned line")
	}
}
