// traceimport converts an actual old OXCE native-screen JSONL capture into a
// bounded, explicitly UNATTESTED cockpit contract fixture. It cannot confer
// native gameplay, source approval, or human quality grade.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"semper-supra/oxce-cockpit/adapters"
	"semper-supra/oxce-cockpit/contract"
	"strings"
)

func fileSHA(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func convert(in io.Reader, out io.Writer, p adapters.Provenance) (int, error) {
	if !contract.ValidSHA(p.EngineBuildSHA) || !contract.ValidSHA(p.ModManifestSHA) {
		return 0, errors.New("verified binary/manifest hashes required")
	}
	s := bufio.NewScanner(in)
	s.Buffer(make([]byte, 4096), contract.MaxRecordBytes+1)
	v := contract.NewValidator()
	n := 0
	total := 0
	for s.Scan() {
		b := s.Bytes()
		total += len(b) + 1
		if total > contract.MaxFixtureBytes {
			return n, errors.New("native fixture byte budget exceeded")
		}
		if len(strings.TrimSpace(string(b))) == 0 {
			continue
		}
		if n >= contract.MaxFixtureEvents {
			return n, errors.New("native fixture event budget exceeded")
		}
		e, err := adapters.ConvertLegacyScreen(b, p, uint64(n+1))
		if err != nil {
			return n, fmt.Errorf("native trace row %d: %w", n+1, err)
		}
		if err = v.Accept(e); err != nil {
			return n, fmt.Errorf("contract row %d: %w", n+1, err)
		}
		line, err := json.Marshal(e)
		if err != nil {
			return n, err
		}
		if _, err = out.Write(append(line, '\n')); err != nil {
			return n, err
		}
		n++
	}
	if err := s.Err(); err != nil {
		return n, err
	}
	if n == 0 {
		return 0, errors.New("no native screen events")
	}
	return n, nil
}

// convertFile guarantees that a rejected or truncated native trace never
// leaves an incomplete output that could be mistaken for qualified evidence.
// All cleanup is performed BEFORE main decides to exit nonzero.
func convertFile(src, dest string, p adapters.Provenance) (int, error) {
	input, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer input.Close()
	outfile, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		outfile.Close()
		if !committed {
			os.Remove(dest)
		}
	}()
	n, err := convert(input, outfile, p)
	if err != nil {
		return n, err
	}
	if err := outfile.Sync(); err != nil {
		return n, err
	}
	if err := outfile.Close(); err != nil {
		return n, err
	}
	committed = true
	return n, nil
}

func run() error {
	src := flag.String("trace", "", "raw GHA screen trace JSONL input")
	bin := flag.String("engine", "", "exact native engine binary to hash")
	mod := flag.String("mod-manifest", "", "exact native mod metadata file to hash")
	expectedEngine := flag.String("expect-engine-sha", "", "optional independently recorded binary SHA-256; mismatch fails closed")
	expectedMod := flag.String("expect-mod-sha", "", "optional independently recorded mod metadata SHA-256; mismatch fails closed")
	out := flag.String("out", "", "output contract JSONL (must not exist)")
	session := flag.String("session", "gha40899658518", "session id")
	actor := flag.String("actor", "observer1", "actor id")
	flag.Parse()
	if *src == "" || *bin == "" || *mod == "" || *out == "" {
		return errors.New("required: -trace -engine -mod-manifest -out")
	}
	ehash, err := fileSHA(*bin)
	if err != nil {
		return err
	}
	mhash, err := fileSHA(*mod)
	if err != nil {
		return err
	}
	if (*expectedEngine != "" && (!contract.ValidSHA(*expectedEngine) || *expectedEngine != ehash)) || (*expectedMod != "" && (!contract.ValidSHA(*expectedMod) || *expectedMod != mhash)) {
		return errors.New("input digest differs from independently recorded evidence")
	}
	n, err := convertFile(*src, *out, adapters.Provenance{EngineBuildSHA: ehash, ModManifestSHA: mhash, SessionID: *session, ActorID: *actor})
	if err != nil {
		return err
	}
	fmt.Printf("CONTRACT_ADAPTED_FROM_REAL_NATIVE_TRACE_NOT_NATIVE_GAMEPLAY_ACCEPTANCE events=%d engine_sha256=%s mod_sha256=%s\n", n, ehash, mhash)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
