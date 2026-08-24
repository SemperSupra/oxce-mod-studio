# Public Agent Contract — OXCE Mod Studio

This repository is the sanitized public product/build/release plane for OXCE Mod Studio.

It follows the project’s **generator–validator** operating model: public implementation/build mechanics may be inspectable and reproducible, while high-value validator-side domain, conformance, adversarial, evaluator, and failure knowledge remains private by default unless deliberately released or required by license/public use.

## Required reading

Before changing build, release, provenance, or public/private promotion behavior, read:

1. `README.md`
2. `release-policy.json`
3. the governing issue/PR and current comments

## Public repository invariants

- Work here must be public-safe by design.
- Do not copy private Git history into this repository.
- Do not import private research, comprehensive conformance/adversarial corpora, evaluator/oracle knowledge, regression/failure history, or other private validation/control-plane material merely because it would improve public tests.
- Public tests should prove the public contract; comprehensive private validators belong in the private control plane unless deliberately released.
- File type/path is not a sufficient public-safety decision.

## Release invariants

- Build public binaries only from a reviewed public candidate commit.
- Generate release source snapshots from that exact public commit.
- Never generate a release source archive from a private checkout.
- Release manifest and source/binary hashes must agree on the exact public commit.
- Preserve applicable source-license obligations.
- `studio-exp-current` is mutable and must not be treated as an immutable package-manager version.
- Independently verify published release assets without credentials when possible.

## Evidence invariants

Preserve evidence origin:

- `SOURCE/TEXT`
- `SCHEMA`
- `STATIC-SEMANTIC`
- `ENGINE-AUTHORITATIVE`
- `RUNTIME/BEHAVIORAL`

Do not upgrade the meaning of a result. A browser/static check cannot claim engine-authoritative behavior.

## Product invariants

- ordinary OXCE mod files remain authoritative;
- raw source remains first-class;
- structured edits are minimal and source-preserving;
- canonical OXCE IDs/kinds/source locations remain visible;
- OXCE is the authority for engine-effective semantics;
- reduce accidental complexity while preserving/exposing essential complexity;
- agent-generated edits should remain inspectable through normal source diffs and, when available, semantic/effective evidence.

## Handoff

Leave durable PR/issue evidence with the exact head SHA, public artifacts changed, validation run, release/provenance state, and any residual blocker. Do not rely on chat-only context.
