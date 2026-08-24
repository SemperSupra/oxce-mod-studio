const OBSERVER_SCHEMA = 1;
const AUTHORITATIVE_PHASE = 'validate-rulesets';
const AUTHORITATIVE_KIND = 'snapshot';
const SUPPORTED_CATEGORIES = new Set(['items', 'research']);
const SUPPORTED_OPERATIONS = new Set(['created-by', 'effective-rule']);
const ISO8601_UTC = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?Z$/;

function assertString(value, field, lineNumber) {
  if (typeof value !== 'string') {
    throw new Error(`OXCE observer line ${lineNumber}: ${field} must be a string.`);
  }
  return value;
}

function optionalTimestamp(value, lineNumber) {
  if (value == null) return null;
  const timestamp = assertString(value, 'timestamp', lineNumber);
  if (!ISO8601_UTC.test(timestamp)) {
    throw new Error(`OXCE observer line ${lineNumber}: timestamp must be ISO 8601 UTC.`);
  }
  return timestamp;
}

function optionalCorrelationId(value, lineNumber) {
  if (value == null) return null;
  const correlationId = assertString(value, 'correlation_id', lineNumber);
  if (!correlationId) {
    throw new Error(`OXCE observer line ${lineNumber}: correlation_id must not be empty.`);
  }
  return correlationId;
}

function optionalSequence(value, lineNumber) {
  if (value == null) return null;
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`OXCE observer line ${lineNumber}: sequence must be a positive integer.`);
  }
  return value;
}

export function parseCompileObserverJsonl(text) {
  if (typeof text !== 'string') throw new TypeError('OXCE observer evidence must be JSON Lines text.');

  const events = [];
  const lines = text.split(/\r?\n/);
  let previousSequence = null;
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

    const sequence = optionalSequence(event.sequence, lineNumber);
    if (sequence != null && previousSequence != null && sequence <= previousSequence) {
      throw new Error(`OXCE observer line ${lineNumber}: sequence must increase monotonically.`);
    }
    if (sequence != null) previousSequence = sequence;

    events.push({
      schema: OBSERVER_SCHEMA,
      timestamp: optionalTimestamp(event.timestamp, lineNumber),
      correlationId: optionalCorrelationId(event.correlation_id, lineNumber),
      sequence,
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

function traceCorrelationId(events) {
  const ids = [...new Set(events.map(event => event.correlationId).filter(Boolean))];
  if (ids.length > 1) {
    throw new Error(`OXCE observer evidence contains multiple correlation IDs: ${ids.join(', ')}.`);
  }
  return ids[0] ?? null;
}

export function collectAuthoritativeRuleEvidence(jsonl, { category, identity }) {
  if (!SUPPORTED_CATEGORIES.has(category)) {
    throw new Error(`Unsupported authoritative evidence category: ${String(category)}.`);
  }
  if (typeof identity !== 'string' || identity.length === 0) {
    throw new Error('Authoritative evidence identity must be a non-empty canonical OXCE ID.');
  }

  const parsed = parseCompileObserverJsonl(jsonl);
  const correlationId = traceCorrelationId(parsed);
  const matched = parsed.filter(event =>
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
      correlationId,
      createdBy: null,
      effectiveRule: null
    };
  }

  const created = byOperation.get('created-by') ?? null;
  const effective = byOperation.get('effective-rule') ?? null;
  const provenance = event => event ? {
    source: event.source || null,
    outcome: event.outcome,
    timestamp: event.timestamp,
    sequence: event.sequence,
    lineNumber: event.lineNumber
  } : null;

  return {
    evidenceOrigin: 'ENGINE-AUTHORITATIVE',
    state: 'available',
    schema: OBSERVER_SCHEMA,
    phase: AUTHORITATIVE_PHASE,
    category,
    identity,
    correlationId,
    createdBy: provenance(created),
    effectiveRule: provenance(effective)
  };
}

export const engineEvidenceContract = Object.freeze({
  schema: OBSERVER_SCHEMA,
  phase: AUTHORITATIVE_PHASE,
  kind: AUTHORITATIVE_KIND,
  categories: Object.freeze([...SUPPORTED_CATEGORIES]),
  operations: Object.freeze([...SUPPORTED_OPERATIONS])
});
