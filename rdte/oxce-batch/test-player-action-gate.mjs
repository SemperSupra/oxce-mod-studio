import test from 'node:test';
import assert from 'node:assert/strict';
import { classifyPlayerAction } from './player-action-gate.mjs';

const native = {
  GameplayQualified:false,NativeSnapshots:12,MalformedFrames:0,
  PlayerObserved:true,EnemyObserved:true,PlayerMoveObserved:true,
  PlayerTUDecreased:true,PlayerMovementSpentTU:true,
  HostileFactionTurnObserved:true,ReturnedToPlayerTurn:true,
  EnemyActionObserved:false,FirstFailure:'no-native-enemy-position-or-tu-delta'
};
const commands='CLICK 320 204\nCLICK 510 304\n';
const log='[INFO] OpenXcom started successfully!\n';
const trace='{"event":"battle-snapshot","turn":1}\n';
function check(mod={},c=commands,l=log,t=trace){return classifyPlayerAction({...native,...mod},c,l,t)}
test('actual split hostile gate: player move/TU accepted without fake enemy action',()=>{
  const r=check();assert.equal(r.playerActionQualified,true);assert.equal(r.hostileActionInThisTrial,false);assert.equal(r.gameplayQualified,false);assert.equal(r.firstFailure,'');
});
test('independently observed hostile action also permitted',()=>{
  const r=check({FirstFailure:'',EnemyActionObserved:true});assert.equal(r.playerActionQualified,true);assert.equal(r.hostileActionInThisTrial,true);
});
test('enemy-only failure with contradictory enemy=true is refused',()=>{
  assert.equal(check({EnemyActionObserved:true}).playerActionQualified,false);
});
test('a missing or unrelated first failure is never swallowed',()=>{
  for(const FirstFailure of ['malformed-native-snapshot:1','missing-opposing-factions','native-telemetry-unavailable','engine-exit-124',null]){
    assert.equal(check({FirstFailure}).playerActionQualified,false,String(FirstFailure));
  }
});
test('no player movement or TU or movement-cost pairing is rejected',()=>{
  for(const key of ['PlayerMoveObserved','PlayerTUDecreased','PlayerMovementSpentTU','PlayerObserved']){
    assert.equal(check({[key]:false}).playerActionQualified,false,key);
  }
});
test('no hostile turn or return or opposing faction is refused',()=>{
  for(const key of ['HostileFactionTurnObserved','ReturnedToPlayerTurn','EnemyObserved']){
    assert.equal(check({[key]:false}).playerActionQualified,false,key);
  }
});
test('malformed or insufficient snapshots are refused',()=>{
  for(const mod of [{NativeSnapshots:0},{MalformedFrames:1},{NativeSnapshots:2}])assert.equal(check(mod).playerActionQualified,false);
});
test('the actual click and battle trace are required',()=>{
  assert.equal(check({},'CLICK 510 304\n').playerActionQualified,false);
  assert.equal(check({},commands,log,'').playerActionQualified,false);
});
test('native fatal is refused regardless of otherwise positive action',()=>{
  assert.equal(check({},commands,'[FATAL] missing terrain\n').playerActionQualified,false);
});
test('gameplay acceptance cannot be asserted',()=>{
  assert.equal(check({GameplayQualified:true}).playerActionQualified,false);
});
