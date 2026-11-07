#!/usr/bin/env node
// Original (fictional) 5x7 terminal glyph art; for public engine RDTE only.
// No source from the commercial games, fonts, mods or private Studio code.
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve, join } from 'node:path';
import { deflateSync } from 'node:zlib';
import { createHash } from 'node:crypto';
const root=resolve(process.argv[2]||'.out/synthetic');
const dir=join(root,'user/mods/rdte-synthetic-master/Language');
mkdirSync(dir,{recursive:true});
// Each row encodes 5 columns, high bit first. Minimal original legible atlas.
const glyphs={
 A:[14,17,17,31,17,17,17],B:[30,17,17,30,17,17,30],C:[15,16,16,16,16,16,15],
 D:[30,17,17,17,17,17,30],E:[31,16,16,30,16,16,31],F:[31,16,16,30,16,16,16],
 G:[15,16,16,19,17,17,15],H:[17,17,17,31,17,17,17],I:[31,4,4,4,4,4,31],
 J:[7,2,2,2,18,18,12],K:[17,18,20,24,20,18,17],L:[16,16,16,16,16,16,31],
 M:[17,27,21,21,17,17,17],N:[17,25,21,19,17,17,17],O:[14,17,17,17,17,17,14],
 P:[30,17,17,30,16,16,16],Q:[14,17,17,17,21,18,13],R:[30,17,17,30,20,18,17],
 S:[15,16,16,14,1,1,30],T:[31,4,4,4,4,4,4],U:[17,17,17,17,17,17,14],
 V:[17,17,17,17,17,10,4],W:[17,17,17,21,21,21,10],X:[17,17,10,4,10,17,17],
 Y:[17,17,10,4,4,4,4],Z:[31,1,2,4,8,16,31],
 '0':[14,17,19,21,25,17,14],'1':[4,12,4,4,4,4,14],'2':[14,17,1,2,4,8,31],
 '3':[30,1,1,14,1,1,30],'4':[2,6,10,18,31,2,2],'5':[31,16,30,1,1,17,14],
 '6':[6,8,16,30,17,17,14],'7':[31,1,2,4,8,8,8],'8':[14,17,17,14,17,17,14],
 '9':[14,17,17,15,1,2,12],
 '?':[14,17,1,2,4,0,4],'.':[0,0,0,0,0,4,4],',':[0,0,0,0,4,4,8],
 ':':[0,4,4,0,4,4,0],'-':[0,0,0,31,0,0,0],'+':[0,4,4,31,4,4,0],
 '/':[1,1,2,4,8,16,16],' ':[0,0,0,0,0,0,0],
 '!':[4,4,4,4,4,0,4],'_':[0,0,0,0,0,0,31],
 '(':[2,4,8,8,8,4,2],')':[8,4,2,2,2,4,8],
 '[':[14,8,8,8,8,8,14],']':[14,2,2,2,2,2,14],
 '=':[0,0,31,0,31,0,0],
};
const chars=Array.from({length:95},(_,i)=>String.fromCharCode(i+32)).join('');
const W=16*8,H=6*12;const pixels=Buffer.alloc(H*(W+1));
for(let index=0;index<chars.length;index++){
 const c=chars[index],map=glyphs[c]||glyphs[c.toUpperCase()]||glyphs['?'];
 const cx=(index%16)*8,cy=Math.floor(index/16)*12;
 for(let y=0;y<7;y++)for(let x=0;x<5;x++){
  if((map[y]>>(4-x))&1)pixels[(cy+y+2)*(W+1)+cx+x+2+1]=1; // uniform visible foreground; alternate index 2 is near-black in native UI palettes
 }
}
const ct=Array.from({length:256},(_,i)=>{let c=i;for(let j=0;j<8;j++)c=(c&1)?(0xedb88320^(c>>>1)):(c>>>1);return c>>>0;});
function crc(bytes){let c=0xffffffff;for(let i=0;i<bytes.length;i++)c=ct[(c^bytes[i])&255]^(c>>>8);return (c^0xffffffff)>>>0;}
function chunk(name,body){const kind=Buffer.from(name,'ascii'),length=Buffer.alloc(4),checksum=Buffer.alloc(4);length.writeUInt32BE(body.length);checksum.writeUInt32BE(crc(Buffer.concat([kind,body])));return Buffer.concat([length,kind,body,checksum]);}
const ihdr=Buffer.alloc(13);ihdr.writeUInt32BE(W,0);ihdr.writeUInt32BE(H,4);ihdr[8]=8;ihdr[9]=3;
const png=Buffer.concat([Buffer.from('89504e470d0a1a0a','hex'),chunk('IHDR',ihdr),chunk('PLTE',Buffer.from([0,0,0,222,225,215,91,226,244])),chunk('tRNS',Buffer.from([0,255,255])),chunk('IDAT',deflateSync(pixels,{level:9})),chunk('IEND',Buffer.alloc(0))]);
const names=['FONT_BIG','FONT_SMALL','FONT_GEO_BIG','FONT_GEO_SMALL','FONT_TINY','BIG','SMALL'];
const yaml=['fonts:'];
for(const id of names)yaml.push('  - id: '+id,'    width: 8','    height: 12','    spacing: 1','    monospace: true','    images:','      - file: rdte-fonts.png','        chars: '+JSON.stringify(chars));
writeFileSync(join(dir,'rdte-fonts.yml'),yaml.join('\n')+'\n');
writeFileSync(join(dir,'rdte-fonts.png'),png);
console.log(JSON.stringify({fonts:names.length,characters:chars.length,atlas:[W,H],sha256:createHash('sha256').update(png).digest('hex'),nativeEngineVerified:false}));
