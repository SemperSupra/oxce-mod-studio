#!/usr/bin/env node
// Original-only *native evidence* classifier. Rejects command-only and UI-only claims.
// This does not itself run OXCE and never upgrades gameplay, retail, or human gates.
import fs from 'node:fs';
import path from 'node:path';
import { createHash } from 'node:crypto';

const MAX_TRACE_BYTES = 2 * 1024 * 1024;
const MAX_LOG_BYTES = 2 * 1024 * 1024;
const MAX_SAVES = 128;
const SHA256 = b => createHash('sha256').update(b).digest('hex');

function boundedText(file, limit = MAX_LOG_BYTES) {
  const st = fs.lstatSync(file);
  if (!st.isFile() || st.isSymbolicLink() || st.size > limit) throw Error(`untrusted or oversized evidence: ${file}`);
  return fs.readFileSync(file, 'utf8');
}
function stateNames(folder) {
  const lines = boundedText(path.join(folder, 'trace.jsonl'), MAX_TRACE_BYTES).split(/\r?\n/).filter(Boolean);
  if (lines.length < 1 || lines.length > 4096) throw Error('missing or excessive native trace events');
  return lines.map((line,i) => {
    let e;
    try { e = JSON.parse(line); } catch { throw Error(`malformed native trace line ${i+1}`); }
    if (!e || e.event !== 'screen-class' || typeof e.class !== 'string' || !e.class) {
      throw Error(`invalid native state event ${i+1}`);
    }
    return e.class;
  });
}
function after(states, token, start) {return states.findIndex((s,i) => i > start && s.includes(token));}
function listManualSaves(root) {
  const user = path.join(root,'user');
  const results = [];
  function walk(dir) {
    for (const name of fs.readdirSync(dir).sort()) {
      const full = path.join(dir,name);
      const stat = fs.lstatSync(full);
      if (stat.isSymbolicLink() || (!stat.isFile() && !stat.isDirectory())) throw Error('untrusted save tree object');
      if (stat.isDirectory()) {walk(full);continue;}
      if (!/\.sav$/i.test(name) || /\.asav$/i.test(name)) continue;
      if (results.length >= MAX_SAVES) throw Error('too many candidate save files');
      if (stat.size < 512 || stat.size > 64*1024*1024) throw Error(`invalid manual save file size: ${name}`);
      const bytes = fs.readFileSync(full);
      if (bytes.length !== stat.size) throw Error('manual save changed during read');
      results.push({relativePath:path.relative(user,full).split(path.sep).join('/'),size:stat.size,sha256:SHA256(bytes)});
    }
  }
  walk(user);
  return results;
}
export function evaluateBoundary(folder, kind) {
  if (!['save-load','abort'].includes(kind)) throw Error('unsupported boundary kind');
  const states = stateNames(folder);
  const commands = boundedText(path.join(folder,'commands.txt'), 64 * 1024);
  const engineLog = boundedText(path.join(folder,'user','openxcom.log'));
  const firstFatal = engineLog.split(/\r?\n/).find(s=>s.includes('[FATAL]')) || '';
  const battle = after(states, 'BattlescapeState', -1);
  const base = {schema:'oxce.native.boundary-evidence.v1', kind,
    evidenceClass:'ACTUAL_MODIFIED_OXCE_BOUNDARY_PROBE_NOT_STOCK_OR_FULL_GAMEPLAY',
    nativeBattlescapeObserved:battle>=0, stateEvents:states.length,
    stateTraceSha256:SHA256(boundedText(path.join(folder,'trace.jsonl'),MAX_TRACE_BYTES)),
    firstFatal, gameplayQualified:false, stockEngineQualified:false,
    originalRetailGameQualified:false, humanApproved:false};
  if (kind === 'save-load') {
    const saveState = after(states,'SaveGameState',battle);
    const loadState = after(states,'LoadGameState',saveState);
    const returnToBattle = after(states,'BattlescapeState',loadState);
    const manualSaves = listManualSaves(folder);
    const commandF5 = commands.split(/\r?\n/).includes('KEY 286');
    const commandF9 = commands.split(/\r?\n/).includes('KEY 290');
    const evidence = {...base, keyF5Sent:commandF5, keyF9Sent:commandF9,
      saveStateObserved:saveState>battle, loadStateObserved:loadState>saveState,
      returnToBattleObserved:returnToBattle>loadState, manualSaves,
      saveReloadQualified:false, savedStateSemanticsQualified:false};
    evidence.saveReloadQualified = !firstFatal && battle>=0 && commandF5 && commandF9 &&
      evidence.saveStateObserved && evidence.loadStateObserved &&
      evidence.returnToBattleObserved && manualSaves.length>0;
    // Even if the same process returns from LoadGameState, this does not prove
    // that loaded unit/TU state equals the saved state; that is a separate gate.
    return evidence;
  }
  const dialog = after(states,'AbortMissionState',battle);
  const debrief = after(states,'DebriefingState',dialog);
  const abortSent=commands.split(/\r?\n/).includes('CLICK 510 336');
  const confirmSent=commands.split(/\r?\n/).includes('CLICK 100 236');
  return {...base, abortClickSent:abortSent, abortConfirmSent:confirmSent,
    abortDialogObserved:dialog>battle, debriefObserved:debrief>dialog,
    abortDebriefQualified: !firstFatal && battle>=0 && abortSent && confirmSent &&
      dialog>battle && debrief>dialog};
}

if (process.argv[1] && path.resolve(process.argv[1]) === new URL(import.meta.url).pathname) {
  if (process.argv.length !== 5) {
    console.error('usage: node boundary-check.mjs save-load|abort <native-run-root> <receipt.json>');
    process.exitCode=2;
  } else {
    try {
      const kind = process.argv[2],folder=process.argv[3],dest=process.argv[4];
      const receipt = evaluateBoundary(folder,kind);
      fs.writeFileSync(dest, JSON.stringify(receipt,null,2)+'\n',{flag:'wx',mode:0o600});
      console.log(JSON.stringify(receipt));
      if (!(kind==='save-load'?receipt.saveReloadQualified:receipt.abortDebriefQualified)) process.exitCode=1;
    } catch(e) {console.error('NATIVE_BOUNDARY_FAIL: '+e.message);process.exitCode=1;}
  }
}
