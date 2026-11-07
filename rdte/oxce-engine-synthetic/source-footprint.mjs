#!/usr/bin/env node
// Source-only, independent inventory/ablation plan for the public synthetic OXCE master.
// Does not mutate assets, impersonate OXCE's FileMap, run a binary, or qualify gameplay.
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const MAX_FILES = 10000;
const MAX_BYTES = 256 * 1024 * 1024;
const REQUIRED = Object.freeze([
  'config/options.cfg',
  'user/mods/rdte-synthetic-master/metadata.yml',
  'user/mods/rdte-synthetic-master/Ruleset/10-synthetic.rul',
  'user/mods/rdte-synthetic-master/Language/en-US.yml',
  'user/mods/rdte-synthetic-master/Language/rdte-fonts.yml',
  'user/mods/rdte-synthetic-master/Language/rdte-fonts.png',
  'data/SYNTHETIC/GEODATA/PALETTES.DAT',
  'data/SYNTHETIC/SOUND/SAMPLE.CAT',
  'data/SYNTHETIC/SOUND/SAMPLE2.CAT',
  'data/SYNTHETIC/MAPS/SFREL00.MAP',
  'data/SYNTHETIC/ROUTES/SFREL00.RMP',
  'data/SYNTHETIC/TERRAIN/BLANKS.MCD',
]);
const hash = value => createHash('sha256').update(value).digest('hex');
const stable = entries => hash(JSON.stringify(entries.map(({path: p, size, sha256}) => ({path:p,size,sha256}))));

function validateRel(rel) {
  if (rel.startsWith('/') || /[\\:<>"|?*\u0000-\u001f\u007f]/u.test(rel) ||
      rel.split('/').some(v => !v || v === '.' || v === '..' || v.endsWith(' ') || v.endsWith('.') ||
        /^(?:con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\.|$)/i.test(v))) {
    throw Error(`unsafe synthetic member path: ${rel}`);
  }
  if (!(new Set(['config','user','user/mods','user/mods/rdte-synthetic-master', 'data','data/SYNTHETIC']).has(rel) ||
        rel === 'config/options.cfg' ||
        rel.startsWith('user/mods/rdte-synthetic-master/') ||
        rel.startsWith('data/SYNTHETIC/'))) {
    throw Error(`unadmitted synthetic member: ${rel}`);
  }
}

function groupFor(rel) {
  if (rel === 'config/options.cfg') return 'config';
  if (rel.startsWith('user/mods/')) return 'master-source';
  const segment = rel.slice('data/SYNTHETIC/'.length).split('/')[0];
  return segment === 'rdte' ? 'source-provenance' : `native-resource:${segment}`;
}

export function inspectSource(root) {
  const rootStat = fs.lstatSync(root);
  if (!rootStat.isDirectory() || rootStat.isSymbolicLink()) throw Error('source root must be a real directory');
  const entries = [];
  const spellings = new Set();
  let bytes = 0;
  function walk(directory) {
    const names = fs.readdirSync(directory).sort();
    for (const name of names) {
      const abs = path.join(directory, name);
      const rel = path.relative(root, abs).split(path.sep).join('/');
      validateRel(rel);
      const folded = rel.toLowerCase();
      if (spellings.has(folded)) throw Error(`casefold collision: ${rel}`);
      spellings.add(folded);
      const stat = fs.lstatSync(abs);
      if (stat.isSymbolicLink()) throw Error(`symlink rejected: ${rel}`);
      if (stat.isDirectory()) { walk(abs); continue; }
      if (!stat.isFile()) throw Error(`non-regular synthetic member: ${rel}`);
      bytes += stat.size;
      if (entries.length >= MAX_FILES || bytes > MAX_BYTES) throw Error('synthetic corpus exceeds bounded source scan');
      const data = fs.readFileSync(abs);
      if (data.length !== stat.size) throw Error(`source changed while reading: ${rel}`);
      entries.push({path: rel, size: data.length, sha256: hash(data), group: groupFor(rel)});
    }
  }
  walk(root);
  entries.sort((a,b) => a.path < b.path ? -1 : a.path > b.path ? 1 : 0);
  const present = new Set(entries.map(e => e.path));
  for (const name of REQUIRED) if (!present.has(name)) throw Error(`missing required public synthetic source: ${name}`);
  const groupNames = [...new Set(entries.map(e => e.group))].sort();
  const groups = groupNames.map(group => {
    const members = entries.filter(e => e.group === group);
    return {group, files:members.length, bytes:members.reduce((n,e)=>n+e.size,0),
      manifestSha256:stable(members)};
  });
  const ablationCandidates = groups.filter(g => g.group.startsWith('native-resource:')).map(g => {
    const survivors = entries.filter(e => e.group !== g.group);
    return {removeGroup:g.group, removedFiles:g.files, removedBytes:g.bytes,
      remainingFiles:survivors.length, remainingBytes:bytes-g.bytes,
      remainingManifestSha256:stable(survivors), nativeResult:'NOT_RUN'};
  });
  return {schema:'oxce.synthetic.source-footprint.v1',scope:'PUBLIC_ORIGINAL_ONLY_SOURCE_ABLATION_PLAN',
    sourceOnly:true, baseline:{files:entries.length, bytes, manifestSha256:stable(entries)},
    groups, ablationCandidates, unmodifiedStockOXCE:'NOT_RUN',
    selectedSignalFrontierGame:'NOT_RUN', gameplayQualified:false,
    caveat:'These are candidate source-byte removals only. Actual minimum requires per-candidate native startup, resource provenance, gameplay and filesystem evidence.'};
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 3) {
    console.error('usage: node source-footprint.mjs <generated-public-synthetic-root>');
    process.exitCode=2;
  } else {
    try { console.log(JSON.stringify(inspectSource(path.resolve(process.argv[2])),null,2)); }
    catch (e) {console.error('SOURCE_FOOTPRINT_FAIL: '+e.message);process.exitCode=1;}
  }
}
