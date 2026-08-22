const OBSERVER_SCHEMA = 1;
const AUTHORITATIVE_PHASE = 'validate-rulesets';
const AUTHORITATIVE_KIND = 'snapshot';
const SUPPORTED_CATEGORIES = new Set(['items', 'research']);
const SUPPORTED_OPERATIONS = new Set(['created-by', 'effective-rule']);

function assertString(value, field, lineNumber) {
  if (typeof value !== 'string') {
    throw new Error(`OXCE observer line ${lineNumber}: ${field} must be a string.`);
  }
  return value;
}

export function parseCompileObserverJsonl(text) {
  if (typeof text !== 'string') throw new TypeError('OXCE observer evidence must be JSON Lines text.');

  const events = [];
  const lines = text.split(/\r?\n/);
  for (let index = 0; index < lines.length; index += 1) {
    const raw = lines[index].trim();
    if (!raw) continue;
    const lineNumber = index + 1;

    let event;
    try {
      event = JSON.parse(raw);
    } catch (error) {
      throw new Error(`OXCE observer line ${lineNumber}: invalid JSON (${error.message}).`);
    }

    if (!event || Array.isArray(event) || typeof event !== 'object') {
      throw new Error(`OXCE observer line ${lineNumber}: event must be an object.`);
    }
    if (event.schema !== OBSERVER_SCHEMA) {
      throw new Error(`OXCE observer line ${lineNumber}: unsupported observer schema ${String(event.schema)}.`);
    }

    events.push({
      schema: OBSERVER_SCHEMA,
      kind: assertString(event.kind, 'kind', lineNumber),
      phase: assertString(event.phase, 'phase', lineNumber),
      category: assertString(event.category, 'category', lineNumber),
      operation: assertString(event.operation, 'operation', lineNumber),
      identity: assertString(event.identity, 'identity', lineNumber),
      source: assertString(event.source, 'source', lineNumber),
      outcome: assertString(event.outcome, 'outcome', lineNumber),
      lineNumber
    });
  }
  return events;
}

export function collectAuthoritativeRuleEvidence(jsonl, { category, identity }) {
  if (!SUPPORTED_CATEGORIES.has(category)) {
    throw new Error(`Unsupported authoritative evidence category: ${String(category)}.`);
  }
  if (typeof identity !== 'string' || identity.length === 0) {
    throw new Error('Authoritative evidence identity must be a non-empty canonical OXCE ID.');
  }

  const matched = parseCompileObserverJsonl(jsonl).filter(event =>
    event.kind === AUTHORITATIVE_KIND &&
    event.phase === AUTHORITATIVE_PHASE &&
    event.category === category &&
    event.identity === identity &&
    SUPPORTED_OPERATIONS.has(event.operation)
  );

  const byOperation = new Map();
  for (const event of matched) {
    if (byOperation.has(event.operation)) {
      const previous = byOperation.get(event.operation);
      throw new Error(
        `Ambiguous OXCE authoritative evidence for ${identity}: ${event.operation} appears on lines ${previous.lineNumber} and ${event.lineNumber}.`
      );
    }
    byOperation.set(event.operation, event);
  }

  if (byOperation.size === 0) {
    return {
      evidenceOrigin: 'ENGINE-AUTHORITATIVE',
      state: 'not-found',
      schema: OBSERVER_SCHEMA,
      phase: AUTHORITATIVE_PHASE,
      category,
      identity,
      createdBy: null,
      effectiveRule: null
    };
  }

  const created = byOperation.get('created-by') ?? null;
  const effective = byOperation.get('effective-rule') ?? null;
  return {
    evidenceOrigin: 'ENGINE-AUTHORITATIVE',
    state: 'available',
    schema: OBSERVER_SCHEMA,
    phase: AUTHORITATIVE_PHASE,
    category,
    identity,
    createdBy: created ? {
      source: created.source || null,
      outcome: created.outcome,
      lineNumber: created.lineNumber
    } : null,
    effectiveRule: effective ? {
      source: effective.source || null,
      outcome: effective.outcome,
      lineNumber: effective.lineNumber
    } : null
  };
}

export const engineEvidenceContract = Object.freeze({
  schema: OBSERVER_SCHEMA,
  phase: AUTHORITATIVE_PHASE,
  kind: AUTHORITATIVE_KIND,
  categories: Object.freeze([...SUPPORTED_CATEGORIES]),
  operations: Object.freeze([...SUPPORTED_OPERATIONS])
});
