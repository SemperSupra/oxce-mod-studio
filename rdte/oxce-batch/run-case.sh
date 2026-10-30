#!/usr/bin/env bash
# Runs inside one isolated Xvfb display; same verified executable, separate users.
set -u
package="$1"; scenario="$2"; target="$3"
mkdir -p "$target"
cp -a "$package/user" "$target/user"
cp -a "$package/config" "$target/config"
: > "$target/commands.txt"
chmod 600 "$target/commands.txt"
OXCE_RDTE_MAILBOX="$PWD/$target/commands.txt" \
OXCE_RDTE_TRACE="$PWD/$target/trace.jsonl" \
SDL_AUDIODRIVER=dummy "$package/openxcom-sf-rdte" \
 -data "$PWD/$package/data" -user "$PWD/$target/user" -config "$PWD/$target/config" \
 > "$target/engine-stdout.log" 2>&1 &
pid=$!
sleep 5
if kill -0 "$pid" 2>/dev/null; then
 case "$scenario" in
 boot) ;;
 new-game) printf 'CLICK 220 200\n' >> "$target/commands.txt" ;;
 new-battle) printf 'CLICK 420 200\n' >> "$target/commands.txt" ;;
 new-battle-confirm) printf 'CLICK 420 200\n' >> "$target/commands.txt"; sleep 2; printf 'CLICK 116 368\n' >> "$target/commands.txt" ;;
 new-battle-briefing-continue)
  printf 'CLICK 420 200\n' >> "$target/commands.txt"; sleep 2
  printf 'CLICK 116 368\n' >> "$target/commands.txt"; sleep 2
  printf 'CLICK 320 346\n' >> "$target/commands.txt"
  if [ "${SF_INVENTORY_CONFIRM:-0}" = 1 ]; then
    sleep 2
    # Pinned InventoryState::_btnOk: x237..272, y1..23 on 320x200
    # coordinate scale 2 in original modified 640x400 mailbox.
    printf 'CLICK 510 22\n' >> "$target/commands.txt"
  fi ;;
 new-game-confirm) printf 'CLICK 220 200\n' >> "$target/commands.txt"; sleep 2; printf 'CLICK 238 344\n' >> "$target/commands.txt" ;;
 *) echo 'invalid scenario' >&2; exit 2 ;;
 esac
fi
sleep 5
import -display "$DISPLAY" -window root "$target/screen.png" 2>"$target/screenshot-error.txt" || true
kill -TERM "$pid" 2>/dev/null || true
wait "$pid" 2>/dev/null || true
