# Signal Frontier source-derived fixture preflight

This public, independently authored Go stdlib tool checks a LIMITED subset of source-derived OXCE and OpenXcom Co-op prerequisites. Its source anchors are the pinned public OXCE engine commit 668880bb756852f0734cfe92e438e400080a2f4d and co-op 33f5c251d29b257130cfbb7634735d2e9c946299: nonzero country count and geographic funding data for Mod::newSave, a region and starting base, new-game/geoscape/pause/save UI types and slacking/training indicator IDs, globe texture id 0 for NewBattleState, original co-op base and avatar PNG signatures, disabled retail intro, and no UFO/TFTD directories.

It is a deliberately **restricted text/stanza scanner**, NOT a complete YAML parser, game API verifier, resource loader, playable game, or accessibility assessment. It checks necessary and not sufficient conditions. Do not infer native success from PASS. Work to replace it with a qualified YAML schema/reference graph is tracked in private DLE #79.

The workflow executes public original-only fixture generator code at exact verified source revisions, plus public original Go art generation, on $0 GitHub-hosted runners. No private Mod Studio source, X-COM commercial assets, privileged token, merge or release.
