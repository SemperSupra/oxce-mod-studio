#!/usr/bin/env node
// No original game bytes, native engine or simulated gameplay acceptance.
import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { inspectSource } from './source-footprint.mjs';

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
 'data/SYNTHETIC/TERRAIN/BLANKS.MCD',
 ];
const put=(root,p,contents='first-party fixture')=>{
 const dest=path.join(root,p);fs.mkdirSync(path.dirname(dest),{recursive:true});fs.writeFileSync(dest,contents);
};
function fixture(t){
 const root=fs.mkdtempSync(path.join(os.tmpdir(),'oxce-source-footprint-'));
 t.after(()=>fs.rmSync(root,{recursive:true,force:true}));
 for(const r of required)put(root,r);
 put(root,'data/SYNTHETIC/GEOGRAPH/BACK00.SCR','back');
 put(root,'data/SYNTHETIC/UFOGRAPH/TAC00.SCR','tac');
 put(root,'data/SYNTHETIC/rdte/synthetic-legacy-manifest.json','{}');
 return root;
}

test('deterministic, bounded source census and independent group ablation plan',t=>{
 const root=fixture(t), a=inspectSource(root), b=inspectSource(root);
 assert.deepEqual(a,b);
 assert.equal(a.baseline.files,15);
 assert.equal(a.ablationCandidates.length,7);
 assert.equal(a.sourceOnly,true);assert.equal(a.gameplayQualified,false);
 assert.equal(a.selectedSignalFrontierGame,'NOT_RUN');
 for(const c of a.ablationCandidates){
   assert.equal(c.nativeResult,'NOT_RUN');assert.equal(c.remainingFiles+c.removedFiles,15);
   assert.ok(/^[a-f0-9]{64}$/.test(c.remainingManifestSha256));
 }
});
test('same-size byte corruption changes baseline and relevant candidate digest',t=>{
 const root=fixture(t), before=inspectSource(root);
 put(root,'data/SYNTHETIC/GEOGRAPH/BACK00.SCR','bork');
 const after=inspectSource(root);
 assert.notEqual(after.baseline.manifestSha256,before.baseline.manifestSha256);
 assert.equal(after.baseline.bytes,before.baseline.bytes);
 const ca=before.ablationCandidates.find(c=>c.removeGroup==='native-resource:GEOGRAPH');
 const cb=after.ablationCandidates.find(c=>c.removeGroup==='native-resource:GEOGRAPH');
 assert.equal(ca.remainingManifestSha256,cb.remainingManifestSha256);
 const otherBefore=before.ablationCandidates.find(c=>c.removeGroup==='native-resource:UFOGRAPH');
 const otherAfter=after.ablationCandidates.find(c=>c.removeGroup==='native-resource:UFOGRAPH');
 assert.notEqual(otherBefore.remainingManifestSha256,otherAfter.remainingManifestSha256);
});
test('missing required source is refused, not classified as a smaller valid master',t=>{
 const root=fixture(t);fs.unlinkSync(path.join(root,'data/SYNTHETIC/SOUND/SAMPLE2.CAT'));
 assert.throws(()=>inspectSource(root),/missing required public synthetic source: data\/SYNTHETIC\/SOUND\/SAMPLE2.CAT/);
});
test('unadmitted retail folder is refused',t=>{
 const root=fixture(t);put(root,'data/UFO/GEODATA/PALETTES.DAT','retail-not-allowed');
 assert.throws(()=>inspectSource(root),/unadmitted synthetic member/);
});
test('symlink cannot enter the census',t=>{
 const root=fixture(t);
 try{fs.symlinkSync(path.join(root,'data/SYNTHETIC/SOUND/SAMPLE.CAT'),path.join(root,'data/SYNTHETIC/SOUND/OTHER.CAT'));}
 catch(e){if(['EPERM','EACCES'].includes(e.code)){t.skip('host blocks symlink');return;}throw e;}
 assert.throws(()=>inspectSource(root),/symlink rejected/);
});
test('casefold collision is rejected for cross-platform resource compatibility',t=>{
 const root=fixture(t);put(root,'data/SYNTHETIC/SOUND/sample.cat','case-collision');
 assert.throws(()=>inspectSource(root),/casefold collision/);
});
test('untrusted directory aliases and control characters are refused',t=>{
 const root=fixture(t);put(root,'data/SYNTHETIC/SOUND/aux.','invalid');
 assert.throws(()=>inspectSource(root),/unsafe synthetic member path/);
});
test('casefold collisions between directories are rejected',t=>{
 const root=fixture(t);put(root,'data/SYNTHETIC/sound/OTHER.PCK','alias');
 assert.throws(()=>inspectSource(root),/casefold collision/);
});
test('reserved device aliases are rejected',t=>{
 const root=fixture(t);put(root,'data/SYNTHETIC/SOUND/CON.MCD','device');
 assert.throws(()=>inspectSource(root),/unsafe synthetic member path/);
});
test('no native claim appears even if all source groups exist',t=>{
 const a=inspectSource(fixture(t));
 assert.ok(a.groups.length>0);
 assert.equal(a.unmodifiedStockOXCE,'NOT_RUN');
 assert.equal(a.gameplayQualified,false);
 assert.ok(a.ablationCandidates.every(c=>c.nativeResult==='NOT_RUN'));
});
