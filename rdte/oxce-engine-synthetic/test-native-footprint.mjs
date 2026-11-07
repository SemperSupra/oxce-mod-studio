#!/usr/bin/env node
// Test doubles validate the harness contract, never engine or game behavior.
import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { ablate } from './native-footprint.mjs';
const required=[
'config/options.cfg',
'user/mods/rdte-synthetic-master/metadata.yml',
'user/mods/rdte-synthetic-master/Ruleset/10-synthetic.rul',
'user/mods/rdte-synthetic-master/Language/en-US.yml',
'user/mods/rdte-synthetic-master/Language/rdte-fonts.yml',
'user/mods/rdte-synthetic-master/Language/rdte-fonts.png',
'data/SYNTHETIC/GEODATA/PALETTES.DAT',
'data/SYNTHETIC/SOUND/SAMPLE.CAT',
'data/SYNTHETIC/SOUND/SAMPLE2.CAT',
'data/SYNTHETIC/MAPS/SFREL00.MAP',
'data/SYNTHETIC/ROUTES/SFREL00.RMP',
'data/SYNTHETIC/TERRAIN/BLANKS.MCD'];
function setup(t){
 const d=fs.mkdtempSync(path.join(os.tmpdir(),'oxce-native-ablation-test-'));
 t.after(()=>fs.rmSync(d,{recursive:true,force:true}));
 const source=path.join(d,'src');fs.mkdirSync(source);
 for(const rel of [...required,'data/common/public.txt','data/standard/public.txt']){
  const dest=path.join(source,rel);fs.mkdirSync(path.dirname(dest),{recursive:true});fs.writeFileSync(dest,`synthetic:${rel}`);
 }
 return {d,source,work:path.join(d,'output')};
}
function syntheticExecutor(groupLog={}){
 return (_,caseRoot)=>{
  const name=path.basename(caseRoot);
  const log=Object.hasOwn(groupLog,name)?groupLog[name]:
   '[INFO] Loading begins...\n[INFO] Loading ended.\n';
  fs.writeFileSync(path.join(caseRoot,'user/openxcom.log'),log);
  return {exitCode:124,signal:null,error:null,elapsedMs:7,stdout:'',stderr:''};
 };
}
test('baseline and independent source group candidates preserve NO-GAMEPLAY boundary',t=>{
 const {source,work}=setup(t);
 const r=ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor()});
 assert.equal(r.baselineLoadPass,true);assert.equal(r.nativeMinimumProven,false);
 assert.equal(r.originalCommercialGame,'NOT_RUN');assert.equal(r.gameplayQualified,false);
 assert.equal(r.groups,5);assert.equal(r.cases.length,5);
 assert.ok(r.cases.every(x=>x.bootClass==='ENGINE_LOAD_REACHED'&&x.gameplayQualified===false));
 assert.ok(fs.existsSync(path.join(work,'native-ablation-report.json')));
 assert.ok(fs.existsSync(path.join(source,'data/SYNTHETIC/SOUND/SAMPLE.CAT')));
});
test('resource loss is reported as failed native load rather than minimum proof',t=>{
 const {source,work}=setup(t);
 const r=ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor({
  'without-SOUND':'[INFO] Loading begins...\n[ERROR] GEO.CAT not found\n'})});
 const c=r.cases.find(x=>x.removeGroup==='native-resource:SOUND');
 assert.equal(c.bootClass,'ENGINE_LOAD_INCOMPLETE_WITH_ERROR');
 assert.equal(c.loadCompleted,false);assert.equal(r.nativeMinimumProven,false);
});
test('baseline failure is fail closed even if one candidate boots',t=>{
 const {source,work}=setup(t);
 assert.throws(()=>ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor({
  baseline:'[INFO] Loading begins...\n[FATAL] missing resource\n'})}),/baseline native startup did not reach/);
 assert.ok(fs.existsSync(path.join(work,'native-ablation-report.json')));
});
test('never clobber an existing native output',t=>{
 const {source,work}=setup(t);fs.mkdirSync(work);
 assert.throws(()=>ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor()}),/refusing to overwrite/);
});
test('missing immutable public upstream data closure is rejected',t=>{
 const {source,work}=setup(t);fs.rmSync(path.join(source,'data/common'),{recursive:true});
 assert.throws(()=>ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor()}),/missing stock public data\/common/);
});
test('old success log cannot qualify a new process that does nothing',t=>{
 const {source,work}=setup(t);
 fs.writeFileSync(path.join(source,'user/openxcom.log'),'Loading begins...\nLoading ended.\n');
 assert.throws(()=>ablate({sourceRoot:source,workRoot:work,invoke:()=>({exitCode:124,signal:null,error:null,elapsedMs:3,stdout:'',stderr:''})}),/baseline native startup did not reach/);
});

test('retail original game root fails before any native execution',t=>{
 const {source,work}=setup(t);
 const bad=path.join(source,'data/UFO');fs.mkdirSync(bad);
 assert.throws(()=>ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor()}),/commercial game path not admitted/);
});

test('a nonfatal source error still prevents an unqualified clean-load claim',t=>{
 const {source,work}=setup(t);
 const r=ablate({sourceRoot:source,workRoot:work,invoke:syntheticExecutor({
  'without-SOUND':'[INFO] Loading begins...\n[ERROR] Missing optional audio\n[INFO] Loading ended.\n'})});
 const c=r.cases.find(x=>x.removeGroup==='native-resource:SOUND');
 assert.equal(c.bootClass,'ENGINE_LOAD_REACHED_WITH_ERRORS');
 assert.equal(c.loadCompleted,true);assert.ok(c.firstError);
 assert.equal(c.gameplayQualified,false);assert.equal(r.nativeMinimumProven,false);
});
