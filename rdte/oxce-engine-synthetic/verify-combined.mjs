#!/usr/bin/env node
// Public ORIGINAL-ONLY source integration. Never launches native OXCE or dispatches CI.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scripts = path.dirname(fileURLToPath(import.meta.url));
const expected = [
  'config/options.cfg',
  'user/mods/rdte-synthetic-master/metadata.yml',
  'user/mods/rdte-synthetic-master/Ruleset/10-synthetic.rul',
  'user/mods/rdte-synthetic-master/Language/en-US.yml',
  'user/mods/rdte-synthetic-master/Language/rdte-fonts.yml',
  'user/mods/rdte-synthetic-master/Language/rdte-fonts.png',
  'data/SYNTHETIC/rdte/synthetic-legacy-manifest.json',
  'data/SYNTHETIC/MAPS/SFREL00.MAP',
  'data/SYNTHETIC/ROUTES/SFREL00.RMP',
  'data/SYNTHETIC/TERRAIN/BLANKS.MCD',
];

async function run(script, args = [], cwd = scripts) {
  const file = path.join(cwd, script);
  return await new Promise((resolve, reject) => {
    const child = spawn(process.execPath, [file, ...args], { cwd, stdio: ['ignore', 'pipe', 'pipe'], env: { ...process.env, CI: 'true' } });
    let stdout = '', stderr = '';
    child.stdout.on('data', b => { stdout += String(b); if (stdout.length > 2 ** 20) child.kill(); });
    child.stderr.on('data', b => { stderr += String(b); if (stderr.length > 2 ** 20) child.kill(); });
    child.on('error', reject);
    child.on('close', (code, signal) => {
      if (code !== 0 || signal) return reject(new Error(`${script}: exit=${code} signal=${signal} stderr=${stderr.slice(-2500)} stdout=${stdout.slice(-1500)}`));
      resolve({ script, exit: code, output: stdout.trim().slice(-2000) });
    });
  });
}

function index(root) {
  const files = [];
  function walk(current) {
    for (const entry of fs.readdirSync(current, { withFileTypes: true }).sort((a,b) => a.name.localeCompare(b.name, 'en'))) {
      const absolute = path.join(current, entry.name);
      if (entry.isSymbolicLink()) throw Error(`symlink in synthetic root: ${absolute}`);
      if (entry.isDirectory()) { walk(absolute); continue; }
      if (!entry.isFile()) throw Error(`nonregular synthetic output: ${absolute}`);
      const rel = path.relative(root, absolute).split(path.sep).join('/');
      if (/^(?:data\/(?:UFO|TFTD)(?:\/|$)|user\/mods\/(?:xcom1|xcom2)(?:\/|$))/i.test(rel)) throw Error(`commercial game path present: ${rel}`);
      const bytes = fs.readFileSync(absolute);
      files.push({ path: rel, size: bytes.length, sha256: createHash('sha256').update(bytes).digest('hex') });
    }
  }
  walk(root);
  files.sort((a, b) => a.path < b.path ? -1 : a.path > b.path ? 1 : 0);
  for (const rel of expected) assert.ok(files.find(f => f.path === rel), `missing integrated output ${rel}`);
  assert.ok(files.length >= 80, `expected a full synthetic master plus original resource files; got ${files.length}`);
  const snapshot = JSON.stringify(files);
  return { files, manifestSha256: createHash('sha256').update(snapshot).digest('hex') };
}

async function build(root, scriptsDir) {
  fs.mkdirSync(root, { recursive: true });
  const work = [];
  work.push(await run('generate.mjs', [root], scriptsDir));
  work.push(await run('generate-font.mjs', [root], scriptsDir));
  work.push(await run('generate-legacy.mjs', [path.join(root, 'data/SYNTHETIC')], scriptsDir));
  for (const test of ['test-generated-master.mjs', 'test-generated-font.mjs', 'test-stock-spawn-prerequisites.mjs']) {
    work.push(await run(test, [root], scriptsDir));
  }
  return { work, ...index(root) };
}

export async function verify({ workdir = null, scriptsDir = scripts } = {}) {
  // scriptsDir is injected by offline tests only; production CLI has no override.
  const root = fs.mkdtempSync(path.join(workdir ?? os.tmpdir(), 'oxce-original-combined-'));
  try {
    const [a,b] = await Promise.all([build(path.join(root, 'run-a'), scriptsDir), build(path.join(root, 'run-b'), scriptsDir)]);
    assert.deepEqual(a.files, b.files, 'generated original fixture differs across independent runs');
    const independent = await run('test-generated-legacy.mjs', [], scriptsDir);
    return { result:'PASS', scope:'public original-only generated source integration', files:a.files.length, manifestSha256:a.manifestSha256, replay:'BYTE_IDENTICAL', tests:[...a.work.map(x=>x.script), independent.script], runBTests:b.work.map(x=>x.script), nativeOXCE:'NOT_RUN', selectedSignalFrontierGame:'NOT_RUN' };
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 2) { console.error('usage: node verify-combined.mjs'); process.exitCode = 2; }
  else verify().then(r=>console.log(JSON.stringify(r,null,2))).catch(e=>{console.error(`COMBINED_SOURCE_FAIL: ${e.message}`);process.exitCode=1;});
}
