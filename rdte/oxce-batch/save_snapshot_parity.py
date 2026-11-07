#!/usr/bin/env python3
"""Read-only comparison of an OXCE manual save against last pre-reload native snapshot.

Does NOT demonstrate loaded state restoration, game completion, stock-engine parity,
or human acceptance. Uses distro PyYAML, not a bespoke OXCE YAML parser.
"""
import argparse
import hashlib
import json
from pathlib import Path
import sys

try:
    import yaml
except ImportError as exc:
    raise SystemExit('PyYAML required: install distro python3-yaml') from exc

MAX_SAVE_BYTES = 64 * 1024 * 1024
MAX_TRACE_BYTES = 4 * 1024 * 1024
MAX_FRAMES = 4096
MAX_UNITS = 2048


def _read_regular(path, limit):
    path = Path(path)
    st = path.lstat()
    if not path.is_file() or path.is_symlink() or st.st_size < 1 or st.st_size > limit:
        raise ValueError(f'untrusted, empty or oversized evidence file: {path.name}')
    content = path.read_bytes()
    if len(content) != st.st_size:
        raise ValueError('evidence changed during read')
    return content


def _unit(u, saved=False):
    if not isinstance(u, dict):
        raise ValueError('unit must be a mapping')
    uid = u.get('id')
    faction = u.get('faction')
    if not isinstance(uid, int) or isinstance(uid, bool) or uid < 0:
        raise ValueError('invalid unit ID')
    if not isinstance(faction, int) or isinstance(faction, bool) or faction not in (0, 1, 2):
        raise ValueError('invalid faction')
    if saved:
        pos = u.get('position')
        hp = u.get('health')
    else:
        pos = [u.get('x'), u.get('y'), u.get('z')]
        hp = u.get('hp')
    if not isinstance(pos, list) or len(pos) != 3 or not all(type(x) is int and 0 <= x < 1024 for x in pos):
        raise ValueError('invalid position')
    for val in (u.get('tu'), hp, u.get('status')):
        if type(val) is not int or val < 0 or val > 10000:
            raise ValueError('invalid TU/health/status')
    return uid, (faction, tuple(pos), u['tu'], hp, u['status'])


def _units(items, saved=False):
    if not isinstance(items, list) or not 1 <= len(items) <= MAX_UNITS:
        raise ValueError('missing or excessive units')
    out = {}
    for u in items:
        uid, state = _unit(u, saved)
        if uid in out:
            raise ValueError('duplicate unit ID')
        out[uid] = state
    return out


def evaluate(save_path, native_trace_path):
    save_bytes = _read_regular(save_path, MAX_SAVE_BYTES)
    trace_bytes = _read_regular(native_trace_path, MAX_TRACE_BYTES)
    docs = list(yaml.safe_load_all(save_bytes))
    if len(docs) != 2 or not all(isinstance(d, dict) for d in docs):
        raise ValueError('OXCE save must contain exactly two YAML mapping documents')
    header, game = docs
    if header.get('name') != '_quick_.asav' or not isinstance(game.get('battleGame'), dict):
        raise ValueError('not an OXCE quick-save with battleGame')
    saved_battle = game['battleGame']
    saved_units = _units(saved_battle.get('units'), saved=True)
    raw_lines = trace_bytes.decode('utf-8').splitlines()
    if not 1 <= len(raw_lines) <= MAX_FRAMES:
        raise ValueError('missing or excessive native battle telemetry')
    frames = []
    for idx, line in enumerate(raw_lines):
        try:
            event = json.loads(line)
        except (json.JSONDecodeError, TypeError) as exc:
            raise ValueError(f'malformed native battle line {idx + 1}') from exc
        if not isinstance(event, dict) or event.get('event') != 'battle-snapshot':
            raise ValueError(f'unexpected native event at line {idx + 1}')
        if type(event.get('turn')) is not int or type(event.get('side')) is not int:
            raise ValueError('invalid native turn/side')
        if event['turn'] < 0 or event['side'] not in (0, 1, 2):
            raise ValueError('invalid native turn/side range')
        # Validate all frames, not just the last one, to fail closed on corruption.
        frames.append((event['turn'], event['side'], _units(event.get('units'))))
    last_turn, last_side, native_units = frames[-1]
    saved_turn = saved_battle.get('turn')
    saved_side = saved_battle.get('side')
    if type(saved_turn) is not int:
        raise ValueError('invalid saved battle turn')
    # Pinned OXCE 8.6.5 quick-saves may omit battleGame.side entirely.
    # Never fabricate saved side parity from the native trace.
    side_serialized = 'side' in saved_battle
    if side_serialized and type(saved_side) is not int:
        raise ValueError('invalid saved battle side')
    missing = sorted(set(native_units) - set(saved_units))
    extra = sorted(set(saved_units) - set(native_units))
    differences = sorted(uid for uid in set(saved_units) & set(native_units)
                         if saved_units[uid] != native_units[uid])
    side_admitted = saved_side == last_side if side_serialized else last_side == 0
    qualified = (saved_turn == last_turn and side_admitted
                 and not missing and not extra and not differences)
    return {
        'schema': 'oxce.public.native-save-snapshot-parity.v1',
        'evidenceClass': 'REAL_OXCE_QUICKSAVE_VS_PRE_RELOAD_LAST_SNAPSHOT_ONLY',
        'saveSha256': hashlib.sha256(save_bytes).hexdigest(),
        'nativeTraceSha256': hashlib.sha256(trace_bytes).hexdigest(),
        'snapshots': len(frames), 'nativeUnits': len(native_units),
        'savedUnits': len(saved_units), 'savedTurn': saved_turn, 'nativeTurn': last_turn,
        'savedSide': saved_side, 'nativeSide': last_side,
        'savedSideSerialized': side_serialized, 'savedSideCompared': side_serialized,
        'missingSavedUnitIds': missing, 'extraSavedUnitIds': extra,
        'differentUnitIds': differences,
        'savedSnapshotEqualsLastNativeSnapshot': qualified,
        'loadedStateSemanticsQualified': False,
        'gameplayQualified': False, 'stockEngineQualified': False,
        'originalRetailGameQualified': False, 'humanApproved': False,
        'limitation': ('A matching pre-reload snapshot checks what F5 wrote (side is not necessarily serialized), not what F9 restored; '
                       'post-reload native state evidence is still required.'),
    }


def main(argv=None):
    p = argparse.ArgumentParser()
    p.add_argument('save_path')
    p.add_argument('native_trace_path')
    p.add_argument('receipt_path')
    args = p.parse_args(argv)
    try:
        result = evaluate(args.save_path, args.native_trace_path)
        path = Path(args.receipt_path)
        with path.open('x', encoding='utf-8') as f:
            json.dump(result, f, indent=2, sort_keys=True)
            f.write('\n')
        print(json.dumps(result, sort_keys=True))
        return 0 if result['savedSnapshotEqualsLastNativeSnapshot'] else 1
    except (OSError, UnicodeError, yaml.YAMLError, ValueError) as exc:
        print('SAVE_SNAPSHOT_EVIDENCE_FAIL: ' + str(exc), file=sys.stderr)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
