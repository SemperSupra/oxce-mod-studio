package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"semper-supra/oxce-cockpit/contract"
	"semper-supra/oxce-cockpit/coopeval"
)

func read(name string) ([]contract.Event, error) {
	f, e := os.Open(name)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return contract.ReadJSONL(f)
}
func main() {
	host := flag.String("host", "", "host JSONL trace")
	client := flag.String("client", "", "client JSONL trace")
	in := flag.String("in", "", "one-engine JSONL trace")
	flag.Parse()
	var err error
	count := 0
	mode := "single"
	if *host != "" && *client != "" {
		var h, c []contract.Event
		h, err = read(*host)
		if err == nil {
			c, err = read(*client)
		}
		if err == nil {
			err = coopeval.Correlate(h, c)
		}
		count = len(h) + len(c)
		mode = "coop-contract"
	} else if *in != "" {
		var a []contract.Event
		a, err = read(*in)
		count = len(a)
	} else {
		fmt.Fprintln(os.Stderr, "use -in <trace> or -host <trace> -client <trace>")
		os.Exit(2)
	}
	status := "PASS"
	reason := "syntactic contract validation only; not native-gameplay evidence"
	if err != nil {
		status = "FAIL"
		reason = err.Error()
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"status": status, "events": count, "mode": mode, "evidence_class": "CONTRACT_ONLY_NOT_NATIVE_ACCEPTANCE", "reason": reason})
	if err != nil {
		os.Exit(1)
	}
}
