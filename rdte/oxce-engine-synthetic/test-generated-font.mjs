#!/usr/bin/env node
// Independently validates public-original indexed font generation; not a stock-engine test.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { inflateSync } from 'node:zlib';
import { createHash } from 'node:crypto';

const root=path.resolve(process.argv[2] ?? '.out/synthetic');
const dir=path.join(root,'user/mods/rdte-synthetic-master/Language');
const expectedNames=['FONT_BIG','FONT_SMALL','FONT_GEO_BIG','FONT_GEO_SMALL','FONT_TINY','BIG','SMALL'];
const sha256=b=>createHash('sha256').update(b).digest('hex');
const alphabet=Array.from({length:95},(_,i)=>String.fromCharCode(i+32)).join('');
const png=fs.readFileSync(path.join(dir,'rdte-fonts.png'));
const yaml=fs.readFileSync(path.join(dir,'rdte-fonts.yml'),'utf8');
function crc32(data){let c=0xffffffff;for(const b of data){c^=b;for(let i=0;i<8;i++)c=(c>>>1)^((c&1)?0xedb88320:0);}return (c^0xffffffff)>>>0;}
function parsePng(b){
 assert.equal(b.subarray(0,8).toString('hex'),'89504e470d0a1a0a','PNG signature');
 let off=8,ihdr,plte,trns,idat=[],iend=false;
 while(off<b.length){
  assert.ok(off+12<=b.length,'truncated chunk header');
  const len=b.readUInt32BE(off),type=b.toString('ascii',off+4,off+8);
  assert.ok(off+12+len<=b.length,`truncated ${type}`);
  const chunk=b.subarray(off+8,off+8+len),read=b.readUInt32BE(off+8+len);
  assert.equal(read,crc32(b.subarray(off+4,off+8+len)),`${type} CRC`);
  if(type==='IHDR')ihdr=chunk;
  if(type==='PLTE')plte=chunk;
  if(type==='tRNS')trns=chunk;
  if(type==='IDAT')idat.push(chunk);
  if(type==='IEND'){iend=true;assert.equal(off+12+len,b.length,'trailing bytes after IEND');}
  off+=12+len;
 }
 assert.ok(iend,'IEND is required');
 assert.ok(ihdr && plte && trns && idat.length,'missing required indexed PNG chunks');
 assert.deepEqual([ihdr.readUInt32BE(0),ihdr.readUInt32BE(4)],[128,72],'16 x 6 font atlas');
 assert.deepEqual([ihdr[8],ihdr[9],ihdr[10],ihdr[11],ihdr[12]],[8,3,0,0,0],'8bpp indexed noninterlaced');
 assert.equal(plte.length,9,'3 palette entries');
 assert.equal(trns.length,3,'three alpha entries');
 assert.equal(trns[0],0,'transparent background palette 0');
 assert.ok(trns[1]>0,'visible foreground palette 1');
 const scan=inflateSync(Buffer.concat(idat));
 assert.equal(scan.length,72*129,'unfiltered scanlines');
 let ink=0;
 for(let y=0;y<72;y++){
  assert.equal(scan[y*129],0,`scanline ${y} must be unfiltered`);
  for(let x=0;x<128;x++){
   const palette=scan[y*129+1+x];assert.ok(palette<3,`palette index at ${x},${y}`);
   if(palette===1)ink++;
  }
 }
 assert.ok(ink>500,`glyph atlas missing visible pixels: ${ink}`);
 const cellHasInk=(index)=>{
  const x0=(index%16)*8,y0=Math.floor(index/16)*12;
  for(let y=y0;y<y0+12;y++)for(let x=x0;x<x0+8;x++)if(scan[y*129+1+x]===1)return true;
  return false;
 };
 let populated=0;
 for(let i=0;i<95;i++){
  const hasInk=cellHasInk(i);
  if(alphabet[i]===' ')assert.ok(!hasInk,'space must not render ink');
  else {assert.ok(hasInk,`visible glyph ${JSON.stringify(alphabet[i])} blank`);populated++;}
 }
 return {width:128,height:72,visiblePixels:ink,populatedGlyphs:populated};
}
function checkYaml(text){
 assert.equal((text.match(/^  - id: /gm)||[]).length,7,'seven font definitions');
 const ids=[...text.matchAll(/^  - id: (.+)$/gm)].map(m=>m[1]);
 assert.deepEqual(ids,expectedNames,'font identifiers must match stock aliases');
 assert.equal((text.match(/    width: 8/g)||[]).length,7,'cell width');
 assert.equal((text.match(/    height: 12/g)||[]).length,7,'cell height');
 assert.equal((text.match(/      - file: rdte-fonts.png/g)||[]).length,7,'atlas references');
 const charLines=[...text.matchAll(/^        chars: (.+)$/gm)].map(m=>JSON.parse(m[1]));
 assert.equal(charLines.length,7,'character maps');
 for(const chars of charLines)assert.equal(chars,alphabet,'printable ASCII 32..126 order');
 return {definitions:ids.length,characterCount:alphabet.length};
}
function mustFail(label,fn){assert.throws(fn,assert.AssertionError,`negative ${label}`);}
const image=parsePng(png),font=checkYaml(yaml);
const corrupt=Buffer.from(png);corrupt[corrupt.length-5]^=1;
mustFail('CRC corruption',()=>parsePng(corrupt));
const wrongDims=Buffer.from(png);wrongDims.writeUInt32BE(127,16);
mustFail('wrong atlas width',()=>parsePng(wrongDims));
mustFail('missing font definition',()=>checkYaml(yaml.replace('  - id: FONT_BIG\n','')));
mustFail('wrong glyph mapping',()=>checkYaml(yaml.replace(/chars: .+/, 'chars: "ABC"')));
const missingGlyph=Buffer.from(png);
// Clear one non-space glyph ink pixel, then deliberately detect integrity failure.
for(let i=0;i<missingGlyph.length;i++)if(missingGlyph[i]===255){missingGlyph[i]^=0x1;break;}
mustFail('corrupt output',()=>parsePng(missingGlyph));
console.log(JSON.stringify({result:'PASS',scope:'exact-source generated original font artifacts, independent PNG decoder',pngSha256:sha256(png),yamlSha256:sha256(Buffer.from(yaml)),image,font,negativeMutations:5,nativeFontRendered:false},null,2));
