import test from 'node:test';
import assert from 'node:assert/strict';
import {readCampaign} from './validate.mjs';
import {emitSubset,key} from './emit-oxce.mjs';
const campaign=readCampaign(new URL('./campaign.json',import.meta.url));
test('canonical identifiers are deterministic and sanitized',()=>{
 assert.equal(key('quiet-array-signals'),'STR_SF_QUIET_ARRAY_SIGNALS');
 assert.throws(()=>key('../xcom1'),/unsafe-identifier/);
});
test('emitted subset preserves research graph and ammo dependency',()=>{
 const d=emitSubset(campaign);
 for(const k of ['STR_SF_QUIET_ARRAY_SIGNALS','STR_SF_PULSE_COUPLING','STR_SF_RESCUE_PROTOCOLS','STR_SF_PULSE_EMITTER','STR_SF_CHARGED_CELL'])assert.ok(d.ruleset.includes(k),k);
 assert.match(d.ruleset,/compatibleAmmo: \[STR_SF_CHARGED_CELL\]/);
 assert.match(d.ruleset,/dependencies: \[STR_SF_QUIET_ARRAY_SIGNALS\]/);
});
test('emitter declares all unsupported gameplay domains explicitly',()=>{
 const d=emitSubset(campaign);
 assert.equal(d.report.engineLoadVerified,false);
 assert.equal(d.report.playable,false);
 assert.ok(d.report.excludedDomains.includes('startingBase'));
 assert.ok(d.report.unmappedEquipment.includes('scanner'));
});
