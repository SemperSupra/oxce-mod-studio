#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

function usage(message = null) {
  if (message) console.error(message);
  console.error('usage: node .github/scripts/verify-release-rights.mjs [--require-package]');
  process.exit(2);
}

function parseArgs(args) {
  let requirePackage = false;
  for (const arg of args) {
    if (arg === '--require-package') {
      requirePackage = true;
      continue;
    }
    if (arg === '--help' || arg === '-h') usage();
    usage(`unknown argument: ${arg}`);
  }
  return { requirePackage };
}

function readText(file) {
  if (!fs.existsSync(file)) throw new Error(`required release-rights file missing: ${file}`);
  return fs.readFileSync(file, 'utf8');
}

function requireMarkers(file, markers) {
  const text = readText(file);
  const missing = markers.filter(marker => !text.includes(marker));
  if (missing.length) throw new Error(`${file} is missing required markers: ${missing.join(', ')}`);
}

const { requirePackage } = parseArgs(process.argv.slice(2));
const policy = JSON.parse(readText('release-policy.json'));
const licensing = policy.licensing ?? {};
const expected = {
  project_license_spdx: 'Apache-2.0',
  root_license_path: 'LICENSE',
  root_notice_path: 'NOTICE',
  third_party_notices_path: 'THIRD_PARTY_NOTICES',
  package_license_must_match_project: true,
  distributed_package_notice_files: ['LICENSE', 'NOTICE', 'THIRD_PARTY_NOTICES'],
  third_party_distribution_revalidation_required_on_shape_change: true,
  private_development_repository_covered_by_public_license: false
};
if (JSON.stringify(licensing) !== JSON.stringify(expected)) {
  throw new Error(`public licensing policy drifted: ${JSON.stringify(licensing)}`);
}

requireMarkers(licensing.root_license_path, ['Apache License', 'Version 2.0, January 2004']);
requireMarkers(licensing.root_notice_path, ['OXCE Mod Studio', 'OpenXcom Team']);
requireMarkers(licensing.third_party_notices_path, [
  'OpenXcom Ruleset Tools v0.9.45',
  '089e560e60819ec0b35d834f0b6dc65b5ea76d06',
  'License: MIT',
  'Copyright (c) 2018 OpenXcom Team',
  'yaml v2.9.0',
  'License: ISC',
  'Copyright Eemeli Aro <eemeli@gmail.com>',
  'Validation-only components'
]);

const packageRoot = path.join('packages', 'weapon-editor');
const packageJsonPath = path.join(packageRoot, 'package.json');
const pkg = JSON.parse(readText(packageJsonPath));
const packageLicense = pkg.license ?? null;

if (packageLicense !== null && packageLicense !== licensing.project_license_spdx) {
  throw new Error(`weapon-editor license ${JSON.stringify(packageLicense)} disagrees with ${licensing.project_license_spdx}`);
}

const packageNoticePaths = licensing.distributed_package_notice_files.map(name => path.join(packageRoot, name));
if (requirePackage && packageLicense !== licensing.project_license_spdx) {
  throw new Error(`authorized publication requires packages/weapon-editor/package.json license=${licensing.project_license_spdx}`);
}

if (packageLicense === licensing.project_license_spdx || requirePackage) {
  const absent = packageNoticePaths.filter(file => !fs.existsSync(file));
  if (absent.length) throw new Error(`Apache-licensed package is missing distributed notice files: ${absent.join(', ')}`);
  requireMarkers(path.join(packageRoot, 'LICENSE'), ['Apache License', 'Version 2.0, January 2004']);
  requireMarkers(path.join(packageRoot, 'NOTICE'), ['OXCE Mod Studio', 'OpenXcom Team']);
  requireMarkers(path.join(packageRoot, 'THIRD_PARTY_NOTICES'), [
    'OpenXcom Ruleset Tools v0.9.45',
    'License: MIT',
    'yaml v2.9.0',
    'License: ISC'
  ]);
}

console.log('release rights validated');
console.log(`projectLicense ${licensing.project_license_spdx}`);
console.log(`packageLicense ${packageLicense ?? 'pending-governed-projection'}`);
console.log(`packageRequired ${requirePackage}`);
