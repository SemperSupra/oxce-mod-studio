#!/usr/bin/env node
// Original-only, dependency-free source preflight for the generated synthetic
// top-level master. This does not substitute for a stock OXCE native execution.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import path from 'node:path';

const root = path.resolve(process.argv[2] || '.out/synthetic');
const mod = path.join(root, 'user/mods/rdte-synthetic-master');
const files = {
  metadata: readFileSync(path.join(mod, 'metadata.yml'), 'utf8'),
  rules: readFileSync(path.join(mod, 'Ruleset/10-synthetic.rul'), 'utf8'),
  language: readFileSync(path.join(mod, 'Language/en-US.yml'), 'utf8'),
  names: readFileSync(path.join(mod, 'SoldierName/sf-names.nam'), 'utf8'),
  options: readFileSync(path.join(root, 'config/options.cfg'), 'utf8'),
};

function rootSection(text, heading) {
  const match = text.match(new RegExp(`^${heading}:\\s*$`, 'm'));
  assert.ok(match, `missing top-level ${heading}`);
  const next = /^([A-Za-z][A-Za-z0-9]*):/gm;
  next.lastIndex = match.index + match[0].length;
  const found = next.exec(text);
  return text.slice(match.index + match[0].length, found ? found.index : text.length);
}
function requireOnce(text, expr, label) {
  assert.equal([...text.matchAll(new RegExp(expr, 'gm'))].length, 1, label);
}
function check({metadata, rules, language, names, options}) {
  assert.match(metadata, /^id: rdte-synthetic-master$/m, 'master identity');
  assert.match(metadata, /^isMaster: true$/m, 'original standalone master');
  assert.match(metadata, /^master: ""$/m, 'no commercial parent');
  assert.match(metadata, /^  - SYNTHETIC$/m, 'isolated original resource VFS');
  assert.match(metadata, /^requiredExtendedVersion: "8\.6"$/m, 'pinned stock compatibility');
  assert.match(options, /^  - id: rdte-synthetic-master$/m, 'enabled original master');
  assert.match(options, /^    active: true$/m, 'master enabled');
  for (const heading of ['extraSprites', 'interfaces', 'countries', 'regions', 'startingBase', 'facilities', 'globe', 'musics', 'crafts', 'alienMissions', 'alienDeployments', 'mapScripts', 'terrains', 'alienRaces', 'alienItemLevels', 'armors', 'soldiers', 'units', 'research', 'invs', 'items']) {
    rootSection(rules, heading);
  }
  const keys = [...rules.matchAll(/^([A-Za-z][A-Za-z0-9]*):/gm)].map(m => m[1]);
  assert.equal(new Set(keys).size, keys.length, 'duplicate root YAML section');
  const scripts = rootSection(rules, 'mapScripts');
  assert.match(scripts, /^  - type: DEFAULT$/m, 'native DEFAULT map script');
  assert.match(scripts, /^      - type: addBlock$/m, 'native block placement');
  assert.match(scripts, /^        size: \[1, 1\]$/m, 'single block');
  assert.match(scripts, /^        blocks: \[0\]$/m, 'first terrain map block');
  const terrain = rootSection(rules, 'terrains');
  assert.match(terrain, /^  - name: SF_RELAY$/m, 'deployment terrain');
  assert.match(terrain, /^    mapDataSets: \[BLANKS\]$/m, 'original synthetic dataset');
  const craft = rootSection(rules, 'crafts');
  assert.match(craft, /^  - type: STR_SF_RESPONDER$/m, 'original responder craft');
  const block = /- name: SFREL00\s+width: 10\s+length: 10\s+height: 1/g;
  assert.equal([...terrain.matchAll(block)].length, 1, '10x10x1 terrain block');
  assert.equal([...craft.matchAll(block)].length, 1, '10x10x1 craft block');
  const deployment = rootSection(rules, 'alienDeployments');
  assert.match(deployment, /^    terrains: \[SF_RELAY\]$/m, 'native deployment terrain reference');
  assert.match(deployment, /^  - type: STR_SF_RELAY_DEPLOYMENT$/m, 'native deployment');
  const race = rootSection(rules, 'alienRaces');
  assert.match(race, /members: \[STR_SF_PROBE/, 'hostile unit binding');
  const unit = rootSection(rules, 'units');
  assert.match(unit, /^  - type: STR_SF_PROBE$/m, 'hostile definition');
  assert.match(unit, /^    armor: STR_SF_SURVEY_ARMOR$/m, 'unit armor reference');
  const armor = rootSection(rules, 'armors');
  assert.match(armor, /^  - type: STR_SF_SURVEY_ARMOR$/m, 'armor definition');
  assert.match(armor, /^    spriteInv: SF_SurveyPaperdoll$/m, 'paperdoll surface reference');
  const sprites = rootSection(rules, 'extraSprites');
  assert.match(sprites, /^  - type: SF_SurveyPaperdoll$/m, 'paperdoll sprite declaration');
  const items = rootSection(rules, 'items');
  for (const item of ['STR_RDTE_TOOL', 'STR_RDTE_CELL', 'STR_SF_SURVEY_SHELL']) {
    requireOnce(items, `^  - type: ${item}$`, `${item} definition`);
  }
  const soldiers = rootSection(rules, 'soldiers');
  assert.match(soldiers, /SoldierName\/sf-names\.nam/, 'named human soldier file');
  assert.match(names, /^maleFirst: \[/m, 'synthetic soldier names');
  assert.match(language, /^en-US:$/m, 'English localization');
  assert.match(language, /STR_SF_RESPONDER:/, 'craft localization');
  assert.match(rules, /^fontName: rdte-fonts\.yml$/m, 'original generated font binding');
  return {sections: keys.length, mapBlocks: 2, mapScripts: 1, itemTypes: 3, references: 'PASS'};
}
function negative(label, patch) {
  const modified = {...files, ...patch(files)};
  assert.throws(() => check(modified), assert.AssertionError, `negative ${label}`);
}
const result = check(files);
negative('master parent required', x => ({metadata: x.metadata.replace('master: ""', 'master: xcom1')}));
negative('missing original VFS', x => ({metadata: x.metadata.replace('  - SYNTHETIC', '  - UFO')}));
negative('wrong map height', x => ({rules: x.rules.replace('          height: 1\n', '          height: 4\n')}));
negative('removed mapScript', x => ({rules: x.rules.replace('mapScripts:', 'omittedScripts:')}));
negative('missing terrain reference', x => ({rules: x.rules.replace('terrains: [SF_RELAY]', 'terrains: [ABSENT]')}));
negative('lost hostile', x => ({rules: x.rules.replace('  - type: STR_SF_PROBE', '  - type: REMOVED_UNIT')}));
negative('lost paperdoll', x => ({rules: x.rules.replace('  - type: SF_SurveyPaperdoll', '  - type: ABSENT_SURFACE')}));
negative('lost tool', x => ({rules: x.rules.replace('  - type: STR_RDTE_TOOL', '  - type: ABSENT_TOOL')}));
console.log(JSON.stringify({result: 'PASS', scope: 'synthetic master source-only references, not native OXCE', positive: result, negativeMutations: 8, nativeGameQualified: false}, null, 2));
