#!/usr/bin/env node
// Validates generated *public synthetic-only* stock OXCE prerequisite bytes.
// Does not run the game or qualify movement, AI, debrief, rights, or a release.
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';

const root = path.resolve(process.argv[2] ?? '.out/synthetic');
const read = p => fs.readFileSync(path.join(root,p));
const mcd = read('data/SYNTHETIC/TERRAIN/BLANKS.MCD');
const map = read('data/SYNTHETIC/MAPS/SFREL00.MAP');
const rmp = read('data/SYNTHETIC/ROUTES/SFREL00.RMP');
const rules = read('user/mods/rdte-synthetic-master/Ruleset/10-synthetic.rul').toString('utf8');

function check(mc,mp,rr,rm) {
  assert.equal(mc.length,3*62,'three MCD entries required: blank, scorched, start floor');
  assert.equal(mp.length,403,'10x10x1 map byte length');
  assert.deepEqual([...mp.subarray(0,3)],[10,10,1],'MAP dimensions');
  assert.equal(mc[62+59],0,'BLANKS scorched slot must remain unaltered');
  for (let n=0;n<100;n++) {
    const off=3+n*4, id=mp[off];
    assert.equal(id,2,`tile ${n} must use dedicated floor MCD index 2`);
    assert.equal(mp[off+1]|mp[off+2]|mp[off+3],0,`tile ${n} unexpected wall/object`);
  }
  const floor=2*62;
  assert.equal(mc[floor+53],0,'floor Tile_Type must be O_FLOOR');
  assert.equal(mc[floor+59],1,'floor Target_Type must be START_POINT');
  assert.equal(mc[floor+39],4,'floor TU_Walk must be engine-walkable');
  const blocks=[...rr.matchAll(/- name: SFREL00\s+width: 10\s+length: 10\s+height: 1/g)];
  const refs=[...rr.matchAll(/- name: SFREL00\b/g)];
  assert.equal(blocks.length,2,'craft and terrain mapBlocks both require explicit height 1');
  assert.equal(refs.length,2,'unexpected number of SFREL00 mapBlocks');
  assert.equal(rm.length,4*24,'expected four source-only route nodes');
  for(let i=0;i<4;i++){
    assert.equal(rm[i*24+20],i<2?1:0,`node ${i} rank`);
    assert.equal(rm[i*24+23],1,`node ${i} positive spawn weight`);
  }
  return {mcdRecords:3,mapTiles:100,mapBlocks:2,rmpNodes:4};
}
function mustReject(name,fn){assert.throws(fn,assert.AssertionError,`negative ${name} must reject`);}
const positive=check(mcd,map,rules,rmp);
const damaged=Buffer.from(mcd);damaged[2*62+59]=0;
mustReject('no start marker',()=>check(damaged,map,rules,rmp));
const blocked=Buffer.from(mcd);blocked[2*62+39]=255;
mustReject('nonwalkable start floor',()=>check(blocked,map,rules,rmp));
const swapped=Buffer.from(map);swapped[3]=1;
mustReject('map references scorched slot',()=>check(mcd,swapped,rules,rmp));
mustReject('missing block height',()=>check(mcd,map,rules.replace('height: 1','height: 4'),rmp));
const badWeight=Buffer.from(rmp);badWeight[23]=0;
mustReject('zero spawn priority',()=>check(mcd,map,rules,badWeight));
console.log(JSON.stringify({result:'PASS',scope:'generated public synthetic source-only preconditions',positive,negativeMutations:5,nativeGameQualified:false},null,2));
