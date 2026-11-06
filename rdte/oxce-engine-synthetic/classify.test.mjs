#!/usr/bin/env node
// Actual classifier CLI contract, no native-game or licensed inputs.
import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync, readFileSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const classifier=fileURLToPath(new URL('./classify.mjs',import.meta.url));
const goodLog='Active mods:\n - rdte-synthetic-master\nLoading rulesets...\nLoading rulesets done.\nLoading ended.\n';
const goodPhase='configure_exit=0\nbuild_exit=0\nnative_launch_attempted=true\nengine_launch_exit=124\n';
function check(label,log,phase,expected) {
  const dir=mkdtempSync(join(tmpdir(),'oxce-classify-'));
  try {
    writeFileSync(join(dir,'openxcom.log'),log);
    writeFileSync(join(dir,'exit.env'),phase);
    const result=spawnSync(process.execPath,[classifier,dir],{encoding:'utf8',timeout:7000});
    assert.equal(result.status,expected?0:1,label+': CLI exit; '+result.stderr);
    const v=JSON.parse(readFileSync(join(dir,'classification.json'),'utf8'));
    assert.equal(v.nativeLoaderQualified,expected,label+': native load grade');
    assert.equal(v.engineAcceptance,false);
    assert.equal(v.gameplayQualified,false);
    return v;
  } finally {rmSync(dir,{recursive:true,force:true});}
}
test('complete synthetic engine load + watchdog is loader-only, not gameplay',()=>{
  const v=check('good',goodLog,goodPhase,true);
  assert.equal(v.actualEngineBuilt,true);
  assert.equal(v.nativeLaunchAttempted,true);
  assert.equal(v.exitStatus,'124');
  assert.equal(v.firstError,null);
});
test('late engine ERROR after load marker remains RED',()=>{
  const v=check('error',goodLog+'[ERROR] Sound 1 in GEO.CAT not found\n',goodPhase,false);
  assert.match(v.firstError,/Sound 1/);
});
test('fatal resource error remains RED',()=>check('fatal',goodLog+'[FATAL] missing resource\n',goodPhase,false));
test('missing ruleset completion and fake truncated trace remain RED',()=>{
  check('missing-rules',goodLog.replace('Loading rulesets done.','truncated'),goodPhase,false);
  check('missing-load',goodLog.replace('Loading ended.','Starting...'),goodPhase,false);
  check('fake-active',goodLog.replace('rdte-synthetic-master','other-master'),goodPhase,false);
});
test('compile absent, launch absent, and crash-style exit remain RED',()=>{
  check('no-build',goodLog,goodPhase.replace('build_exit=0','build_exit=1'),false);
  check('no-launch',goodLog,goodPhase.replace('native_launch_attempted=true','native_launch_attempted=false'),false);
  check('crash',goodLog,goodPhase.replace('engine_launch_exit=124','engine_launch_exit=139'),false);
});
