# OXCE Mod Studio

Public product/build/release repository for OXCE Mod Studio.

OXCE Mod Studio is a browser-first development environment for ordinary OpenXcom Extended (OXCE) mods. The initial target is VS Code for the Web / `github.dev`: authors keep normal OXCE files as the source of truth while Studio adds source-preserving structured editing, schema/static evidence, pinned target information, and reproducible packaging/release evidence.

## Repository role

This repository is the **public product/generator plane**.

It contains only material intentionally promoted for public use, build, audit, and release. Private engineering/research and the higher-value validation/control plane live elsewhere.

Public releases are built from reviewed public commits in this repository. Release source archives are generated from the exact public candidate commit; they are not produced by archiving a private repository and subtracting known-private files.

## Generator–validator boundary

The project treats specialized validation/evaluation capability as a distinct asset. Public implementation may be inspectable, buildable, forkable, and reproducible while comprehensive private validator knowledge remains private.

Examples of private-by-default validator-side value include comprehensive conformance/adversarial corpora, regression/failure history, specialized domain/ontology knowledge, evaluator/oracle expectations, compatibility/failure fingerprints, and red-team/provenance knowledge when those assets are not required to build/use the public product or satisfy a license.

This classification is semantic, not based on file type.

## Public release contract

Experimental Studio releases currently publish:

- an experimental VSIX built from the public candidate;
- a source snapshot generated from the exact public candidate commit;
- `experimental-build-manifest.json` tying source and binary to that commit;
- `SHA256SUMS.txt`;
- public CI evidence for source-fidelity, VS Code Web activation/edit behavior, pinned provider integration, and packaging;
- an independent credential-free verifier for the public release.

The machine-readable policy is [`release-policy.json`](release-policy.json).

## Evidence model

Studio does not use one ambiguous `valid` state. Evidence remains classified by origin:

- `SOURCE/TEXT`
- `SCHEMA`
- `STATIC-SEMANTIC`
- `ENGINE-AUTHORITATIVE`
- `RUNTIME/BEHAVIORAL`

A lower layer must never imply that a higher layer passed.

## Source/UX principles

- ordinary mod artifacts remain authoritative;
- raw YAML remains first-class;
- structured editors preserve unknown valid content and make minimal source edits;
- canonical OXCE identities remain visible;
- Studio should reduce accidental complexity while preserving and exposing essential domain complexity;
- OXCE remains the authority for engine-effective semantics.

## Current status

The public branch `semper/weapon-editor-exp` is an experimental v0.1 weapon-editor vertical slice. It is not yet a stable community release. Human use in actual `github.dev` and bounded OXCE authoritative evidence remain explicit acceptance gates.
