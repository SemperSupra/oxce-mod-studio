# Signal Frontier three-audience contract baseline (SF07)

This is an **independent original Go in-memory simulator contract**, intentionally not connected to OpenXcom Extended. It makes the human, automation, and agent action envelope reviewable using the same source-neutral state, request IDs, expected sequence, deterministic hashes and fail-closed errors.

It is not a replacement for the actual OXCE rules engine; acceptance requires a future native adapter, input injection/screenshot/reading contract and engine receipts. Human screen-reader, real UI, gameplay, engine execution, and action parity are all **unqualified**.

Actions: scan-relay, end-turn, recover-recorder; read-only snapshot always permitted. Unknown audience, stale sequencing, missing request IDs, unscanned recovery and undocumented actions are refused without state mutation.

No source from private Mod Studio, no commercial game assets, no production controller, no release authorization. Authority: private DLE #79; independent public original-source design.
