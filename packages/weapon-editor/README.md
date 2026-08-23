# OXCE Mod Studio — weapon inspector experiment

Status: experimental v0.1 vertical slice.

This package is a VS Code web extension intended to run in VS Code for the Web / github.dev. It operates on ordinary OXCE `.rul` source files. The source remains authoritative.

Current bounded capability:

- recognize explicit local firearm entries with `battleType: 1` **and** bounded partial weapon overrides that inherit `battleType` from the selected master/dependency but contain weapon-signature fields such as `accuracySnap`, `accuracyAimed`, `tuSnap`, `tuAimed`, or `fireSound`;
- display the canonical `STR_*` identity and rule kind;
- label the current view as `SOURCE/TEXT` evidence rather than engine-authoritative evidence;
- edit only already-present supported scalar fields;
- patch the original scalar source span directly;
- preserve comments, unknown fields, ordering and unrelated source text;
- refuse invalid YAML, unsupported fields, missing fields and type-changing edits.

The partial-override recognition rule is intentionally bounded. It does not attempt to reconstruct the master or claim the final effective item kind from source alone. The trusted OXCE oracle remains authoritative for effective master+mod semantics.

The extension deliberately does not claim that a source-valid weapon is engine-valid. OXCE authoritative evidence is a separate layer.

## MVP partial-override fixture

`examples/mvp-weapon-mod/Ruleset/weapon.rul` contains the acceptance seam that matters here:

```yaml
items:
  - type: STR_PISTOL
    accuracySnap: 61
```

The local source does not repeat `battleType`; OXCE 8.6.5 effective-environment evidence has independently established that the rule updates the master-owned `STR_PISTOL`. The inspector must therefore show and edit the local scalar without pretending the local file contains the whole effective entity.

## Development checks

```text
npm install
npm test
npm run build:web
npm run test:web
npm run package:vsix
```

`test:web` requires a Chromium browser installed for Playwright/`@vscode/test-web`; CI installs it explicitly.

## Product principles exercised

1. canonical source remains authoritative;
2. canonical IDs stay visible;
3. structured edits are bounded and inspectable;
4. unknown fields survive;
5. evidence origin is explicit;
6. raw YAML remains first-class;
7. master/dependency inheritance is not flattened into duplicated local source;
8. the web extension works in a virtual workspace rather than requiring a local desktop toolchain.
