#!/usr/bin/env bash
# Only native crash diagnosis: same immutable binary and public original fixture.
# Records SIGSEGV backtrace rather than guessing a correction from YAML.
set -u
package="$1"; dest="$2"
mkdir -p "$dest"
cp -a "$package/user" "$dest/user"
cp -a "$package/config" "$dest/config"
: > "$dest/commands.txt"
chmod 600 "$dest/commands.txt"
OXCE_RDTE_MAILBOX="$PWD/$dest/commands.txt" \
OXCE_RDTE_TRACE="$PWD/$dest/trace.jsonl" \
SDL_AUDIODRIVER=dummy timeout --kill-after=2s 29s \
  gdb -q -batch -ex 'set pagination off' -ex 'set confirm off' \
  -ex 'handle SIGPIPE nostop noprint pass' -ex run \
  -ex 'thread apply all bt 18' --args "$package/openxcom-sf-rdte" \
  -data "$PWD/$package/data" -user "$PWD/$dest/user" -config "$PWD/$dest/config" \
  > "$dest/gdb.log" 2>&1 &
debugger=$!
sleep 6
printf 'CLICK 420 200\n' >> "$dest/commands.txt"
sleep 2
# Reproduce the FIRST fatal: this is New Battle CONFIRM, not merely the menu.
printf 'CLICK 116 368\n' >> "$dest/commands.txt"
sleep 2
# Reproduce the new first fatal: Briefing CONFIRM enters tactical inventory.
# Game mailbox uses engine-local 640x400 coordinates, not screenshot pixels.
printf 'CLICK 320 346\n' >> "$dest/commands.txt"
wait "$debugger" || echo "gdb_exit=$?" > "$dest/gdb-exit.txt"
# A missing backtrace is an open diagnostic, never treated as engine pass.
grep -E 'Program received signal|Segmentation fault|^#([0-9]+) ' "$dest/gdb.log" \
  > "$dest/stack-summary.txt" || true
