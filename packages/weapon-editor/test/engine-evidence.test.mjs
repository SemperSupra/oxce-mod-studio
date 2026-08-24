import assert from 'node:assert/strict';
import { collectAuthoritativeRuleEvidence, engineEvidenceContract, parseCompileObserverJsonl } from '../src/engine-evidence.mjs';

function event(overrides = {}) {
  return JSON.stringify({
    schema: 1,
    timestamp: '2026-08-22T14:15:30.123Z',
    correlation_id: 'mvp-correlation-1',
    sequence: 1,
    kind: 'snapshot',
    phase: 'validate-rulesets',
    category: 'items',
    operation: 'created-by',
    identity: 'STR_TEST_RIFLE',
    source: 'Core',
    outcome: 'present',
    ...overrides
  });
}

assert.deepEqual(engineEvidenceContract, {
  schema: 1,
  phase: 'validate-rulesets',
  kind: 'snapshot',
  categories: ['items', 'research'],
  operations: ['created-by', 'effective-rule']
});

{
  const text = [
    event(),
    event({ sequence: 2, timestamp: '2026-08-22T14:15:30.124Z', operation: 'effective-rule', source: 'Balance Patch' }),
    event({ sequence: 3, timestamp: '2026-08-22T14:15:30.125Z', kind: 'phase-end', category: '', operation: '', identity: '', source: '', outcome: 'success' }),
    ''
  ].join('\n');

  const parsed = parseCompileObserverJsonl(text);
  assert.equal(parsed.length, 3);
  assert.equal(parsed[0].lineNumber, 1);
  assert.equal(parsed[0].correlationId, 'mvp-correlation-1');
  assert.equal(parsed[0].sequence, 1);
  assert.equal(parsed[0].timestamp, '2026-08-22T14:15:30.123Z');

  const evidence = collectAuthoritativeRuleEvidence(text, {
    category: 'items',
    identity: 'STR_TEST_RIFLE'
  });
  assert.deepEqual(evidence, {
    evidenceOrigin: 'ENGINE-AUTHORITATIVE',
    state: 'available',
    schema: 1,
    phase: 'validate-rulesets',
    category: 'items',
    identity: 'STR_TEST_RIFLE',
    correlationId: 'mvp-correlation-1',
    createdBy: {
      source: 'Core',
      outcome: 'present',
      timestamp: '2026-08-22T14:15:30.123Z',
      sequence: 1,
      lineNumber: 1
    },
    effectiveRule: {
      source: 'Balance Patch',
      outcome: 'present',
      timestamp: '2026-08-22T14:15:30.124Z',
      sequence: 2,
      lineNumber: 2
    }
  });
}

{
  const evidence = collectAuthoritativeRuleEvidence(event({ identity: 'STR_OTHER' }), {
    category: 'items',
    identity: 'STR_TEST_RIFLE'
  });
  assert.equal(evidence.evidenceOrigin, 'ENGINE-AUTHORITATIVE');
  assert.equal(evidence.state, 'not-found');
  assert.equal(evidence.correlationId, 'mvp-correlation-1');
  assert.equal(evidence.createdBy, null);
  assert.equal(evidence.effectiveRule, null);
}

{
  const evidence = collectAuthoritativeRuleEvidence([
    event({ operation: 'created-by', source: '', outcome: 'provenance-unavailable' }),
    event({ sequence: 2, operation: 'effective-rule', source: '', outcome: 'provenance-unavailable' })
  ].join('\n'), {
    category: 'items',
    identity: 'STR_TEST_RIFLE'
  });
  assert.equal(evidence.createdBy.source, null);
  assert.equal(evidence.createdBy.outcome, 'provenance-unavailable');
  assert.equal(evidence.effectiveRule.source, null);
}

// Older schema-1 traces without the additive event envelope remain readable.
{
  const legacy = event({ timestamp: undefined, correlation_id: undefined, sequence: undefined });
  const evidence = collectAuthoritativeRuleEvidence(legacy, {
    category: 'items',
    identity: 'STR_TEST_RIFLE'
  });
  assert.equal(evidence.correlationId, null);
  assert.equal(evidence.createdBy.timestamp, null);
  assert.equal(evidence.createdBy.sequence, null);
}

assert.throws(
  () => parseCompileObserverJsonl('{not-json}\n'),
  /invalid JSON/
);

assert.throws(
  () => parseCompileObserverJsonl(event({ schema: 2 })),
  /unsupported observer schema 2/
);

assert.throws(
  () => parseCompileObserverJsonl(event({ timestamp: '08/22/2026 14:15:30' })),
  /timestamp must be ISO 8601 UTC/
);

assert.throws(
  () => parseCompileObserverJsonl(event({ correlation_id: '' })),
  /correlation_id must not be empty/
);

assert.throws(
  () => parseCompileObserverJsonl([
    event({ sequence: 2 }),
    event({ sequence: 1, operation: 'effective-rule' })
  ].join('\n')),
  /sequence must increase monotonically/
);

assert.throws(
  () => collectAuthoritativeRuleEvidence([
    event(),
    event({ sequence: 2, correlation_id: 'different-correlation', operation: 'effective-rule' })
  ].join('\n'), { category: 'items', identity: 'STR_TEST_RIFLE' }),
  /multiple correlation IDs/
);

assert.throws(
  () => collectAuthoritativeRuleEvidence([
    event(),
    event({ sequence: 2, source: 'Duplicate Core' })
  ].join('\n'), { category: 'items', identity: 'STR_TEST_RIFLE' }),
  /Ambiguous OXCE authoritative evidence/
);

assert.throws(
  () => collectAuthoritativeRuleEvidence(event(), { category: 'units', identity: 'STR_TEST_RIFLE' }),
  /Unsupported authoritative evidence category/
);

console.log('engine evidence tests passed');
