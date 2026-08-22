import assert from 'node:assert/strict';
import { collectAuthoritativeRuleEvidence, parseCompileObserverJsonl } from '../src/engine-evidence.mjs';

const created = JSON.stringify({schema:1,kind:'snapshot',phase:'validate-rulesets',category:'items',operation:'created-by',identity:'STR_TEST_RIFLE',source:'Core',outcome:'present'});
const effective = JSON.stringify({schema:1,kind:'snapshot',phase:'validate-rulesets',category:'items',operation:'effective-rule',identity:'STR_TEST_RIFLE',source:'Balance Patch',outcome:'present'});

const evidence = collectAuthoritativeRuleEvidence(`${created}\n${effective}\n`, {category:'items', identity:'STR_TEST_RIFLE'});
assert.equal(evidence.evidenceOrigin, 'ENGINE-AUTHORITATIVE');
assert.equal(evidence.state, 'available');
assert.equal(evidence.createdBy.source, 'Core');
assert.equal(evidence.effectiveRule.source, 'Balance Patch');

const missing = collectAuthoritativeRuleEvidence(created, {category:'items', identity:'STR_OTHER'});
assert.equal(missing.state, 'not-found');

assert.throws(() => parseCompileObserverJsonl('{bad json}'), /invalid JSON/);
assert.throws(() => parseCompileObserverJsonl(created.replace('"schema":1', '"schema":2')), /unsupported observer schema 2/);
assert.throws(() => collectAuthoritativeRuleEvidence(`${created}\n${created}`, {category:'items', identity:'STR_TEST_RIFLE'}), /Ambiguous OXCE authoritative evidence/);
assert.throws(() => collectAuthoritativeRuleEvidence(created, {category:'units', identity:'STR_TEST_RIFLE'}), /Unsupported authoritative evidence category/);

console.log('engine evidence tests passed');
