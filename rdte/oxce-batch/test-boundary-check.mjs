#!/usr/bin/env node
// Pure synthetic negative controls: not native game acceptance.
import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { evaluateBoundary } from './boundary-check.mjs';
function fixture(t,kind='save-load') {
  const root=fs.mkdtempSync(path.join(os.tmpdir(),'oxce-boundary-'));
  t.after(()=>fs.rmSync(root,{recursive:true,force:true}));
  const user=path.join(root,'user','rdte-synthetic-master');fs.mkdirSync(user,{recursive:true});
  fs.writeFileSync(path.join(root,'user','openxcom.log'),'[INFO] OXCE startup\n');
  const states=kind==='save-load'?
    ['MainMenuState','BattlescapeState','SaveGameState','BattlescapeState','LoadGameState','BattlescapeState']:
    ['MainMenuState','BattlescapeState','AbortMissionState','DebriefingState'];
  fs.writeFileSync(path.join(root,'trace.jsonl'), states.map(s=>JSON.stringify({event:'screen-class',class:'N8OpenXcom'+s})).join('\n')+'\n');
  fs.writeFileSync(path.join(root,'commands.txt'),kind==='save-load'?'KEY 286\nKEY 290\n':'CLICK 510 336\nCLICK 100 236\n');
  fs.writeFileSync(path.join(user,'_quicksave_.sav'), 'X'.repeat(2048));
  return root;
}
test('manual save/load UI and real save bytes evidence, no full gameplay promotion',t=>{
  const r=evaluateBoundary(fixture(t),'save-load');
  assert.equal(r.saveReloadQualified,true);assert.equal(r.savedStateSemanticsQualified,false);
  assert.equal(r.gameplayQualified,false);assert.equal(r.manualSaves.length,1);
});
test('abort confirmation reaches a debrief without gameplay promotion',t=>{
  const r=evaluateBoundary(fixture(t,'abort'),'abort');
  assert.equal(r.abortDebriefQualified,true);assert.equal(r.gameplayQualified,false);
});
test('commands alone without state evidence cannot qualify',t=>{
  const root=fixture(t);let lines=fs.readFileSync(path.join(root,'trace.jsonl'),'utf8').split('\n').filter(Boolean);
  lines=lines.filter(x=>!x.includes('LoadGameState'));fs.writeFileSync(path.join(root,'trace.jsonl'),lines.join('\n')+'\n');
  assert.equal(evaluateBoundary(root,'save-load').saveReloadQualified,false);
});
test('only autobattle save never qualifies manual save',t=>{
  const root=fixture(t);fs.unlinkSync(path.join(root,'user/rdte-synthetic-master/_quicksave_.sav'));
  fs.writeFileSync(path.join(root,'user/rdte-synthetic-master/_autobattle_.asav'),'Y'.repeat(2000));
  assert.equal(evaluateBoundary(root,'save-load').saveReloadQualified,false);
});
test('missing F9 cannot qualify even with UI state',t=>{
  const root=fixture(t);fs.writeFileSync(path.join(root,'commands.txt'),'KEY 286\n');
  assert.equal(evaluateBoundary(root,'save-load').saveReloadQualified,false);
});
test('no abort debrief must fail despite confirmation',t=>{
  const root=fixture(t,'abort');fs.writeFileSync(path.join(root,'trace.jsonl'),
    JSON.stringify({event:'screen-class',class:'BattlescapeState'})+'\n'+JSON.stringify({event:'screen-class',class:'AbortMissionState'})+'\n');
  assert.equal(evaluateBoundary(root,'abort').abortDebriefQualified,false);
});
test('native fatal prevents acceptance',t=>{
  const root=fixture(t);fs.appendFileSync(path.join(root,'user/openxcom.log'),'[FATAL] Test failure\n');
  assert.equal(evaluateBoundary(root,'save-load').saveReloadQualified,false);
});
test('malformed trace fails closed',t=>{
  const root=fixture(t);fs.writeFileSync(path.join(root,'trace.jsonl'),'{NOT JSON}\n');
  assert.throws(()=>evaluateBoundary(root,'save-load'),/malformed native trace/);
});
test('symlinked manual save rejected',t=>{
  const root=fixture(t);const target=path.join(root,'user/rdte-synthetic-master/_quicksave_.sav');
  fs.symlinkSync(target,path.join(root,'user/rdte-synthetic-master/other.sav'));
  assert.throws(()=>evaluateBoundary(root,'save-load'),/untrusted save tree object/);
});
test('missing evidence rejects instead of inventing an accepted result',t=>{
  const root=fixture(t);fs.unlinkSync(path.join(root,'trace.jsonl'));
  assert.throws(()=>evaluateBoundary(root,'save-load'));
});
