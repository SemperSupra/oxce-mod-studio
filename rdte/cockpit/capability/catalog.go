// Package capability is a narrow catalog of SOURCE-PRESENT co-op TestServer commands.
// This package builds a proposed, validated JSON wire request. It NEVER executes it.
package capability

import (
	"bytes"
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

//go:embed catalog.json
var source embed.FS

type Param struct {
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Values   []string `json:"values,omitempty"`
}
type Descriptor struct {
	SourceLine    int              `json:"source_line"`
	Idempotent    bool             `json:"idempotent"`
	ID            string           `json:"id"`
	NativeCommand string           `json:"native_command"`
	Kind          string           `json:"kind"`
	Args          map[string]Param `json:"args"`
	Completion    string           `json:"completion"`
	Description   string           `json:"description"`
}
type Catalog struct {
	Version        string       `json:"catalog_version"`
	EngineLineage  string       `json:"engine_lineage"`
	UpstreamCommit string       `json:"upstream_commit"`
	NativeBlobSHA  string       `json:"native_blob_sha"`
	NativeSource   string       `json:"native_source"`
	EvidenceClass  string       `json:"evidence_class"`
	Capabilities   []Descriptor `json:"capabilities"`
}

var gitShaRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

// GitBlobSHA reproduces git hash-object for an unchanged source blob.
func GitBlobSHA(src []byte) string {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d\x00", len(src))
	h.Write(src)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySource is a source-presence gate, NOT a native behavior or safety proof.
// It refuses drift in the pinned upstream file and obsolete recorded line numbers.
func (c Catalog) VerifySource(src []byte) error {
	if GitBlobSHA(src) != c.NativeBlobSHA {
		return errors.New("pinned native TestServer source blob changed")
	}
	lines := bytes.Split(src, []byte("\n"))
	for _, op := range c.Capabilities {
		if op.SourceLine < 1 || op.SourceLine > len(lines) {
			return fmt.Errorf("%s: bad source line", op.ID)
		}
		target := []byte(`cmd == "` + op.NativeCommand + `"`)
		if !bytes.Contains(lines[op.SourceLine-1], target) {
			return fmt.Errorf("%s: handler not at pinned source line", op.ID)
		}
	}
	return nil
}

type Grant struct {
	Actor            string
	ControlLease     string
	ExpectedRevision *uint64
	SealedDiagnostic bool
}

func Load() (Catalog, error) {
	b, e := source.ReadFile("catalog.json")
	if e != nil {
		return Catalog{}, e
	}
	var c Catalog
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&c); e != nil {
		return c, e
	}
	if c.Version != "0.1.0" || c.EngineLineage != "oxce-coop" || !gitShaRE.MatchString(c.UpstreamCommit) || !gitShaRE.MatchString(c.NativeBlobSHA) {
		return c, errors.New("unadmitted catalog provenance")
	}
	ids := map[string]bool{}
	native := map[string]bool{}
	for _, op := range c.Capabilities {
		if op.ID == "" || op.NativeCommand == "" || op.SourceLine < 1 || ids[op.ID] || native[op.NativeCommand] {
			return c, errors.New("missing or duplicate capability")
		}
		ids[op.ID] = true
		native[op.NativeCommand] = true
		if op.Kind != "read" && op.Kind != "control" && op.Kind != "diagnostic" {
			return c, errors.New("bad operation classification")
		}
		if op.Idempotent != (op.Kind == "read") {
			return c, errors.New("unsafe retry metadata")
		}
		if op.Completion != "snapshot" && op.Completion != "ack_only" {
			return c, errors.New("bad completion classification")
		}
	}
	return c, nil
}
func (c Catalog) Lookup(id string) (Descriptor, bool) {
	for _, op := range c.Capabilities {
		if op.ID == id {
			return op, true
		}
	}
	return Descriptor{}, false
}

// Prepare validates a planned request; enforcing a grant in Go is NECESSARY but
// NOT SUFFICIENT for any live engine: it must also be checked at dispatch.
func (c Catalog) Prepare(id string, raw []byte, g Grant) ([]byte, error) {
	if !validActor(g.Actor) {
		return nil, errors.New("unrecognized actor")
	}
	if len(raw) > 2048 {
		return nil, errors.New("oversize arguments")
	}
	op, ok := c.Lookup(id)
	if !ok {
		return nil, errors.New("unknown capability")
	}
	switch op.Kind {
	case "read": // observation allowed to all three audiences
	case "control":
		if !validActor(g.Actor) || g.ControlLease == "" || g.ExpectedRevision == nil {
			return nil, errors.New("missing controller lease/revision")
		}
	case "diagnostic":
		if g.Actor != "automation" || !g.SealedDiagnostic || g.ControlLease == "" || g.ExpectedRevision == nil {
			return nil, errors.New("diagnostic capability requires sealed test grant")
		}
	}
	if len(raw) == 0 {
		raw = []byte(`{}`)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var args map[string]json.RawMessage
	if e := dec.Decode(&args); e != nil || args == nil {
		return nil, errors.New("args must be JSON object")
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		return nil, errors.New("trailing or malformed JSON")
	}
	// All parameter names and types are explicitly admitted. Native defaults are
	// never relied upon for safety-critical TCP launch.
	payload := map[string]any{"cmd": op.NativeCommand}
	for name := range args {
		if _, ok := op.Args[name]; !ok {
			return nil, fmt.Errorf("unknown argument %s", name)
		}
	}
	for name, p := range op.Args {
		rawValue, has := args[name]
		if !has {
			if p.Required {
				return nil, fmt.Errorf("missing argument %s", name)
			}
			continue
		}
		var s string
		if e := json.Unmarshal(rawValue, &s); e != nil || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("bad string %s", name)
		}
		switch p.Type {
		case "string":
			if len(s) > 128 {
				return nil, fmt.Errorf("argument too long %s", name)
			}
		case "enum":
			found := false
			for _, v := range p.Values {
				if s == v {
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("unsupported enum %s", name)
			}
		case "tcp_port":
			n, e := strconv.Atoi(s)
			if e != nil || n < 1 || n > 65535 || strconv.Itoa(n) != s {
				return nil, fmt.Errorf("bad port %s", name)
			}
		case "ipv4_loopback":
			ip := net.ParseIP(s)
			if ip == nil || ip.To4() == nil || !ip.IsLoopback() {
				return nil, fmt.Errorf("nonloopback test target %s", name)
			}
		default:
			return nil, fmt.Errorf("unknown parameter type %s", p.Type)
		}
		payload[name] = s
	}
	return json.Marshal(payload)
}
func validActor(s string) bool { return s == "human" || s == "automation" || s == "agent" }
