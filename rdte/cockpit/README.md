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

`tools/native_gate.py` performs a **nonexecuting source-level direct ELF
DT_NEEDED inspection** against a SHA256-bound open file descriptor, using trusted
`readelf` and the host's `ldconfig -p` cache. It never invokes `ldd` or the
target ELF/interpreter, and checks for path/inode/content drift. A missing SONAME
in that cache blocks direct-dependency admission. A positive result now reports
`oxce.native-readiness.v2 / DIRECT_DEPS_IN_HOST_CACHE`, **not** v1
`RUNTIME_DEPS_READY`: transitive dependencies, RPATH, actual dynamic loading,
headless SDL runtime, network containment and gameplay are all independently
unverified. Use original supported libraries on a separately admitted compatible
worker; never fake ABI compatibility with symlinks.

B45 source regression: 13/13 offline unit tests PASS, including hash/symlink,
rename-after-hash, in-place mutation, timeout, parser error, no-import and host
cache-miss negatives. Nonexecuting direct-import check of local pinned `/bin/ls`
returned `DIRECT_DEPS_IN_HOST_CACHE` with `native_launched=false`. This is NOT
a new native OXCE game run or public GitHub Actions dispatch. Exact source/test
blobs are in this existing DRAFT PR; original first failure and previous
historical native evidence are preserved.

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
