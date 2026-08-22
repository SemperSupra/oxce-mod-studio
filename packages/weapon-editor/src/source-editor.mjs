import { isMap, isScalar, parseDocument } from 'yaml';

export const EDITABLE_WEAPON_FIELDS = Object.freeze([
  'accuracySnap',
  'accuracyAimed',
  'tuSnap',
  'tuAimed',
  'weight',
  'costBuy',
  'costSell',
  'bigSprite',
  'fireSound'
]);

export const SOURCE_EVIDENCE_ORIGIN = 'SOURCE/TEXT';

const EDITABLE_FIELD_SET = new Set(EDITABLE_WEAPON_FIELDS);

function parseRuleset(source) {
  if (typeof source !== 'string') throw new Error('source must be a string');
  const doc = parseDocument(source, {
    keepSourceTokens: true,
    prettyErrors: true,
    strict: true
  });
  if (doc.errors.length > 0) {
    const message = doc.errors.map(error => error.message).join('\n');
    throw new Error(`Refusing structured edit because YAML is invalid:\n${message}`);
  }
  const items = doc.get('items', true);
  if (!items || !Array.isArray(items.items)) {
    throw new Error('Expected an items sequence in the ruleset document.');
  }
  return { items };
}

function findWeaponMap(items, weaponId) {
  const matches = [];
  for (const entry of items.items) {
    if (!isMap(entry)) continue;
    const typeNode = entry.get('type', true);
    if (!isScalar(typeNode) || typeNode.value !== weaponId) continue;
    matches.push(entry);
  }
  if (matches.length === 0) throw new Error(`Weapon ${weaponId} was not found.`);
  if (matches.length > 1) throw new Error(`Weapon ${weaponId} is ambiguous in this document.`);
  return matches[0];
}

function scalarValue(entry, field) {
  const node = entry.get(field, true);
  return isScalar(node) ? node.value : undefined;
}

function validateReplacementType(field, current, value) {
  if (typeof current === 'number') {
    if (typeof value !== 'number' || !Number.isFinite(value)) throw new Error(`${field} must remain a finite number.`);
    return;
  }
  if (typeof current === 'string') {
    if (typeof value !== 'string') throw new Error(`${field} must remain a string.`);
    return;
  }
  if (typeof current === 'boolean') {
    if (typeof value !== 'boolean') throw new Error(`${field} must remain a boolean.`);
    return;
  }
  throw new Error(`Field ${field} is not a supported scalar type.`);
}

function renderScalar(value) {
  if (typeof value === 'number') return String(value);
  if (typeof value === 'boolean') return value ? 'true' : 'false';
  if (typeof value === 'string') return JSON.stringify(value);
  throw new Error('Unsupported scalar replacement value.');
}

function replaceScalarSourceSpan(source, node, value) {
  if (!Array.isArray(node.range) || node.range.length < 2) {
    throw new Error('YAML parser did not expose the scalar source range; refusing broad rewrite.');
  }
  const [start, valueEnd] = node.range;
  if (!Number.isInteger(start) || !Number.isInteger(valueEnd) || start < 0 || valueEnd < start || valueEnd > source.length) {
    throw new Error('YAML parser returned an invalid scalar source range; refusing broad rewrite.');
  }
  return source.slice(0, start) + renderScalar(value) + source.slice(valueEnd);
}

export function readWeaponDocument(source) {
  const { items } = parseRuleset(source);
  const weapons = [];
  for (const entry of items.items) {
    if (!isMap(entry)) continue;
    const type = scalarValue(entry, 'type');
    if (typeof type !== 'string' || type.length === 0) continue;
    if (scalarValue(entry, 'battleType') !== 1) continue;

    const fields = {};
    for (const field of EDITABLE_WEAPON_FIELDS) {
      const value = scalarValue(entry, field);
      if (value !== undefined) fields[field] = value;
    }
    weapons.push({id: type, kind: 'item', evidenceOrigin: SOURCE_EVIDENCE_ORIGIN, fields});
  }
  return { weapons };
}

export function editWeaponScalar(source, { weaponId, field, value }) {
  if (typeof weaponId !== 'string' || weaponId.length === 0) throw new Error('weaponId must be non-empty');
  if (!EDITABLE_FIELD_SET.has(field)) throw new Error(`Field ${field} is not editable in the v0.1 weapon form.`);

  const { items } = parseRuleset(source);
  const weapon = findWeaponMap(items, weaponId);
  if (scalarValue(weapon, 'battleType') !== 1) throw new Error(`${weaponId} is not a firearm weapon (battleType 1).`);

  const node = weapon.get(field, true);
  if (!isScalar(node)) throw new Error(`Weapon ${weaponId} has no existing scalar ${field} field to edit.`);

  const current = node.value;
  validateReplacementType(field, current, value);
  if (Object.is(current, value)) return source;
  return replaceScalarSourceSpan(source, node, value);
}
