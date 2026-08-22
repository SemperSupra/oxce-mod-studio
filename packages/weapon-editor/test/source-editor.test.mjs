#!/usr/bin/env node

import assert from 'node:assert/strict';
import { editWeaponScalar, readWeaponDocument, SOURCE_EVIDENCE_ORIGIN } from '../src/source-editor.mjs';

const source = `# document comment\nitems:\n  # target weapon comment\n  - type: STR_TEST_WEAPON\n    battleType: 1\n    accuracySnap: 65 # preserve inline comment\n    accuracyAimed: 90\n    weight: 4\n    futureField: 123 # unknown field must survive\n  - type: STR_OTHER_ITEM\n    battleType: 2\n    clipSize: 12\n`;

assert.deepEqual(readWeaponDocument(source), {
  weapons: [{
    id: 'STR_TEST_WEAPON',
    kind: 'item',
    evidenceOrigin: SOURCE_EVIDENCE_ORIGIN,
    fields: {accuracySnap: 65, accuracyAimed: 90, weight: 4}
  }]
});

const edited = editWeaponScalar(source, {weaponId: 'STR_TEST_WEAPON', field: 'accuracySnap', value: 70});
assert.equal(edited, source.replace('accuracySnap: 65 # preserve inline comment', 'accuracySnap: 70 # preserve inline comment'));
assert.match(edited, /futureField: 123 # unknown field must survive/);
assert.match(edited, /# target weapon comment/);

const noOp = editWeaponScalar(source, {weaponId: 'STR_TEST_WEAPON', field: 'accuracySnap', value: 65});
assert.equal(noOp, source, 'no-op structured edit must produce byte-identical source');

const oddFormatting = `items:\n- { type: STR_COMPACT_WEAPON, battleType: 1, accuracySnap: 55, futureField: [1, 2, 3] } # preserve compact form\n`;
const oddEdited = editWeaponScalar(oddFormatting, {weaponId: 'STR_COMPACT_WEAPON', field: 'accuracySnap', value: 56});
assert.equal(oddEdited, oddFormatting.replace('accuracySnap: 55', 'accuracySnap: 56'));

const partialOverride = `items:\n  - type: STR_PISTOL\n    accuracySnap: 61\n  - type: STR_PISTOL_CLIP\n    costBuy: 25\n`;
assert.deepEqual(readWeaponDocument(partialOverride), {
  weapons: [{
    id: 'STR_PISTOL',
    kind: 'item',
    evidenceOrigin: SOURCE_EVIDENCE_ORIGIN,
    fields: {accuracySnap: 61}
  }]
});
assert.equal(
  editWeaponScalar(partialOverride, {weaponId: 'STR_PISTOL', field: 'accuracySnap', value: 62}),
  partialOverride.replace('accuracySnap: 61', 'accuracySnap: 62')
);

const genericOnlyOverride = `items:\n  - type: STR_GENERIC_ITEM\n    weight: 3\n    costBuy: 10\n`;
assert.deepEqual(readWeaponDocument(genericOnlyOverride), {weapons: []});
assert.throws(
  () => editWeaponScalar(genericOnlyOverride, {weaponId: 'STR_GENERIC_ITEM', field: 'weight', value: 4}),
  /not a bounded firearm weapon source node/
);

const explicitNonFirearmWithWeaponishField = `items:\n  - type: STR_EXPLICIT_AMMO\n    battleType: 2\n    accuracySnap: 10\n`;
assert.deepEqual(readWeaponDocument(explicitNonFirearmWithWeaponishField), {weapons: []});
assert.throws(
  () => editWeaponScalar(explicitNonFirearmWithWeaponishField, {weaponId: 'STR_EXPLICIT_AMMO', field: 'accuracySnap', value: 11}),
  /not a bounded firearm weapon source node/
);

assert.throws(() => editWeaponScalar('items: [\n', {weaponId: 'STR_TEST_WEAPON', field: 'accuracySnap', value: 70}), /Refusing structured edit because YAML is invalid/);
assert.throws(() => editWeaponScalar(source, {weaponId: 'STR_TEST_WEAPON', field: 'power', value: 40}), /not editable/);
assert.throws(() => editWeaponScalar(source, {weaponId: 'STR_TEST_WEAPON', field: 'tuSnap', value: 30}), /no existing scalar tuSnap field/);
assert.throws(() => editWeaponScalar(source, {weaponId: 'STR_TEST_WEAPON', field: 'accuracySnap', value: '70'}), /must remain a finite number/);
assert.throws(() => editWeaponScalar(source, {weaponId: 'STR_OTHER_ITEM', field: 'weight', value: 2}), /not a bounded firearm weapon source node/);

console.log(JSON.stringify({
  result: 'pass',
  evidenceOrigin: SOURCE_EVIDENCE_ORIGIN,
  proven: [
    'direct-source-span-edit',
    'byte-identical-no-op',
    'inline-comment-preserved',
    'unknown-field-preserved',
    'compact-formatting-preserved',
    'partial-weapon-override-recognized',
    'generic-only-override-not-promoted-to-weapon',
    'explicit-non-firearm-wins-over-source-shape',
    'canonical-kind-visible'
  ]
}, null, 2));
