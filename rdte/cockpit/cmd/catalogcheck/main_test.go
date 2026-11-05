package main

import (
	"encoding/json"
	"os/exec"
	"testing"
)

func TestCLIReadOnlyAndControlRefusal(t *testing.T) {
	tests := []struct {
		id   string
		pass bool
		cmd  string
	}{
		{"game.inspect_state", true, "get_state"},
		{"coop.inspect_checksum", true, "shared_checksum"},
		{"session.join", false, ""},
		{"coop.force_resync", false, ""},
		{"game.fake", false, ""},
	}
	for _, tc := range tests {
		out, err := exec.Command("go", "run", ".", "-read-id", tc.id).CombinedOutput()
		if (err == nil) != tc.pass {
			t.Errorf("%s expectedPass=%v, output=%q, err=%v", tc.id, tc.pass, out, err)
			continue
		}
		if tc.pass {
			var v struct {
				Cmd string `json:"cmd"`
			}
			if json.Unmarshal(out, &v) != nil || v.Cmd != tc.cmd {
				t.Errorf("%s bad wire %q", tc.id, out)
			}
		}
	}
}
