#!/usr/bin/env bash
# Runs inside one isolated Xvfb display; same verified executable, separate users.
set -u
package="$1"; scenario="$2"; target="$3"
mkdir -p "$target"
cp -a "$package/user" "$target/user"
cp -a "$package/config" "$target/config"
if [ "${SF_RESTORE_COMPARE:-0}" = 1 ]; then
  test "${SF_SAVE_RELOAD:-0}" = 1 || { echo 'RESTORE_PROBE_REQUIRES_QUICKSAVE' >&2; exit 2; }
  test ! -e "$target/user/rdte-synthetic-master/_quick_.asav" || {
    echo 'STALE_QUICKSAVE_PRESENT_BEFORE_NATIVE_PROBE' >&2; exit 3;
  }
fi
: > "$target/commands.txt"
chmod 600 "$target/commands.txt"
OXCE_RDTE_MAILBOX="$PWD/$target/commands.txt" \
OXCE_RDTE_TRACE="$PWD/$target/trace.jsonl" \
OXCE_RDTE_BATTLE="$PWD/$target/battle.jsonl" \
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
    if [ "${SF_NEXT_TURN_CONTINUE:-0}" = 1 ]; then
      sleep 2
      # Pinned NextTurnState::handle closes on a real mouse-down, any point.
      if [ "${SF_NEXT_TURN_KEYBOARD:-0}" = 1 ]; then
        # SDL 1.2 SDLK_RETURN=13; NextTurnState::handle closes on key down.
        printf 'KEY 13\n' >> "$target/commands.txt"
      else
        printf 'CLICK 320 200\n' >> "$target/commands.txt"
      fi
      if [ "${SF_PLAYER_MOVE:-0}" = 1 ]; then
        sleep 2
        # Native first-action probe: click vacant (1,2) next to player at (2,2).
        # Confirm actual position/TU from engine data, never command text alone.
        printf 'CLICK 320 204\n' >> "$target/commands.txt"
        sleep 3
      fi
      if [ "${SF_SAVE_RELOAD:-0}" = 1 ]; then
        sleep 2
        # Actual SDL1 F5/286; only native SaveGameState and save bytes qualify.
        printf 'KEY 286\n' >> "$target/commands.txt"
        sleep 3
        if [ "${SF_RESTORE_COMPARE:-0}" = 1 ]; then
          # The controlled session must start without a stale manual quicksave.
          quick="$target/user/rdte-synthetic-master/_quick_.asav"
          test -s "$quick" && test ! -L "$quick" || { echo 'MISSING_FRESH_FIRST_QUICKSAVE' >&2; exit 3; }
          cp "$quick" "$target/pre-f9-quicksave.asav" || exit 3
          before_stamp="$(stat -c '%y' "$quick")"
          # Mutate the *real* battle by ending the player's turn. Await the
          # second observed NextTurnState rather than relying on a fixed sleep.
          printf 'CLICK 510 304\n' >> "$target/commands.txt"
          second_overlay=0
          for attempt in 1 2 3 4 5 6 7 8 9 10 11 12 13 14; do
            count="$(grep -c 'NextTurnState' "$target/trace.jsonl" 2>/dev/null || true)"
            if [ "${count:-0}" -ge 2 ]; then second_overlay=1; break; fi
            sleep 1
          done
          test "$second_overlay" = 1 || { echo 'NO_INTERVENING_NATIVE_TURN_OVERLAY' >&2; exit 4; }
          printf 'CLICK 320 200\n' >> "$target/commands.txt"
          sleep 2
        fi
        # Original engine F9/290 quickloads the still-intact first quicksave.
        printf 'KEY 290\n' >> "$target/commands.txt"
        sleep 4
        if [ "${SF_RESTORE_COMPARE:-0}" = 1 ]; then
          # Reserialize the restored state. Never rely on the F9 command alone.
          printf 'KEY 286\n' >> "$target/commands.txt"
          sleep 3
          test -s "$quick" && test ! -L "$quick" || { echo 'MISSING_POST_RELOAD_QUICKSAVE' >&2; exit 5; }
          after_stamp="$(stat -c '%y' "$quick")"
          test "$before_stamp" != "$after_stamp" || { echo 'QUICKSAVE_NOT_REWRITTEN' >&2; exit 5; }
          cp "$quick" "$target/post-f9-quicksave.asav" || exit 5
          printf 'before=%s\nafter=%s\n' "$before_stamp" "$after_stamp" > "$target/save-write-times.txt"
        fi
      fi
      if [ "${SF_ABORT_CONFIRM:-0}" = 1 ]; then
        sleep 2
        # Frozen BattlescapeState buttons: Abort at (x+240,y+16).
        # 320x200 logical interface -> 640x400 mailbox coordinates.
        printf 'CLICK 510 336\n' >> "$target/commands.txt"
        sleep 2
        # Frozen AbortMissionState: confirm button x16..136,y110..126.
        printf 'CLICK 100 236\n' >> "$target/commands.txt"
        sleep 3
      fi
      if [ "${SF_BATTLE_END_TURN:-0}" = 1 ]; then
        sleep 3
        # 320x200 logical battlescape: EndTurn (x+240,y)=(240,144).
        # 640x400 mailbox input scales logical coordinates by two.
        printf 'CLICK 510 304\n' >> "$target/commands.txt"
        if [ "${SF_SECOND_NEXT_TURN_CONTINUE:-0}" = 1 ]; then
          sleep 2
          # Actual NextTurnState::handle dismisses on any mouse-down.
          printf 'CLICK 320 200\n' >> "$target/commands.txt"
        fi
      fi
    fi
  fi ;;
 new-game-confirm) printf 'CLICK 220 200\n' >> "$target/commands.txt"; sleep 2; printf 'CLICK 238 344\n' >> "$target/commands.txt" ;;
 *) echo 'invalid scenario' >&2; exit 2 ;;
 esac
fi
sleep 5
import -display "$DISPLAY" -window root "$target/screen.png" 2>"$target/screenshot-error.txt" || true
kill -TERM "$pid" 2>/dev/null || true
wait "$pid" 2>/dev/null || true
