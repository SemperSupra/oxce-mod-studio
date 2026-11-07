#!/usr/bin/env python3
"""First-party public synthetic OXCE native snap-shot experiment.

Calibrated screen clicks are hypotheses; a real on-run save+trace is the oracle.
This script never grants gameplay, stock-engine, victory or human approval.
"""
import argparse
import hashlib
import json
from pathlib import Path
import sys
import yaml

MAX_SAVE = 64 * 1024 * 1024
MAX_TRACE = 4 * 1024 * 1024
SHOOTER_ID = 1
HOSTILE_ID = 1000000
CALIBRATED_TILE = (2, 3, 0)
CALIBRATED_MAILBOX = (320, 204)  # real player-walk click from B66
SOURCE_TILE_WIDTH = 32  # pinned Camera.cpp projection, 2x engine-local mailbox
EXPECTED_PLAYER = (2, 2, 0)
EXPECTED_HOSTILE = (5, 5, 0)


def plain_file(path, limit):
    p = Path(path)
    st = p.lstat()
    if not st or not p.is_file() or p.is_symlink() or not 1 <= st.st_size <= limit:
        raise ValueError('missing, unsafe or oversized native evidence')
    b = p.read_bytes()
    if len(b) != st.st_size:
        raise ValueError('native evidence changed during read')
    return b


def save_game(p):
    raw = plain_file(p, MAX_SAVE)
    docs = list(yaml.safe_load_all(raw))
    if len(docs) != 2 or any(not isinstance(d, dict) for d in docs):
        raise ValueError('expected two genuine OXCE save documents')
    b = docs[1].get('battleGame')
    if not isinstance(b, dict) or not isinstance(b.get('units'), list) or not isinstance(b.get('items'), list):
        raise ValueError('no native battle units or items')
    ids = [u.get('id') for u in b['units'] if isinstance(u, dict)]
    if len(ids) != len(b['units']) or len(set(ids)) != len(ids):
        raise ValueError('invalid or duplicate unit IDs')
    return b, hashlib.sha256(raw).hexdigest()


def unit(b, unit_id):
    matches = [u for u in b['units'] if u.get('id') == unit_id]
    if len(matches) != 1: raise ValueError(f'missing/duplicate unit {unit_id}')
    u = matches[0]
    if not isinstance(u.get('position'), list) or len(u['position']) != 3:
        raise ValueError('missing native unit position')
    if any(type(u.get(k)) is not int for k in ('tu', 'health', 'faction', 'status')):
        raise ValueError('invalid native unit state')
    return u


def magazine(b):
    items = b['items']
    guns = [i for i in items if isinstance(i, dict) and i.get('type') == 'STR_RDTE_TOOL' and i.get('owner') == SHOOTER_ID]
    if len(guns) != 1 or guns[0].get('inventoryslot') != 'STR_RIGHT_HAND':
        raise ValueError('player right-hand weapon not equipped exactly once')
    weapon = guns[0]
    clip_id = weapon.get('ammoItem')
    if type(clip_id) is not int: raise ValueError('missing loaded clip link')
    clips = [i for i in items if isinstance(i, dict) and i.get('id') == clip_id and i.get('type') == 'STR_RDTE_CELL']
    if len(clips) != 1 or type(clips[0].get('ammoqty')) is not int:
        raise ValueError('missing unique loaded magazine with count')
    return clip_id, clips[0]['ammoqty']


def target_from_calibration(b):
    player, hostile = unit(b, SHOOTER_ID), unit(b, HOSTILE_ID)
    if tuple(player['position']) != EXPECTED_PLAYER or tuple(hostile['position']) != EXPECTED_HOSTILE:
        raise ValueError('drift from independently observed B66 native camera/mission calibration')
    if player['faction'] != 0 or hostile['faction'] != 1 or b.get('selectedUnit') != SHOOTER_ID or b.get('turn') != 1:
        raise ValueError('wrong selected unit, faction or starting turn')
    clip_id, ammo = magazine(b)
    if ammo < 1: raise ValueError('empty native magazine')
    dx, dy = hostile['position'][0]-CALIBRATED_TILE[0], hostile['position'][1]-CALIBRATED_TILE[1]
    x = CALIBRATED_MAILBOX[0] + (dx-dy)*SOURCE_TILE_WIDTH
    y = CALIBRATED_MAILBOX[1] + (dx+dy)*(SOURCE_TILE_WIDTH//2)
    if not (24 <= x < 616 and 0 <= y < 288):
        raise ValueError('derived target outside native tactical map; do not send click')
    return {'targetMailboxX':x,'targetMailboxY':y,'targetTile':hostile['position'],
            'calibrationTile':list(CALIBRATED_TILE),'calibrationClick':list(CALIBRATED_MAILBOX),
            'loadedAmmo':ammo,'clipId':clip_id,'sourceOnlyTargetHypothesis':True}


def assess(folder):
    root = Path(folder)
    b0, sha0 = save_game(root/'pre-shot.asav')
    b1, sha1 = save_game(root/'post-shot.asav')
    calibrated = target_from_calibration(b0)
    lines = plain_file(root/'trace.jsonl',MAX_TRACE).decode('utf-8').splitlines()
    events = [json.loads(line) for line in lines]
    if not 3 <= len(events) <= 4096 or any(e.get('event')!='screen-class' or not isinstance(e.get('class'),str) for e in events):
        raise ValueError('invalid native screen trace envelopes')
    states = [e['class'].split('OpenXcom')[-1] for e in events]
    menu = next((i for i,s in enumerate(states) if 'ActionMenuState' in s),-1)
    saves = [i for i,s in enumerate(states) if 'SaveGameState' in s]
    commands = plain_file(root/'commands.txt',65536).decode('utf-8').splitlines()
    wanted = ['KEY 286','KEY 101','KEY 50',f"CLICK {calibrated['targetMailboxX']} {calibrated['targetMailboxY']}",'KEY 286']
    idx=-1
    for token in wanted:
        idx=next((j for j in range(idx+1,len(commands)) if commands[j]==token),-1)
        if idx<0: raise ValueError('native action/shot/save command sequence missing or out of order')
    first_fatal = next((s for s in plain_file(root/'user'/'openxcom.log',MAX_TRACE).decode('utf-8').splitlines() if '[FATAL]' in s),'')
    u0,u1=unit(b0,SHOOTER_ID),unit(b1,SHOOTER_ID)
    h0,h1=unit(b0,HOSTILE_ID),unit(b1,HOSTILE_ID)
    a_id,a0=magazine(b0)
    b_id,a1=magazine(b1)
    correct_time=b0.get('turn')==1 and b1.get('turn')==1
    ordered_states=(len(saves)>=2 and saves[0]<menu<saves[-1] and
                    any('BattlescapeState' in s for s in states[menu+1:saves[-1]]))
    native_shot=not first_fatal and ordered_states and correct_time and a_id==b_id and a1==a0-1 and u1['tu']<u0['tu']
    native_damage=h1['health']<h0['health'] or (h0['status']==0 and h1['status']!=0)
    return {'schema':'oxce.synthetic.native-snap-shot.v1',
      'evidenceClass':'ACTUAL_MODIFIED_OXCE_SHOT_AND_DAMAGE_EXPERIMENT',
      'nativeActionMenuObserved':menu>=0,'nativeSaveStateCount':len(saves),
      'nativeFiringStateSequenceObserved':ordered_states,'sameTurn':correct_time,
      'playerTUBefore':u0['tu'],'playerTUAfter':u1['tu'],
      'loadedClipId':a_id,'ammoBefore':a0,'ammoAfter':a1,
      'enemyHealthBefore':h0['health'],'enemyHealthAfter':h1['health'],
      'enemyStatusBefore':h0['status'],'enemyStatusAfter':h1['status'],
      'nativeShotWithTUAndAmmoSpent':native_shot,
      'nativeEnemyDamageObserved':bool(native_shot and native_damage),
      'firstFatal':first_fatal,'baselineSha256':sha0,'afterSha256':sha1,
      'calibration':calibrated,'winningDebriefQualified':False,
      'gameplayQualified':False,'stockEngineQualified':False,
      'signalFrontierMvpQualified':False,'humanApproved':False}


def main(argv=None):
    p=argparse.ArgumentParser()
    p.add_argument('mode',choices=['target','qualify'])
    p.add_argument('source')
    p.add_argument('output')
    a=p.parse_args(argv)
    try:
        if a.mode=='target':
            data,_=save_game(a.source)
            result=target_from_calibration(data)
        else: result=assess(a.source)
        with open(a.output,'x',encoding='utf-8') as fp: json.dump(result,fp,sort_keys=True,indent=2);fp.write('\n')
        print(json.dumps(result,sort_keys=True))
        if a.mode=='qualify' and not result['nativeShotWithTUAndAmmoSpent']: return 1
        return 0
    except (ValueError,OSError,TypeError,KeyError,yaml.YAMLError,json.JSONDecodeError) as exc:
        print('NATIVE_SNAP_PROBE_RED: '+str(exc),file=sys.stderr)
        return 1

if __name__=='__main__': sys.exit(main())
