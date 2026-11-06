// Cockpit is an isolated, fixture-backed, localhost-only read-only prototype.
// Nothing in this process can apply game edits, operate the engine, or publish.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"semper-supra/oxce-cockpit/contract"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed index.html
var ui embed.FS

const maxFeedback = 4096

var feedbackCategories = map[string]bool{"visual": true, "audio": true, "motion": true, "mechanics": true, "network": true, "accessibility": true}

type Feedback struct {
	SessionID string `json:"session_id"`
	ActorID   string `json:"actor_id"`
	EventSeq  uint64 `json:"event_seq"`
	Category  string `json:"category"`
	Rating    int    `json:"rating"`
	Note      string `json:"note"`
}
type CapturedFeedback struct {
	Feedback
	EngineBuildSHA string `json:"engine_build_sha"`
	ModManifestSHA string `json:"mod_manifest_sha"`
	EngineLineage  string `json:"engine_lineage"`
	AssetID        string `json:"asset_id,omitempty"`
	RecordedAt     string `json:"recorded_at"`
	EvidenceClass  string `json:"evidence_class"`
}
type Server struct {
	token        string
	events       []contract.Event
	evidencePath string
	mu           sync.Mutex
}

func NewServer(events []contract.Event, token, path string) (*Server, error) {
	if len(events) == 0 || len(token) < 24 {
		return nil, errors.New("events and random secret required")
	}
	if err := contract.NewValidator().Accept(events[0]); err != nil {
		return nil, err
	}
	v := contract.NewValidator()
	for _, e := range events {
		// The public fixture UI cannot validate engine execution or human approval.
		// A declared native/human grade in imported JSON is not trustworthy.
		if e.EvidenceGrade != "synthetic-contract" {
			return nil, errors.New("unattested evidence grade in fixture")
		}
		if err := v.Accept(e); err != nil {
			return nil, err
		}
	}
	return &Server{events: events, token: token, evidencePath: path}, nil
}
func (s *Server) auth(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("X-Cockpit-Token") != s.token {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
		http.Error(w, "origin denied", http.StatusForbidden)
		return false
	}
	return true
}
func jsonResponse(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		b, _ := ui.ReadFile("index.html")
		w.Write(b)
	})
	mux.HandleFunc("/api/v0/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method denied", 405)
			return
		}
		if !s.auth(w, r) {
			return
		}
		jsonResponse(w, 200, map[string]any{"schema_version": contract.Version, "mode": "FIXTURE_REVIEW_ONLY", "capabilities": []string{"inspect", "annotate"}, "denied": []string{"input", "preview", "apply", "publish", "external_network"}})
	})
	mux.HandleFunc("/api/v0/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method denied", 405)
			return
		}
		if !s.auth(w, r) {
			return
		}
		jsonResponse(w, 200, map[string]any{"evidence_class": "SYNTHETIC_CONTRACT_ONLY_NOT_NATIVE_GAMEPLAY", "source_attested": false, "human_approved": false, "events": s.events})
	})
	mux.HandleFunc("/api/v0/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method denied", 405)
			return
		}
		if !s.auth(w, r) {
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported content type", 415)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxFeedback)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var f Feedback
		if err := dec.Decode(&f); err != nil {
			http.Error(w, "invalid feedback", 400)
			return
		}
		var tail any
		if err := dec.Decode(&tail); err != io.EOF {
			http.Error(w, "trailing payload", 400)
			return
		}
		if !feedbackCategories[f.Category] || f.Rating < 1 || f.Rating > 5 || len(f.Note) > 2000 || strings.TrimSpace(f.Note) == "" {
			http.Error(w, "invalid feedback fields", 400)
			return
		}
		var found *contract.Event
		for i := range s.events {
			e := &s.events[i]
			if e.SessionID == f.SessionID && e.ActorID == f.ActorID && e.Seq == f.EventSeq {
				found = e
				break
			}
		}
		if found == nil {
			http.Error(w, "event reference not found", 409)
			return
		}
		record := CapturedFeedback{Feedback: f, EngineBuildSHA: found.EngineBuildSHA, ModManifestSHA: found.ModManifestSHA, EngineLineage: found.EngineLineage, AssetID: found.AssetID, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), EvidenceClass: "USER_FEEDBACK_ON_SYNTHETIC_FIXTURE"}
		b, _ := json.Marshal(record)
		h := sha256.Sum256(b)
		s.mu.Lock()
		defer s.mu.Unlock()
		out, err := os.OpenFile(s.evidencePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			http.Error(w, "evidence write failed", 500)
			return
		}
		n, writeErr := out.Write(append(b, '\n'))
		closeErr := out.Close()
		if writeErr != nil || closeErr != nil || n != len(b)+1 {
			http.Error(w, "evidence write failed", 500)
			return
		}
		jsonResponse(w, 201, map[string]any{"status": "stored_local_only", "record_sha256": hex.EncodeToString(h[:]), "evidence_class": record.EvidenceClass})
	})
	return mux
}
func main() {
	trace := flag.String("events", "", "versioned JSONL event fixture")
	out := flag.String("feedback", ".out/feedback.jsonl", "private local output file")
	port := flag.Int("port", 0, "localhost port (0 = random)")
	flag.Parse()
	if *trace == "" {
		log.Fatal("-events is required")
	}
	if *port < 0 || *port > 65535 {
		log.Fatal("bad port")
	}
	b, err := os.Open(*trace)
	if err != nil {
		log.Fatal(err)
	}
	events, err := contract.ReadJSONL(b)
	b.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0700); err != nil {
		log.Fatal(err)
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		log.Fatal(err)
	}
	token := hex.EncodeToString(raw[:])
	s, err := NewServer(events, token, *out)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Cockpit FIXTURE_REVIEW_ONLY on http://%s/#token=%s\n", l.Addr().String(), token)
	log.Fatal(http.Serve(l, s.Handler()))
}
