# OXCE Mod Studio — experimental weapon inspector

This public package is an experimental VS Code web-extension slice for OXCE Mod Studio.

It keeps ordinary `.rul` source authoritative, exposes canonical IDs/kinds, labels current evidence as `SOURCE/TEXT`, and edits only bounded existing scalar source spans. It does not claim engine-validity; OXCE authoritative evidence is a separate layer.

Current public checks cover:

- source-fidelity unit tests;
- browser bundling;
- real VS Code Web/Chromium activation;
- command registration;
- opening the weapon inspector in a virtual workspace;
- applying a bounded edit without disturbing comments/unknown fields;
- VSIX packaging.
