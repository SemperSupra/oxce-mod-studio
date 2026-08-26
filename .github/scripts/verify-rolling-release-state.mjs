#!/usr/bin/env node

import crypto from 'node:crypto';
import fs from 'node:fs';
import process from 'node:process';

function usage(message = null) {
  if (message) console.error(message);
  console.error('usage: node .github/scripts/verify-rolling-release-state.mjs [--baseline <file>]');
  process.exit(2);
}

function parseArgs(args) {
  let baseline = '.github/evidence/72-pre-hardening-rolling-state.json';
  for (let i = 0; i < args.length; i += 1) {
    const arg = args[i];
    if (arg === '--help' || arg === '-h') usage();
    if (arg === '--baseline') {
      if (i === args.length - 1 || args[i + 1].startsWith('--')) usage('missing value for --baseline');
      baseline = args[++i];
      continue;
    }
    usage(`unknown argument: ${arg}`);
  }
  return { baseline };
}

function sha256(text) {
  return crypto.createHash('sha256').update(text, 'utf8').digest('hex');
}

function stable(value) {
  if (Array.isArray(value)) return value.map(stable);
  if (value && typeof value === 'object') {
    return Object.fromEntries(Object.keys(value).sort().map(key => [key, stable(value[key])]));
  }
  return value;
}

function equal(left, right) {
  return JSON.stringify(stable(left)) === JSON.stringify(stable(right));
}

async function githubJson(url) {
  const headers = {
    Accept: 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28',
    'User-Agent': 'oxce-mod-studio-publication-negative-proof'
  };
  if (process.env.GITHUB_TOKEN) headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
  const response = await fetch(url, { headers });
  if (!response.ok) {
    throw new Error(`GitHub API ${response.status} ${response.statusText} for ${url}: ${await response.text()}`);
  }
  return response.json();
}

const { baseline: baselinePath } = parseArgs(process.argv.slice(2));
const baseline = JSON.parse(fs.readFileSync(baselinePath, 'utf8'));
if (baseline.schema !== 1 || baseline.kind !== 'oxce-mod-studio-rolling-release-state') {
  throw new Error(`unsupported baseline ${baselinePath}`);
}
if (!/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(baseline.repository ?? '')) {
  throw new Error(`unsafe repository in baseline: ${baseline.repository}`);
}
if (!/^[A-Za-z0-9._-]+$/.test(baseline.tag ?? '')) {
  throw new Error(`unsafe tag in baseline: ${baseline.tag}`);
}

const api = `https://api.github.com/repos/${baseline.repository}`;
const ref = await githubJson(`${api}/git/ref/tags/${encodeURIComponent(baseline.tag)}`);
const release = await githubJson(`${api}/releases/tags/${encodeURIComponent(baseline.tag)}`);

const observed = {
  schema: 1,
  kind: baseline.kind,
  repository: baseline.repository,
  tag: baseline.tag,
  tagCommitSha: ref.object?.sha ?? null,
  release: {
    id: release.id,
    tagName: release.tag_name,
    targetCommitish: release.target_commitish,
    name: release.name,
    draft: release.draft,
    prerelease: release.prerelease,
    immutable: release.immutable ?? false,
    createdAt: release.created_at,
    updatedAt: release.updated_at,
    publishedAt: release.published_at,
    bodySha256: sha256(release.body ?? '')
  },
  assets: (release.assets ?? []).map(asset => ({
    id: asset.id,
    name: asset.name,
    size: asset.size,
    digest: asset.digest ?? null,
    createdAt: asset.created_at,
    updatedAt: asset.updated_at
  })).sort((a, b) => a.id - b.id)
};

const expected = {
  schema: baseline.schema,
  kind: baseline.kind,
  repository: baseline.repository,
  tag: baseline.tag,
  tagCommitSha: baseline.tagCommitSha,
  release: baseline.release,
  assets: [...baseline.assets].sort((a, b) => a.id - b.id)
};

if (!equal(observed, expected)) {
  console.error('ROLLING RELEASE STATE CHANGED');
  console.error(JSON.stringify({ expected, observed }, null, 2));
  process.exit(1);
}

console.log('ROLLING RELEASE STATE UNCHANGED');
console.log(`repository ${observed.repository}`);
console.log(`tag ${observed.tag}`);
console.log(`tagCommitSha ${observed.tagCommitSha}`);
console.log(`releaseId ${observed.release.id}`);
console.log(`releaseUpdatedAt ${observed.release.updatedAt}`);
console.log(`assets ${observed.assets.length}`);
