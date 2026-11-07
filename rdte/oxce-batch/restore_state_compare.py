#!/usr/bin/env python3
"""Original-only native quickload restoration subset proof after intervening turn.

Requires independently captured two quicksave versions and a real intervening
turn >=2 in native battle telemetry. Does not assert stock/OXCE retail parity,
complete game mechanics, mission victory, or human approval.
"""
import argparse
import hashlib
import json
from pathlib import Path
import sys

import yaml
from save_snapshot_parity import _read_regular, _units, MAX_SAVE_BYTES, MAX_TRACE_BYTES, MAX_FRAMES


def _save(path):
    raw = _read_regular(path, MAX_SAVE_BYTES)
    docs = list(yaml.safe_load_all(raw))
    if len(docs) != 2 or not all(type(x) is dict for x in docs):
        raise ValueError('not two-document OXCE quicksave')
    if docs[0].get('name') != '_quick_.asav':
        raise ValueError('wrong quicksave name')
    battle = docs[1].get('battleGame')
    if not isinstance(battle, dict) or type(battle.get('turn')) is not int:
        raise ValueError('missing battle turn')
    units = _units(battle.get('units'), saved=True)
    return raw, battle['turn'], units


def _class_trace(path):
    raw = _read_regular(path, MAX_TRACE_BYTES).decode('utf-8')
    lines = raw.splitlines()
    if not 1 <= len(lines) <= MAX_FRAMES:
        raise ValueError('missing/oversized state evidence')
    states = []
    for i,line in enumerate(lines):
        try: v=json.loads(line)
        except json.JSONDecodeError as exc: raise ValueError(f'malformed state {i}') from exc
        if not isinstance(v, dict) or v.get('event') != 'screen-class' or not isinstance(v.get('class'),str):
            raise ValueError('unexpected state event')
        states.append(v['class'])
    return states


def _native_turns(path):
    raw = _read_regular(path, MAX_TRACE_BYTES)
    lines = raw.decode('utf-8').splitlines()
    if not 1 <= len(lines) <= MAX_FRAMES:
        raise ValueError('missing/oversized battle evidence')
    turns=[]
    for i,line in enumerate(lines):
        try: v=json.loads(line)
        except json.JSONDecodeError as exc: raise ValueError(f'malformed native battle line {i}') from exc
        if not isinstance(v,dict) or v.get('event') != 'battle-snapshot' or type(v.get('turn')) is not int or type(v.get('side')) is not int:
            raise ValueError('invalid battle snapshot')
        _units(v.get('units'))
        if v['turn'] < 0 or v['side'] not in (0,1,2):
            raise ValueError('invalid battle side/turn')
        turns.append((v['turn'],v['side']))
    return turns


def evaluate(before_path, after_path, battle_path, screen_path, commands_path, engine_log_path):
    pre, preturn, preunits = _save(before_path)
    post, postturn, postunits = _save(after_path)
    turns = _native_turns(battle_path)
    states = _class_trace(screen_path)
    commands = _read_regular(commands_path,64 * 1024).decode('utf-8').splitlines()
    engine_log = _read_regular(engine_log_path, 2 * 1024 * 1024).decode('utf-8')
    fatal = next((ln for ln in engine_log.splitlines() if '[FATAL]' in ln),'')
    positions={}
    for i,state in enumerate(states):
        for name in ['BattlescapeState','LoadGameState','SaveGameState']:
            if name in state: positions.setdefault(name,[]).append(i)
    saves = positions.get('SaveGameState',[])
    loads = positions.get('LoadGameState',[])
    battles = positions.get('BattlescapeState',[])
    # Two actual native SaveGameState transitions: one before, another after load.
    state_sequence = bool(loads and saves and battles and
        any(s<loads[0] for s in saves) and
        any(b>loads[0] for b in battles) and
        any(s>loads[0] for s in saves))
    # Require real intervening second player turn before restoring a turn-1 save.
    intervening_turn = any(turn >= 2 and side == 0 for turn,side in turns)
    # Exact command sequence, not command presence alone; native states are separate.
    required=['KEY 286','CLICK 510 304','CLICK 320 200','KEY 290','KEY 286']
    it=iter(commands)
    command_sequence=all(any(value==need for value in it) for need in required)
    missing=sorted(set(preunits)-set(postunits))
    extra=sorted(set(postunits)-set(preunits))
    altered=sorted(uid for uid in set(preunits)&set(postunits) if preunits[uid]!=postunits[uid])
    match = preturn == postturn == 1 and not missing and not extra and not altered
    no_fatal = not fatal
    verified = bool(match and intervening_turn and state_sequence and command_sequence and no_fatal)
    return {
      'schema':'oxce.public.modified-native-restore-subset.v1',
      'evidenceClass':'ACTUAL_MODIFIED_ENGINE_QUICKSAVE_RESTORE_OBSERVABLE_UNITS_AFTER_INTERVENING_TURN',
      'beforeSha256':hashlib.sha256(pre).hexdigest(), 'afterSha256':hashlib.sha256(post).hexdigest(),
      'beforeTurn':preturn, 'afterTurn':postturn,
      'unitCountBefore':len(preunits),'unitCountAfter':len(postunits),
      'missingUnits':missing,'extraUnits':extra,'changedUnits':altered,
      'nativeSecondPlayerTurnObserved':intervening_turn,
      'nativeReloadSaveStateSequenceObserved':state_sequence,
      'boundedMailboxCommandOrderObserved':command_sequence,
      'firstFatal':fatal,
      'restoredObservableUnitStateSubsetQualified':verified,
      'fullSaveRestoreSemanticsQualified':False,
      'originalRetailGameQualified':False,'stockEngineQualified':False,
      'signalFrontierMvpQualified':False,'humanApproved':False,
      'limitation':'Compares selected unit, position, faction, TU, HP, status and battle turn across two real quicksaves with an intervening native turn, not all game-state internals.'
    }


def main(argv=None):
    p=argparse.ArgumentParser()
    for x in ['before_save','after_save','battle_trace','screen_trace','commands','engine_log','receipt']:
        p.add_argument(x)
    args=p.parse_args(argv)
    try:
        r=evaluate(args.before_save,args.after_save,args.battle_trace,args.screen_trace,args.commands,args.engine_log)
        with Path(args.receipt).open('x',encoding='utf-8') as fd: json.dump(r,fd,indent=2,sort_keys=True);fd.write('\n')
        print(json.dumps(r,sort_keys=True))
        return 0 if r['restoredObservableUnitStateSubsetQualified'] else 1
    except (OSError, UnicodeError, ValueError, yaml.YAMLError) as exc:
        print('NATIVE_RESTORED_SUBSET_FAIL: '+str(exc),file=sys.stderr)
        return 1

if __name__=='__main__':
    raise SystemExit(main())
