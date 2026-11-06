#!/usr/bin/env node
// Public original-only RDTE: exact legacy generator/source test, no native game.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dir = path.dirname(fileURLToPath(import.meta.url));
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'oxce-public-legacy-'));
const hash = bytes => createHash('sha256').update(bytes).digest('hex');
const check = (condition, message) => assert.ok(condition, message);
function verify(root) {
  const read = p => fs.readFileSync(path.join(root, p));
  const manifest = JSON.parse(read('rdte/synthetic-legacy-manifest.json'));
  assert.equal(manifest.originalCommercialAssets, false);
  assert.equal(manifest.engineCompatibilityVerified, false);
  assert.equal(manifest.files.length, 79);
  for (const file of manifest.files) {
    check(!path.isAbsolute(file.path) && !file.path.includes('..'), 'unsafe asset path');
    const data = read(file.path);
    assert.equal(data.length, file.length);
    assert.equal(hash(data), file.sha256);
  }
  const mcd = read('TERRAIN/BLANKS.MCD');
  const map = read('MAPS/SFREL00.MAP');
  const rmp = read('ROUTES/SFREL00.RMP');
  const tab = read('TERRAIN/BLANKS.TAB');
  const pck = read('TERRAIN/BLANKS.PCK');
  assert.equal(mcd.length, 3*62);
  assert.equal(mcd[62+59], 0);
  assert.equal(mcd[2*62+53], 0);
  assert.equal(mcd[2*62+59], 1);
  assert.equal(mcd[2*62+39], 4);
  assert.deepEqual([...map.subarray(0,3)], [10,10,1]);
  for(let i=0; i<100; i++) assert.deepEqual([...map.subarray(3+i*4,7+i*4)], [2,0,0,0]);
  assert.equal(rmp.length, 6*24);
  for(let i=0;i<6;i++) {
    const n=rmp.subarray(i*24,(i+1)*24);
    assert.equal(n[20],i<2?1:0);
    assert.equal(n[21],i>=2?1:0);
    assert.equal(n[23],i<4?1:0);
    if(i===2||i===3) {
      const linked=[0,1,2,3,4].map(j=>n[4+3*j]);
      check(linked.some(id=>id>=4&&id<6&&rmp[id*24+23]===0), 'hostile lacks vacant patrol-only waypoint');
    }
  }
  assert.equal(tab.length,512*4);
  assert.equal(pck.length,512*2);
  for(let i=0;i<512;i++) {
    assert.equal(tab.readUInt32LE(i*4),i*2);
    assert.deepEqual([...pck.subarray(i*2,i*2+2)],[0,255]);
  }
  for(const [name,w,h] of [['touch-controls',320,24],['tiny-ranks',42,7],['kneel-button',64,16]]) {
    const png=read('multiplayer/'+name+'.png');
    assert.equal(png.readUInt32BE(16),w);
    assert.equal(png.readUInt32BE(20),h);
    assert.equal(png[25],3);
  }
  for(const name of ['gmgeo1','gminter','gmtactic1','gmenbase']) {
    const wav=read('SOUND/'+name+'.wav');
    assert.equal(wav.subarray(0,4).toString(),'RIFF');
    assert.equal(wav.readUInt32LE(24),22050);
    assert.equal(wav.readUInt32LE(40),88200);
  }
  // Real OXCE CatFile.cpp reads a {offset,size} LE table; for SAMPLE.CAT
  // SoundSet::loadCatByIndex consumes one-byte name length + name, then RIFF.
  // Exact source bytes are generated here, not copied from UFO/TFTD retail.
  function assertSyntheticCat(catName,label) {
    const cat=read('SOUND/'+catName+'.CAT');
    function decode(b) {
      assert.ok(b.length>60);
      const offset=b.readUInt32LE(0);
      assert.equal(offset,8); // exactly one index entry
      const n=b[offset];
      assert.ok(n>0 && n<64 && offset+1+n+44<=b.length);
      assert.equal(b.toString('ascii',offset+1,offset+1+n),label);
      const pos=offset+1+n;
      assert.equal(b.readUInt32LE(4),b.length-pos); // SAMPLE: wav size only
      assert.equal(b.toString('ascii',pos,pos+4),'RIFF');
      assert.equal(b.toString('ascii',pos+8,pos+12),'WAVE');
      assert.equal(b.readUInt32LE(pos+24),22050);
      assert.equal(b.readUInt32LE(pos+40),88200);
      assert.equal(b.readUInt32LE(pos+4)+8,b.length-pos);
    }
    decode(cat);
    const badOffset=Buffer.from(cat);badOffset.writeUInt32LE(0,0);
    const badSize=Buffer.from(cat);badSize.writeUInt32LE(1,4);
    const badName=Buffer.from(cat);badName[8]=255;
    const badRiff=Buffer.from(cat);badRiff.write('NOPE',9+label.length);
    const badWave=Buffer.from(cat);badWave.write('NOPE',9+label.length+8);
    for(const bad of [badOffset,badSize,badName,badRiff,badWave])
      assert.throws(()=>decode(bad),'must reject corrupted CAT header/name/WAV');
  }
  assertSyntheticCat('SAMPLE','SF-GEO');
  assertSyntheticCat('SAMPLE2','SF-BATTLE');
  return {files:manifest.files.length,manifest:hash(read('rdte/synthetic-legacy-manifest.json')),mcd:hash(mcd),map:hash(map),rmp:hash(rmp)};
}
try {
  const first=path.join(temporary,'a'),second=path.join(temporary,'b');
  for(const target of [first,second]) execFileSync(process.execPath,[path.join(dir,'generate-legacy.mjs'),target]);
  const a=verify(first),b=verify(second);
  assert.deepEqual(a,b,'generator output must be deterministic');
  console.log(JSON.stringify({result:'PASS',scope:'public original-only source',generator:'exact',checks:a,deterministic:true,nativeOXCE:false},null,2));
} finally { fs.rmSync(temporary,{recursive:true,force:true}); }
