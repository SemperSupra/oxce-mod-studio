// Source-only catalog CLI. No network I/O and no credentials.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"semper-supra/oxce-cockpit/capability"
)

func main() {
	sourceFile := flag.String("verify-source", "", "verify exact pinned TestServer.cpp blob and all handler lines")
	readID := flag.String("read-id", "", "emit a validated read-only native TestServer request")
	args := flag.String("args", "{}", "JSON arguments for selected read-only request")
	flag.Parse()
	c, err := capability.Load()
	if err != nil {
		fail(err)
	}
	if *sourceFile != "" {
		b, err := os.ReadFile(*sourceFile)
		if err != nil {
			fail(err)
		}
		if err = c.VerifySource(b); err != nil {
			fail(err)
		}
		fmt.Fprintln(os.Stdout, "PINNED_HANDLER_SOURCE_PRESENT_NOT_NATIVE_ACCEPTANCE")
		return
	}
	if *readID == "" {
		if err := json.NewEncoder(os.Stdout).Encode(c); err != nil {
			fail(err)
		}
		return
	}
	op, ok := c.Lookup(*readID)
	if !ok || op.Kind != "read" {
		fail(fmt.Errorf("refused: only cataloged read-only operations permitted"))
	}
	prepared, err := c.Prepare(*readID, []byte(*args), capability.Grant{Actor: "automation"})
	if err != nil {
		fail(err)
	}
	fmt.Println(string(prepared))
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
