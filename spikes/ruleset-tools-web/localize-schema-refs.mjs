#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

const schemaDir = path.resolve(process.argv[2] ?? '');
if (!process.argv[2] || !fs.existsSync(schemaDir)) {
  throw new Error('usage: node localize-schema-refs.mjs <generated-schema-dir>');
}

const OPENXCOM_SCHEMA = /^https:\/\/openxcom\.org\/schemas\/(?:oxc|oxce)\/([^#]+)(#.*)?$/;
let replacements = 0;
let filesChanged = 0;

function localize(value) {
  if (typeof value === 'string') {
    const match = OPENXCOM_SCHEMA.exec(value);
    if (match) {
      replacements += 1;
      return `${match[1]}${match[2] ?? ''}`;
    }
    return value;
  }
  if (Array.isArray(value)) {
    return value.map(localize);
  }
  if (value && typeof value === 'object') {
    for (const [key, nested] of Object.entries(value)) {
      value[key] = localize(nested);
    }
  }
  return value;
}

for (const file of fs.readdirSync(schemaDir).filter(name => name.endsWith('.json')).sort()) {
  const filePath = path.join(schemaDir, file);
  const original = fs.readFileSync(filePath, 'utf8');
  const document = JSON.parse(original);
  const before = replacements;
  localize(document);
  if (replacements !== before) {
    fs.writeFileSync(filePath, `${JSON.stringify(document, null, 2)}\n`);
    filesChanged += 1;
  }
}

if (replacements === 0) {
  throw new Error('no OpenXcom absolute schema references were found; pinned schema shape may have drifted');
}

for (const file of fs.readdirSync(schemaDir).filter(name => name.endsWith('.json'))) {
  const text = fs.readFileSync(path.join(schemaDir, file), 'utf8');
  if (text.includes('https://openxcom.org/schemas/')) {
    throw new Error(`absolute OpenXcom schema reference remains in ${file}`);
  }
}

console.log(JSON.stringify({
  schemaDir,
  filesChanged,
  replacements,
  policy: 'openxcom.org schema URLs -> relative bundled-schema references'
}, null, 2));
