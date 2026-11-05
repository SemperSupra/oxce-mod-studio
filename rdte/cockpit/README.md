# OXCE Cockpit — bounded development prototype

Original, first-party Go + Python source for a read-only OXCE development/feedback prototype.
This branch contains **no** commercial game data, private engine source, credentials,
network endpoints, production deployment, or native gameplay acceptance claims.

## Verification

```sh
cd rdte/cockpit
go test ./...
go test -race ./...
go vet ./...
sh scripts-verify-b42.sh
```

The `fixtures/native-gha/trace.jsonl` file is an original-public historical
native screen-event fixture; adapting it does **not** qualify native gameplay.
The capability catalog contains ten source-verified legacy TestServer operation
names but does not execute them or authenticate their caller. Actual native
host/client tests, wildcard listener isolation, and human review are separate.

`tools/native_gate.py` inspects local ELF binary/runtime dependencies without
starting the game. It blocks when libraries are absent; use original supported
packages on an admitted compatible CI worker rather than creating fake ABI shims.

## Security and provenance

Only localhost fixture review is implemented. Never publicly expose the raw
co-op TestServer. The HTML interface is not a live game window. An event's
`evidence_grade` field does not establish a trusted runtime or human attestation.
All real native build binaries and original CI logs remain in owner-controlled
artifact custody, outside this source branch.

## Scope

This is a source-only integration workcell, not a release. No new scripting VM,
no replacement of upstream Python `GameClient`, no protected mission-binding
changes, and no private Actions jobs are required by this source.
