// Package coopeval checks a contract-only two-peer action/digest transcript.
// It is NOT proof of real co-op networking, engine authority or gameplay.
package coopeval

import (
	"encoding/json"
	"errors"
	"fmt"
	"semper-supra/oxce-cockpit/contract"
)

type State struct {
	ActionID     string `json:"action_id"`
	AuthoritySeq uint64 `json:"authority_seq"`
	Digest       string `json:"digest"`
}

func collect(events []contract.Event, role string) (map[uint64]State, error) {
	out := map[uint64]State{}
	for _, e := range events {
		if e.EngineLineage != "oxce-coop" || e.AuthorityRole != role {
			return nil, errors.New("wrong engine lineage or role")
		}
		if e.Type != "game/state-digest" {
			continue
		}
		var s State
		if err := json.Unmarshal(e.Payload, &s); err != nil {
			return nil, err
		}
		if !contract.ValidID(s.ActionID) || s.AuthoritySeq == 0 || !contract.ValidSHA(s.Digest) {
			return nil, errors.New("bad state-digest payload")
		}
		if _, ok := out[s.AuthoritySeq]; ok {
			return nil, fmt.Errorf("duplicate authority seq %d", s.AuthoritySeq)
		}
		out[s.AuthoritySeq] = s
	}
	if len(out) == 0 {
		return nil, errors.New("no state digests")
	}
	return out, nil
}
func Correlate(host, client []contract.Event) error {
	if len(host) == 0 || len(client) == 0 {
		return errors.New("missing peer")
	}
	if host[0].SessionID != client[0].SessionID || host[0].ModManifestSHA != client[0].ModManifestSHA || host[0].EngineBuildSHA != client[0].EngineBuildSHA {
		return errors.New("session/mod manifest mismatch")
	}
	for _, peer := range [][]contract.Event{host, client} {
		v := contract.NewValidator()
		for _, event := range peer {
			if err := v.Accept(event); err != nil {
				return fmt.Errorf("invalid peer event: %w", err)
			}
		}
	}
	h, err := collect(host, "host")
	if err != nil {
		return fmt.Errorf("host: %w", err)
	}
	c, err := collect(client, "client")
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}
	if len(h) != len(c) {
		return fmt.Errorf("missing peer action %d vs %d", len(h), len(c))
	}
	for seq, s := range h {
		other, ok := c[seq]
		if !ok {
			return fmt.Errorf("client missing authority seq %d", seq)
		}
		if s != other {
			return fmt.Errorf("desynchronization at authority seq %d", seq)
		}
	}
	return nil
}
