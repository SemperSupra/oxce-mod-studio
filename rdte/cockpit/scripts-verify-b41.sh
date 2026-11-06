#!/bin/sh
set -eu
cd "$(dirname "$0")"
mkdir -p .out
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
go build -trimpath -o .out/contractcheck ./cmd/contractcheck
go build -trimpath -o .out/catalogcheck ./cmd/catalogcheck
go build -trimpath -o .out/cockpit-server ./cmd/cockpit
go build -trimpath -o .out/traceimport ./cmd/traceimport
./.out/contractcheck -in fixtures/single.jsonl
./.out/contractcheck -host fixtures/host.jsonl -client fixtures/client.jsonl
if ./.out/contractcheck -host fixtures/host.jsonl -client fixtures/client-desync-negative.jsonl; then echo 'FAIL: desync accepted' >&2; exit 1; fi
./.out/catalogcheck -read-id game.inspect_state
if ./.out/catalogcheck -read-id session.join; then echo 'FAIL: control admitted as read' >&2; exit 1; fi
PYTHONPATH=tools python3 -m unittest -v tools/test_python_native_read.py
python3 tools/netlab_probe.py
python3 - <<'PY'
# Syntax-only JS and HTML checks run elsewhere; browser/AX native QA remains NOT_RUN.
from html.parser import HTMLParser
from pathlib import Path
class H(HTMLParser):
 def __init__(self):super().__init__();self.ids=set();self.buttons=0
 def handle_starttag(self,tag,attrs):
  d=dict(attrs)
  if 'id' in d:self.ids.add(d['id'])
  if tag=='button':self.buttons+=1
p=H();p.feed(Path('cmd/cockpit/index.html').read_text())
assert {'events','details','feedback','status','typefilter'}.issubset(p.ids)
assert p.buttons>=4
print('HTML_STATIC_STRUCTURE_PASS_NOT_BROWSER_ACCEPTANCE')
PY
printf 'B41 SOURCE ONLY PASS; actual browser/engine/co-op native not run by source verifier\n'
