#!/usr/bin/env node
// Public RDTE: fabricate new, non-gameplay-equivalent, legacy binary placeholder
// resources from OXCE's documented reader dimensions; no original game data.
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { createHash } from 'node:crypto';
import { deflateSync } from 'node:zlib';

const root=resolve(process.argv[2]||'.out/synthetic/data');
const made=[];
function put(path,buffer){const dst=join(root,path);mkdirSync(dirname(dst),{recursive:true});writeFileSync(dst,buffer);made.push({path,length:buffer.length,sha256:createHash('sha256').update(buffer).digest('hex')});}
function image(path,width,height){const data=Buffer.alloc(width*height);for(let y=0;y<height;y++)for(let x=0;x<width;x++)data[y*width+x]=((x>>3)+(y>>3))%2?3:2;put(path,data);}
function palettes(){
  const p=Buffer.alloc(5*256*3);
  for(let bank=0;bank<5;bank++)for(let i=0;i<256;i++){
    const at=(bank*256+i)*3;
    p[at]=(i+bank*3)%64;
    p[at+1]=(i*3+bank*5)%64;
    p[at+2]=(i*7+bank*11)%64;
    // Original UI palette: index 1 is briefing/button glyph foreground.
    // Contrast with indexed checkerboard backgrounds 2 and 3.
    if(i===1){p[at]=61;p[at+1]=60;p[at+2]=52;}
  }
  put('GEODATA/PALETTES.DAT',p);
  const b=Buffer.alloc(128*3);
  for(let i=0;i<128;i++){b[i*3]=i%64;b[i*3+1]=(i*2)%64;b[i*3+2]=(i*3)%64;}
  put('GEODATA/BACKPALS.DAT',b);
}
// Reader in frozen OXCE 8.6.5 sees 32-bit little-endian TAB offsets
// when first offset is zero. Each PCK frame begins with zero skipped
// rows and ends with 0xff. All frames are intentionally blank.
function blankPck(dir,name,frames=512) {
  const pck=Buffer.alloc(frames*2);
  const tab=Buffer.alloc(frames*4);
  for(let i=0;i<frames;i++){
    pck[2*i]=0;pck[2*i+1]=255;
    tab.writeUInt32LE(2*i,4*i);
  }
  put(dir+'/'+name+'.PCK',pck);
  put(dir+'/'+name+'.TAB',tab);
}
// Original 9-frame, 3x3 retro-futurist marker strip, encoded with Node stdlib.
function markerPNG(){
  // OXCE Surface::loadImage accepts indexed 8-bit PNG, NOT RGBA.
  const w=27,h=3,scan=Buffer.alloc(h*(1+w));
  for(let y=0;y<h;y++){
    for(let x=0;x<w;x++){
      const i=y*(1+w)+1+x;
      scan[i]=(x%3===1||y===1)?1:0;
    }
  }
  function crc32(src){
    let c=0xffffffff;
    for(const b of src){c^=b;for(let j=0;j<8;j++)c=(c>>>1)^((c&1)?0xedb88320:0);}
    return (c^0xffffffff)>>>0;
  }
  function chunk(name,raw){
    const kind=Buffer.from(name),len=Buffer.alloc(4),check=Buffer.alloc(4);
    len.writeUInt32BE(raw.length);check.writeUInt32BE(crc32(Buffer.concat([kind,raw])));
    return Buffer.concat([len,kind,raw,check]);
  }
  const header=Buffer.alloc(13);
  header.writeUInt32BE(w,0);header.writeUInt32BE(h,4);header[8]=8;header[9]=3;
  return Buffer.concat([
    Buffer.from([137,80,78,71,13,10,26,10]),
    chunk('IHDR',header),
    chunk('PLTE',Buffer.from([0,0,0,36,205,244])),
    chunk('tRNS',Buffer.from([0,255])),
    chunk('IDAT',deflateSync(scan)),chunk('IEND',Buffer.alloc(0))
  ]);
}
put('multiplayer/globe-markers.png',markerPNG());
palettes();
image('GEODATA/INTERWIN.DAT',160,600);
put('GEODATA/SCANG.DAT',Buffer.alloc(4*4));
put('GEODATA/LOFTEMPS.DAT',Buffer.alloc(1024*2));
blankPck('GEOGRAPH','BASEBITS',512);
blankPck('GEOGRAPH','INTICON',512);
image('GEOGRAPH/TEXTURE.DAT',32,32);
image('GEOGRAPH/GEOBORD.SCR',320,200);
for(let i=0;i<20;i++)image('GEOGRAPH/BACK'+String(i).padStart(2,'0')+'.SCR',320,200);
image('UFOGRAPH/SPICONS.DAT',32,24);
blankPck('UFOGRAPH','CURSOR',512);
blankPck('UFOGRAPH','SMOKE',512);
blankPck('UFOGRAPH','HIT',512);
blankPck('UFOGRAPH','X1',512);
image('UFOGRAPH/MEDIBITS.DAT',52,58);
image('UFOGRAPH/DETBLOB.DAT',16,16);
image('UFOGRAPH/TAC00.SCR',320,200);
// These specific UI assets use the Surface::loadSpk reader despite .PCK extension.
// A single skip of 32000 words represents a synthetic 320x200 transparent image.
const blankSpk=Buffer.from([255,255,0,125]);
for(const name of ['TAC01.SCR','DETBORD.PCK','DETBORD2.PCK','ICONS.PCK','MEDIBORD.PCK','SCANBORD.PCK','UNIBORD.PCK'])put('UFOGRAPH/'+name,blankSpk);
blankPck('TERRAIN','BLANKS',512);
// Independently authored minimal two-record MCD required by pinned OXCE reader.
put('TERRAIN/BLANKS.MCD',Buffer.alloc(2*62));
// Source-derived synthetic 10x10x1 map using object 1 as its ground tile.
const sfMap=Buffer.alloc(403);
sfMap[0]=10;sfMap[1]=10;sfMap[2]=1;
for(let tile=0;tile<100;tile++)sfMap[3+tile*4]=1;
put('MAPS/SFREL00.MAP',sfMap);
// Source-derived original 24-byte route-node records (no commercial maps).
const sfRoutes=Buffer.alloc(96);
for(let n=0;n<4;n++){
  const b=n*24;
  sfRoutes[b]=2+n;
  sfRoutes[b+1]=2+n;
  for(let k=0;k<5;k++)sfRoutes[b+4+k*3]=255;
  sfRoutes[b+20]=n<2?1:0;
  sfRoutes[b+23]=1;
}
put('ROUTES/SFREL00.RMP',sfRoutes);



for(const n of ['BIGOBS','FLOOROB','HANDOB']) blankPck('UNITS',n,512);
blankPck('UNITS','SIGNALRELAY',512);

// Independent original emergency-response loop (not existing soundtrack).
// Deterministic 2s mono 16-bit PCM RIFF used as a native source-format probe.
function sfScore(hz) {
 const rate=22050,frames=rate*2, pcm=Buffer.alloc(44+frames*2);
 pcm.write('RIFF',0);pcm.writeUInt32LE(pcm.length-8,4);pcm.write('WAVEfmt ',8);
 pcm.writeUInt32LE(16,16);pcm.writeUInt16LE(1,20);pcm.writeUInt16LE(1,22);
 pcm.writeUInt32LE(rate,24);pcm.writeUInt32LE(rate*2,28);
 pcm.writeUInt16LE(2,32);pcm.writeUInt16LE(16,34);
 pcm.write('data',36);pcm.writeUInt32LE(frames*2,40);
 for(let i=0;i<frames;i++){
  const t=i/rate,bar=Math.floor(t*4),pulse=[1,1.25,1.125,1.5][bar%4];
  const envelope=Math.min(1,t*9)*Math.min(1,(2-t)*9);
  const value=Math.round(7500*envelope*(.63*Math.sin(2*Math.PI*hz*pulse*t)+.25*Math.sin(2*Math.PI*hz*.5*t)+.12*Math.sin(2*Math.PI*hz*2*t)));
  pcm.writeInt16LE(Math.max(-32768,Math.min(32767,value)),44+2*i);
 }
 return pcm;
}
for(const [name,hz] of [['gmgeo1',196],['gminter',277],['gmtactic1',147],['gmenbase',220]])
 put('SOUND/'+name+'.wav',sfScore(hz));
const manifest={kind:'artificial-legacy-binary-placeholders',originalCommercialAssets:false,engineCompatibilityVerified:false,files:made};
mkdirSync(join(root,'rdte'),{recursive:true});
writeFileSync(join(root,'rdte/synthetic-legacy-manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log(JSON.stringify({placeholderCount:made.length,root}));
