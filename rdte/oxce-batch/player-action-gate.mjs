#!/usr/bin/env node
// The player action and hostile-action gates are deliberately independent native scenarios.
// This gate never converts an engine assessment's hostile-only failure to gameplay acceptance.
import fs from 'node:fs';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';

const sha256 = bytes => createHash('sha256').update(bytes).digest('hex');
const FAILURE_HOSTILE_ONLY = 'no-native-enemy-position-or-tu-delta';
export function classifyPlayerAction(assessment, commands, engineLog, nativeBattleTrace) {
  const reasons = [];
  if (!assessment || typeof assessment !== 'object' || Array.isArray(assessment)) {
    reasons.push('missing-or-invalid-tactical-assessment');
  }
  const r = assessment && typeof assessment === 'object' ? assessment : {};
  if (r.GameplayQualified !== false) reasons.push('gameplay-acceptance-must-remain-false');
  if (!Number.isInteger(r.NativeSnapshots) || r.NativeSnapshots < 3) reasons.push('insufficient-native-snapshots');
  if (r.MalformedFrames !== 0) reasons.push('malformed-native-snapshot');
  for (const key of ['PlayerObserved', 'EnemyObserved', 'PlayerMoveObserved',
                    'PlayerTUDecreased', 'PlayerMovementSpentTU',
                    'HostileFactionTurnObserved', 'ReturnedToPlayerTurn']) {
    if (r[key] !== true) reasons.push('missing-' + key);
  }
  const hostileGateUnmetHere = r.FirstFailure === FAILURE_HOSTILE_ONLY && r.EnemyActionObserved === false;
  if (!(r.FirstFailure === '' || hostileGateUnmetHere)) {
    reasons.push('unexpected-native-classifier-failure:' + String(r.FirstFailure));
  }
  if (!String(commands).split(/\r?\n/).includes('CLICK 320 204')) reasons.push('native-player-click-not-sent');
  if (typeof engineLog !== 'string' || engineLog.includes('[FATAL]')) reasons.push('engine-fatal-or-missing-log');
  if (typeof nativeBattleTrace !== 'string' || !nativeBattleTrace.includes('"event":"battle-snapshot"')) reasons.push('native-battle-trace-unavailable');
  const qualified = reasons.length === 0;
  return {
    schema:'oxce.public.native-player-action-gate.v1',
    evidenceClass:'ACTUAL_MODIFIED_OXCE_PLAYER_POSITION_AND_TU_NOT_FULL_GAMEPLAY',
    playerActionQualified:qualified,
    hostileActionInThisTrial:r.EnemyActionObserved === true,
    hostileActionMustBeQualifiedSeparately:true,
    nativeSnapshots:r.NativeSnapshots ?? null,
    originalClassifierFirstFailure:r.FirstFailure ?? null,
    firstFailure:reasons[0] ?? '',
    allFailures:reasons,
    gameplayQualified:false,
    unmodifiedStockOXCEQualified:false,
  };
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 7) {
    console.error('usage: node player-action-gate.mjs <native-assessment.json> <commands.txt> <engine.log> <battle.jsonl> <output.json>');
    process.exitCode=2;
  } else {
    try {
      const [assessmentPath, commandPath, logPath, battlePath, outputPath] = process.argv.slice(2);
      const receipt = JSON.parse(fs.readFileSync(assessmentPath, 'utf8'));
      const commands = fs.readFileSync(commandPath,'utf8');
      const log = fs.readFileSync(logPath,'utf8');
      const battle = fs.readFileSync(battlePath,'utf8');
      const report=classifyPlayerAction(receipt,commands,log,battle);
      report.nativeAssessmentSha256=sha256(fs.readFileSync(assessmentPath));
      report.commandSha256=sha256(fs.readFileSync(commandPath));
      report.nativeTraceSha256=sha256(fs.readFileSync(battlePath));
      fs.writeFileSync(outputPath,JSON.stringify(report,null,2)+'\n', {flag:'wx'});
      console.log(JSON.stringify(report));
      if (!report.playerActionQualified) process.exitCode=1;
    } catch(e) { console.error('NATIVE_PLAYER_ACTION_GATE_FAIL: '+e.message); process.exitCode=1; }
  }
}
