#!/usr/bin/env node
// Public-domain, independently written, synthetic-only engine-smoke data.
// No original X-COM files, original X-COM rules, private repo code or licensed assets.
import { mkdirSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';

const out=resolve(process.argv[2]||'.out/engine-smoke');
const mod=join(out,'user/mods/rdte-synthetic-master');
mkdirSync(join(mod,'Ruleset'),{recursive:true});
mkdirSync(join(mod,'Language'),{recursive:true});
mkdirSync(join(out,'config'),{recursive:true});
mkdirSync(join(out,'data'),{recursive:true});
writeFileSync(join(mod,'metadata.yml'), [
  'id: rdte-synthetic-master',
  'name: "Synthetic Master for Public OXCE RDTE"',
  'version: "0.1"',
  'author: "Independent synthetic RDTE fixture"',
  'isMaster: true',
  'master: ""',
  'reservedSpace: 1',
  'requiredExtendedVersion: "8.6"',
  'loadResources:',
  '  - SYNTHETIC',
  ''
].join('\n'));
writeFileSync(join(mod,'Ruleset/10-synthetic.rul'),[
  'fontName: rdte-fonts.yml',
  'interfaces:',
  '  - type: mainMenu',
  '    backgroundImage: BACK01.SCR',
  '    elements:',
  '      - id: palette',
  '        color: 0',
  '      - id: window',
  '        color: 177',
  '      - id: text',
  '        color: 206',
  '      - id: button',
  '        color: 181',
  '      - id: battlescapeTheme',
  '        color: -1',
  '        color2: -1',
  '        border: 0',
  '  - type: newGameMenu',
  '    backgroundImage: BACK02.SCR',
  '    elements:',
  '      - id: window',
  '        color: 177',
  '      - id: text',
  '        color: 206',
  '      - id: button',
  '        color: 181',
  '  - type: newBattleMenu',
  '    backgroundImage: BACK03.SCR',
  '    elements:',
  '      - id: window',
  '        color: 177',
  '      - id: heading',
  '        color: 206',
  '      - id: button1',
  '        color: 181',
  '      - id: button2',
  '        color: 181',
  '# Original Aster Vale geoscape bootstrap: one funded sector avoids zero-country division',
  'countries:',
  '  - type: STR_SF_ASTER_COUNCIL',
  '    fundingBase: 2000',
  '    fundingCap: 10000',
  '    labelLon: 120',
  '    labelLat: -20',
  '    areas:',
  '      - [10, 340, -65, 65]',
  'regions:',
  '  - type: STR_SF_ASTER_SECTOR',
  '    cost: 50000',
  '    regionWeight: 1',
  '    areas:',
  '      - [10, 340, -65, 65]',
  '    missionWeights: {}',
  'startingBase:',
  '  facilities: []',
  '  crafts: []',
  '  items: {}',
  '  randomSoldiers: 0',
  '  scientists: 0',
  '  engineers: 0',
  'research:',
  '  - name: STR_RDTE_RESEARCH',
  '    cost: 1',
  '    points: 1',
  'items:',
  '  - type: STR_RDTE_CELL',
  '    battleType: 2',
  '    clipSize: 4',
  '    power: 5',
  '    damageType: 1',
  '  - type: STR_RDTE_TOOL',
  '    battleType: 1',
  '    compatibleAmmo: [STR_RDTE_CELL]',
  '    accuracySnap: 50',
  '    tuSnap: 40',
  ''
].join('\n'));
writeFileSync(join(mod,'Language/en-US.yml'),[
  'en-US:',
  '  STR_RDTE_RESEARCH: "Synthetic research"',
  '  STR_RDTE_CELL: "Synthetic cell"',
  '  STR_RDTE_TOOL: "Synthetic tool"',
  ''
].join('\n'));
// Explicitly enable ONLY the synthetic master, and do not import "UFO" or "TFTD".
writeFileSync(join(out,'config/options.cfg'),[
  'mods:',
  '  - id: rdte-synthetic-master',
  '    active: true',
  'options:',
  '  mute: true',
  '  playIntro: false',
  '  lazyLoadResources: true',
  ''
].join('\n'));
console.log(JSON.stringify({root:out,master:'rdte-synthetic-master',
  originalGameAssets:false,engineBootVerified:false}));
