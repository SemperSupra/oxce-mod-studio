#!/usr/bin/env node
// Native public-RDTE evidence classification, never falsely turns a
// completed GitHub workflow into successful OXCE engine validation.
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
const root=resolve(process.argv[2]||'.out/rdte-evidence');
const read=(name)=>existsSync(join(root,name))?readFileSync(join(root,name),'utf8'):'';
const nativeLog=read('openxcom.log');
const phase=read('exit.env');
const firstFatal=nativeLog.split('\n').find(l=>/\[(FATAL)\]/.test(l))||null;
const firstError=nativeLog.split('\n').find(l=>/\[(ERROR)\]/.test(l))||null;
const result={
  schemaVersion:1,
  evidenceClass:'PUBLIC_GHA_ACTUAL_OXCE_NATIVE_ENGINE',
  engineSource:'MeridianOXC/OpenXcom@668880bb756852f0734cfe92e438e400080a2f4d',
  originalGameAssetsIncluded:false,
  cmakeConfigured:/configure_exit=0/.test(phase),
  actualEngineBuilt:/build_exit=0/.test(phase),
  nativeLaunchAttempted:/native_launch_attempted=true/.test(phase),
  syntheticMasterActivated:/Active mods:[\s\S]*rdte-synthetic-master/.test(nativeLog),
  sourceRulesetLoadReached:nativeLog.includes('Loading rulesets...'),
  rulesetLoadCompleted:nativeLog.includes('Loading rulesets done.'),
  engineLoadCompleted:nativeLog.includes('Loading ended.'),
  fatalResourceOrEngineLog:firstFatal,
  firstError:firstError,
  exitStatus:(phase.match(/engine_launch_exit=(\d+)/)||[])[1]||null,
  gameplayQualified:false,
  engineAcceptance:false,
  caveat:'Engine acceptance requires post-load semantic checks; watchdog timeout and workflow success are not proof.'
};
mkdirSync(root,{recursive:true});
writeFileSync(join(root,'classification.json'),JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify(result,null,2));
