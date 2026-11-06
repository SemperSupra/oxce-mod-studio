// Package contract defines the zero-dependency, read-only OXCE cockpit event v0.
// Validation proves only syntax/contract, NOT that a source really is stock OXCE.
package contract

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const Version = "0.1.0"
const MaxRecordBytes = 64 * 1024

// Fixture limits, not a substitute for bounded streaming native capture.
const MaxFixtureEvents = 4096
const MaxFixtureBytes = 8 * 1024 * 1024

var idRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.:-]{0,95}$`)
var kinds = map[string]bool{
	"session/start": true, "screen/view": true, "render/frame": true, "palette/active": true,
	"asset/resolve": true, "audio/trigger": true, "game/action": true, "game/state-digest": true,
	"coop/authority": true, "coop/peer": true, "trace/drop": true, "feedback/pin": true,
}
var roles = map[string]bool{"local": true, "host": true, "client": true, "observer": true}
var lineages = map[string]bool{"oxce": true, "oxce-coop": true}
var grades = map[string]bool{"synthetic-contract": true, "native-instrumented": true, "native-stock": true, "human-reviewed": true}

// A tick/frame is optional because legacy screen-class traces do not expose one.
// Engine build and mod manifest hashes must be present, not inferred by an adapter.
type Event struct {
	SchemaVersion  string          `json:"schema_version"`
	Type           string          `json:"event_type"`
	EngineLineage  string          `json:"engine_lineage"`
	EngineBuildSHA string          `json:"engine_build_sha"`
	ModManifestSHA string          `json:"mod_manifest_sha"`
	SessionID      string          `json:"session_id"`
	ActorID        string          `json:"actor_id"`
	AuthorityRole  string          `json:"authority_role"`
	Seq            uint64          `json:"seq"`
	Tick           *uint64         `json:"tick,omitempty"`
	Frame          *uint64         `json:"frame,omitempty"`
	AssetID        string          `json:"asset_id,omitempty"`
	EvidenceGrade  string          `json:"evidence_grade"`
	Payload        json.RawMessage `json:"payload"`
}

func ValidSHA(v string) bool {
	if len(v) != 64 {
		return false
	}
	_, err := hex.DecodeString(v)
	return err == nil && v == strings.ToLower(v)
}
func ValidID(v string) bool { return idRE.MatchString(v) }
func (e Event) Validate() error {
	switch {
	case e.SchemaVersion != Version:
		return fmt.Errorf("unsupported schema_version %q", e.SchemaVersion)
	case !kinds[e.Type]:
		return fmt.Errorf("unknown event_type %q", e.Type)
	case !lineages[e.EngineLineage]:
		return fmt.Errorf("invalid engine_lineage %q", e.EngineLineage)
	case !ValidSHA(e.EngineBuildSHA):
		return errors.New("invalid engine_build_sha")
	case !ValidSHA(e.ModManifestSHA):
		return errors.New("invalid mod_manifest_sha")
	case !ValidID(e.SessionID):
		return errors.New("invalid session_id")
	case !ValidID(e.ActorID):
		return errors.New("invalid actor_id")
	case !roles[e.AuthorityRole]:
		return errors.New("invalid authority_role")
	case e.Seq == 0:
		return errors.New("seq must start at 1")
	case e.AssetID != "" && !ValidID(e.AssetID):
		return errors.New("invalid asset_id")
	case !grades[e.EvidenceGrade]:
		return errors.New("invalid evidence_grade")
	case len(e.Payload) == 0:
		return errors.New("missing payload")
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(e.Payload, &data); err != nil || data == nil {
		return errors.New("payload must be a JSON object")
	}
	return nil
}
func Parse(line []byte) (Event, error) {
	var e Event
	if len(line) > MaxRecordBytes {
		return e, fmt.Errorf("record exceeds %d bytes", MaxRecordBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return e, fmt.Errorf("decode: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return e, errors.New("trailing JSON after event")
	}
	return e, e.Validate()
}

// Validator enforces monotonic per-actor ordering; event arrival order across peers
// is not used as a substitute for host authority or network synchronization.
type Validator struct {
	seen     map[string]uint64
	session  string
	lineage  string
	build    string
	manifest string
	count    int
}

func NewValidator() *Validator { return &Validator{seen: make(map[string]uint64)} }
func (v *Validator) Accept(e Event) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if v.count == 0 {
		v.session = e.SessionID
		v.lineage = e.EngineLineage
		v.build = e.EngineBuildSHA
		v.manifest = e.ModManifestSHA
	} else if e.SessionID != v.session || e.EngineLineage != v.lineage || e.EngineBuildSHA != v.build || e.ModManifestSHA != v.manifest {
		return errors.New("event provenance changed within stream")
	}
	if last, ok := v.seen[e.ActorID]; ok && e.Seq <= last {
		return fmt.Errorf("duplicate/out-of-order seq for actor %s: %d <= %d", e.ActorID, e.Seq, last)
	}
	v.seen[e.ActorID] = e.Seq
	v.count++
	return nil
}
func (v *Validator) Count() int { return v.count }
func ReadJSONL(r io.Reader) ([]Event, error) {
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 4096), MaxRecordBytes+1)
	v := NewValidator()
	var out []Event
	var totalBytes int64
	var lineNumber int
	for scan.Scan() {
		lineNumber++
		line := scan.Bytes()
		totalBytes += int64(len(line)) + 1
		if totalBytes > MaxFixtureBytes {
			return nil, fmt.Errorf("fixture exceeds %d bytes", MaxFixtureBytes)
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if len(out) >= MaxFixtureEvents {
			return nil, fmt.Errorf("fixture exceeds %d events", MaxFixtureEvents)
		}
		e, err := Parse(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		if err = v.Accept(e); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		out = append(out, e)
	}
	if err := scan.Err(); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if len(out) == 0 {
		return nil, errors.New("empty trace")
	}
	return out, nil
}
