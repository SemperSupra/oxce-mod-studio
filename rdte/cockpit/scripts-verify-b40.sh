#!/bin/sh
set -eu
cd "$(dirname "$0")"
mkdir -p .out
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
go build -trimpath -o .out/contractcheck ./cmd/contractcheck
go build -trimpath -o .out/cockpit-server ./cmd/cockpit
go build -trimpath -o .out/catalogcheck ./cmd/catalogcheck
./.out/contractcheck -in fixtures/single.jsonl
./.out/contractcheck -host fixtures/host.jsonl -client fixtures/client.jsonl
if ./.out/contractcheck -host fixtures/host.jsonl -client fixtures/client-desync-negative.jsonl; then
 echo 'FAIL: accepted deliberate co-op desynchronization' >&2; exit 1
fi
./.out/catalogcheck -read-id game.inspect_state
if ./.out/catalogcheck -read-id session.join; then
 echo 'FAIL: read-only catalog permitted write-like control' >&2; exit 1
fi
PYTHONPATH=tools python3 -m unittest -v tools/test_python_native_read.py
python3 tools/netlab_probe.py
printf 'B40 SOURCE+INTEROP+FAKE-SOCKET-ISOLATION PASS; CLOUD/NATIVE OXCE NOT_RUN\n'
