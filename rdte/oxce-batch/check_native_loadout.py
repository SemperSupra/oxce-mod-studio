#!/usr/bin/env python3
"""Read-only actual OXCE native quicksave armament-evidence classifier.

Requires authentic saved game model; never infers ammo loaded, actual shots,
mission victory, vanilla/stock parity or art/human acceptance.
"""
import argparse
import hashlib
import json
from pathlib import Path
import sys
import yaml

MAX_SAVE=64*1024*1024


def classify(save_path):
    p=Path(save_path)
    st=p.lstat()
    if not p.is_file() or p.is_symlink() or not 1<=st.st_size<=MAX_SAVE:
        raise ValueError('invalid original-only native save evidence')
    raw=p.read_bytes()
    if len(raw)!=st.st_size:raise ValueError('changed native save evidence')
    docs=list(yaml.safe_load_all(raw))
    if len(docs)!=2 or any(not isinstance(d,dict) for d in docs):
        raise ValueError('OXCE save must have two YAML mapping documents')
    battle=docs[1].get('battleGame')
    if not isinstance(battle,dict):raise ValueError('missing real battleGame')
    units=battle.get('units')
    items=battle.get('items')
    if not isinstance(units,list) or not 1<=len(units)<=2048 or not isinstance(items,list) or len(items)>4096:
        raise ValueError('invalid unit or inventory collection')
    players=set(); all_ids=set()
    for u in units:
        if not isinstance(u,dict) or type(u.get('id')) is not int or type(u.get('faction')) is not int:
            raise ValueError('invalid battle unit record')
        if u['id'] in all_ids: raise ValueError('duplicate native unit ID')
        all_ids.add(u['id'])
        if u['faction']==0:players.add(u['id'])
    if not players: raise ValueError('no player units present')
    player_tools=[]; player_cells=[]; other=[]
    for item in items:
        if not isinstance(item,dict) or type(item.get('id')) is not int or not isinstance(item.get('type'),str):
            raise ValueError('invalid native battle item')
        owner=item.get('owner')
        if owner is not None and (type(owner) is not int or owner not in all_ids):
            raise ValueError('unresolved native item owner')
        if item['type']=='STR_RDTE_TOOL' and owner in players: player_tools.append(item['id'])
        if item['type']=='STR_RDTE_CELL' and (owner in players or owner is None):player_cells.append(item['id'])
        if owner not in players:other.append(item['id'])
    qualified=bool(player_tools and player_cells)
    return {'schema':'oxce.original-only.native-armed-player.v1',
      'evidenceClass':'NATIVE_MODIFIED_OXCE_ITEM_OWNERSHIP_NOT_SHOOTING',
      'nativeSaveSha256':hashlib.sha256(raw).hexdigest(),
      'playerUnitCount':len(players),'nativeItemCount':len(items),
      'playerOwnedToolItemIds':sorted(player_tools),
      'playerAccessibleCellItemIds':sorted(player_cells),
      'playerLoadoutObserved':qualified,
      'loadedAmmoVerified':False,'weaponActionVerified':False,
      'enemyDamageVerified':False,'victoryQualified':False,
      'stockEngineQualified':False,'signalFrontierMvpQualified':False,'humanApproved':False}


def main(argv=None):
    p=argparse.ArgumentParser()
    p.add_argument('native_save');p.add_argument('receipt')
    a=p.parse_args(argv)
    try:
        r=classify(a.native_save)
        with Path(a.receipt).open('x',encoding='utf-8') as f:json.dump(r,f,indent=2,sort_keys=True);f.write('\n')
        print(json.dumps(r,sort_keys=True))
        return 0 if r['playerLoadoutObserved'] else 1
    except (ValueError,OSError,yaml.YAMLError) as exc:
        print('NATIVE_ARMED_PLAYER_FAIL: '+str(exc),file=sys.stderr);return 1

if __name__=='__main__':raise SystemExit(main())
