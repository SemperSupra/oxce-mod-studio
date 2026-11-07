#!/usr/bin/env bash
# Source-only hostile/positive checks for a bounded native state-event waiter.
set -euo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/native-readiness.sh"
root="$(mktemp -d)"
child=''
trap '[[ -z "$child" ]] || kill "$child" 2>/dev/null || true; rm -rf "$root"' EXIT
failed=0; passed=0
expect() {
  local expected="$1" name="$2"
  shift 2
  local actual=0
  "$@" || actual=$?
  if [[ "$actual" != "$expected" ]];then
    printf 'RED: %s expected %s got %s\n' "$name" "$expected" "$actual" >&2
    failed=$((failed+1))
  else
    passed=$((passed+1))
  fi
}
sleep 10 & child=$!
trace="$root/trace.jsonl"
printf '{"event":"screen-class","class":"N8OpenXcom17GoToMainMenuStateE"}\n' > "$trace"
expect 4 false-prefix-and-timeout oxce_wait_screen MainMenuState "$child" "$trace" 1
printf '{"event":"screen-class","class":"N8OpenXcom13MainMenuStateE"}\n' >> "$trace"
expect 0 exact-native-event oxce_wait_screen MainMenuState "$child" "$trace" 1
printf '{"event":"wrong-event","class":"N8OpenXcom14NewBattleStateE"}\n' >> "$trace"
expect 4 rejects-wrong-event oxce_wait_screen NewBattleState "$child" "$trace" 1
( sleep 0.4; printf '{"event":"screen-class","class":"N8OpenXcom14NewBattleStateE"}\n' >> "$trace" ) &
expect 0 async-event oxce_wait_screen NewBattleState "$child" "$trace" 3
expect 2 invalid-state oxce_wait_screen '../etc/passwd' "$child" "$trace" 1
expect 2 invalid-timeout oxce_wait_screen MainMenuState "$child" "$trace" 31
expect 2 invalid-pid oxce_wait_screen MainMenuState 'untrusted' "$trace" 1
ln -s "$trace" "$root/linked.jsonl"
expect 5 symlink-trace oxce_wait_screen MainMenuState "$child" "$root/linked.jsonl" 1
kill "$child" 2>/dev/null || true
wait "$child" 2>/dev/null || true
child=''
expect 3 dead-process oxce_wait_screen InventoryState 999999 "$trace" 1
if [[ "$failed" -ne 0 || "$passed" -ne 9 ]];then
  echo "RED: native readiness negative controls passed=$passed failed=$failed" >&2
  exit 1
fi
printf 'native-readiness source-only controls: %s PASS, no native engine acceptance\n' "$passed"
