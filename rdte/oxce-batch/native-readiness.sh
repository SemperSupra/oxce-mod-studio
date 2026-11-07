#!/usr/bin/env bash
# Source-only event readiness; never extrapolate a menu event into gameplay proof.
# Call only for the first occurrence of a screen state in each fresh native run.
oxce_wait_screen() {
  local state="${1:-}" pid="${2:-}" trace="${3:-}" seconds="${4:-8}"
  local signature i
  [[ "$state" =~ ^[A-Za-z]+State$ ]] || return 2
  [[ "$pid" =~ ^[0-9]+$ ]] && (( pid > 0 )) || return 2
  [[ "$seconds" =~ ^[0-9]+$ ]] && (( seconds >= 1 && seconds <= 30 )) || return 2
  [[ -n "$trace" && ! -L "$trace" ]] || return 5
  signature="\"event\":\"screen-class\",\"class\":\"N8OpenXcom${#state}${state}E\""
  for ((i = 0; i < seconds * 5; i++)); do
    [[ ! -L "$trace" ]] || return 5
    if [[ -f "$trace" ]] && grep -Fq -- "$signature" "$trace"; then
      return 0
    fi
    if ! kill -0 "$pid" 2>/dev/null; then
      return 3
    fi
    sleep 0.2
  done
  return 4
}
