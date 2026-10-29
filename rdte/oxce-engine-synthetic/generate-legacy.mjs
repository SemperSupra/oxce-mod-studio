#!/usr/bin/env node
// Public RDTE: fabricate new, non-gameplay-equivalent, legacy binary placeholder
// resources from OXCE's documented reader dimensions; no original game data.
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { createHash } from 'node:crypto';

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
palettes();
image('GEODATA/INTERWIN.DAT',160,600);
put('GEODATA/SCANG.DAT',Buffer.alloc(4*4));
put('GEODATA/LOFTEMPS.DAT',Buffer.alloc(1024*2));
blankPck('GEOGRAPH','BASEBITS',512);
blankPck('GEOGRAPH','INTICON',512);
image('GEOGRAPH/TEXTURE.DAT',32,32);
image('GEOGRAPH/GEOBORD.SCR',320,200);
image('UFOGRAPH/SPICONS.DAT',32,24);
blankPck('UFOGRAPH','CURSOR',512);
blankPck('UFOGRAPH','SMOKE',512);
blankPck('UFOGRAPH','HIT',512);
blankPck('UFOGRAPH','X1',512);
image('UFOGRAPH/MEDIBITS.DAT',52,58);
image('UFOGRAPH/DETBLOB.DAT',16,16);
image('UFOGRAPH/TAC00.SCR',320,200);
blankPck('TERRAIN','BLANKS',512);
for(const n of ['BIGOBS','FLOOROB','HANDOB']) blankPck('UNITS',n,512);
const manifest={kind:'artificial-legacy-binary-placeholders',originalCommercialAssets:false,engineCompatibilityVerified:false,files:made};
mkdirSync(join(root,'rdte'),{recursive:true});
writeFileSync(join(root,'rdte/synthetic-legacy-manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log(JSON.stringify({placeholderCount:made.length,root}));
