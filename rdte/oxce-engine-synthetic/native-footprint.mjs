#!/usr/bin/env node
// Actual OXCE native startup ablation harness for PUBLIC FIRST-PARTY synthetic resources.
// Removing one resource group at a time tests main-menu loading only, never gameplay.
import fs from 'node:fs';

import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { inspectSource } from './source-footprint.mjs';

const sha256 = b => createHash('sha256').update(b).digest('hex');
const groupsOk = x => /^[a-zA-Z0-9_-]+$/.test(x);
const req = (condition, description) => { if (!condition) throw new Error(description); };

function copySyntheticProjection(fullRoot, projectionRoot) {
  // The real engine also needs public upstream data/common and data/standard;
  // the original-only census intentionally excludes those unchanged engine inputs.
  for (const rel of ['config','user/mods/rdte-synthetic-master','data/SYNTHETIC']) {
    const from=path.join(fullRoot,rel),to=path.join(projectionRoot,rel);
    req(fs.existsSync(from),`missing original-only generated source: ${rel}`);
    fs.mkdirSync(path.dirname(to),{recursive:true});
    fs.cpSync(from,to,{recursive:true,errorOnExist:true,force:false,dereference:false});
  }
  return inspectSource(projectionRoot);
}

function nativeExec(engine,caseRoot,seconds) {
  const root=path.resolve(caseRoot), bin=path.resolve(engine);
  const t0=Date.now();
  const proc=spawnSync('timeout',['--kill-after=2s',`${seconds}s`,'xvfb-run','-a',bin,
      '-data',path.join(root,'data'),'-user',path.join(root,'user'),
      '-config',path.join(root,'config')],{
    cwd:root,encoding:'utf8',maxBuffer:1024*1024,timeout:(seconds+4)*1000,
    env:{...process.env,SDL_AUDIODRIVER:'dummy',HOME:path.join(root,'home'),
         TMPDIR:path.join(root,'temp'),XDG_CONFIG_HOME:path.join(root,'config')}
  });
  return {exitCode:proc.status,signal:proc.signal||null,error:proc.error?.message||null,
    elapsedMs:Date.now()-t0,stdout:proc.stdout||'',stderr:proc.stderr||''};
}

function classify(log,runner) {
  const firstFatal=log.split(/\r?\n/).find(s=>s.includes('[FATAL]'))||null;
  const firstError=log.split(/\r?\n/).find(s=>s.includes('[ERROR]'))||null;
  const loadCompleted=log.includes('Loading ended.');
  const startupObserved=log.includes('Loading begins...');
  let bootClass='INDETERMINATE';
  if (startupObserved && loadCompleted && !firstFatal) bootClass='ENGINE_LOAD_REACHED';
  else if (startupObserved && firstFatal) bootClass='ENGINE_LOGGED_FATAL';
  else if (startupObserved && !loadCompleted && firstError) bootClass='ENGINE_LOAD_INCOMPLETE_WITH_ERROR';
  return {startupObserved,loadCompleted,firstError,firstFatal,bootClass,
    boundedExit:runner.exitCode,signal:runner.signal,spawnError:runner.error,elapsedMs:runner.elapsedMs};
}

export function ablate({sourceRoot,workRoot,enginePath=null,seconds=7,invoke=nativeExec}={}) {
  req(sourceRoot&&workRoot,'sourceRoot and workRoot are required');
  req(Number.isInteger(seconds)&&seconds>=2&&seconds<=20,'startup timeout must be 2..20 seconds');
  const source=path.resolve(sourceRoot),work=path.resolve(workRoot);
  req(source!==work&&!source.startsWith(work+path.sep)&&!work.startsWith(source+path.sep),
    'work and source roots may not contain one another');
  req(!fs.existsSync(work),'refusing to overwrite an existing work root');
  req(fs.lstatSync(source).isDirectory()&&!fs.lstatSync(source).isSymbolicLink(),'source root must be a real directory');
  if (invoke===nativeExec) {
    req(enginePath&&fs.statSync(enginePath).isFile(),'actual executable path is required');
    fs.accessSync(enginePath,fs.constants.X_OK);
  }
  for (const rel of ['data/common','data/standard']) {
    const loc=path.join(source,rel);
    req(fs.existsSync(loc),`missing stock public ${rel}`);
    const st=fs.lstatSync(loc);
    req(st.isDirectory()&&!st.isSymbolicLink(),`upstream public root must be a real directory: ${rel}`);
  }
  for (const rel of ['data/UFO','data/TFTD']) req(!fs.existsSync(path.join(source,rel)),`commercial game path not admitted: ${rel}`);
  fs.mkdirSync(work,{recursive:false});
  const projection=path.join(work,'projection');fs.mkdirSync(projection);
  const inventory=copySyntheticProjection(source,projection);
  const candidates=inventory.ablationCandidates;
  req(candidates.length>0&&candidates.length<=20,'out-of-bounds original-only resource group closure');
  const results=[];
  for (const selected of [{removeGroup:'baseline',removedFiles:0,removedBytes:0},...candidates]) {
    const group=selected.removeGroup;
    req(group==='baseline'||group.startsWith('native-resource:'),'unexpected source group');
    const segment=group==='baseline'?null:group.slice('native-resource:'.length);
    if (segment!==null) req(groupsOk(segment),'invalid resource segment');
    const caseId=group==='baseline'?'baseline':`without-${segment}`;
    const caseRoot=path.join(work,caseId);
    fs.cpSync(source,caseRoot,{recursive:true,errorOnExist:true,force:false,dereference:false});
    if (segment!==null) {
      const target=path.join(caseRoot,'data','SYNTHETIC',segment);
      req(fs.existsSync(target),'source resource group missing from actual native tree');
      fs.rmSync(target,{recursive:true,force:false});
    }
    for(const rel of ['home','temp'])fs.mkdirSync(path.join(caseRoot,rel));
    const stale=path.join(caseRoot,'user/openxcom.log');
    fs.rmSync(stale,{force:true});
    req(!fs.existsSync(stale),'stale engine log remained in native capsule');
    const probe=invoke(enginePath,caseRoot,seconds);
    const log=fs.existsSync(stale)?fs.readFileSync(stale,'utf8'):'';
    const verdict=classify(log,probe);
    const receipt={caseId,removeGroup:group,removedFiles:selected.removedFiles,
      removedBytes:selected.removedBytes,...verdict,gameplayQualified:false,
      firstPlayerAction:'NOT_RUN',stockOriginalGame:'NOT_RUN'};
    // Preserve bounded stdio, hash it for independent custody.
    for(const channel of ['stdout','stderr']) {
      const raw=probe[channel]||'';
      fs.writeFileSync(path.join(caseRoot,`native-${channel}.log`),raw.slice(0,1024*1024));
      receipt[`${channel}Sha256`]=sha256(raw);
    }
    fs.writeFileSync(path.join(caseRoot,'native-probe.json'),JSON.stringify(receipt,null,2)+'\n');
    results.push(receipt);
    process.stderr.write(JSON.stringify({schema:'oxce.native-ablation-progress.v1',
      timestamp:new Date().toISOString(),caseId,group,status:verdict.bootClass,
      finished:results.length,total:candidates.length+1})+'\n');
  }
  const base=results[0];
  const output={schema:'oxce.synthetic.native-footprint.v1',
    evidenceClass:'PUBLIC_FIRST_PARTY_NATIVE_STARTUP_RESOURCE_ABLATION_NOT_GAMEPLAY',
    sourceManifestSha256:inventory.baseline.manifestSha256,
    groups:candidates.length,baseline:base,cases:results.slice(1),
    baselineLoadPass:base.bootClass==='ENGINE_LOAD_REACHED',
    nativeMinimumProven:false,originalCommercialGame:'NOT_RUN',
    selectedSignalFrontierGame:'NOT_RUN',gameplayQualified:false};
  fs.writeFileSync(path.join(work,'native-ablation-report.json'),JSON.stringify(output,null,2)+'\n');
  if (!output.baselineLoadPass) throw Error('baseline native startup did not reach engine load; no ablation interpretation');
  return output;
}

if (process.argv[1] && path.resolve(process.argv[1])===fileURLToPath(import.meta.url)) {
  if (process.argv.length!==5){console.error('usage: node native-footprint.mjs <full-synthetic-root> <verified-oxce-elf> <new-evidence-root>');process.exitCode=2;}
  else{try{const r=ablate({sourceRoot:process.argv[2],enginePath:process.argv[3],workRoot:process.argv[4]});
    console.log(JSON.stringify({schema:r.schema,baselineLoadPass:r.baselineLoadPass,groups:r.groups,
      cases:r.cases.map(c=>({group:c.removeGroup,bootClass:c.bootClass,firstError:c.firstError,firstFatal:c.firstFatal})),
      gameplayQualified:false},null,2));}
    catch(e){console.error('NATIVE_ABLATION_NO_ACCEPTANCE: '+e.message);process.exitCode=1;}}
}
