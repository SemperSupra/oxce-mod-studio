# Standalone source-only $0 two-process runtime batch

**Authority:** private DLE #79. Depends on public co-op build draft PR #14 and its immutable GPL co-op binary artifact from native GHA #40894894174. This public PR has **no C++ build workflow**, preventing an expensive recompile when only rules/fixtures/probes change.

The same two real co-op executables are started with separate isolated user/config and two ephemeral local TestServer ports. A bounded Go probe verifies independent main-menu availability; a second Go probe attempts the actual source-audited New Game → Host → Join sequence. Native original logs and structured first failures retained. Paths and artifact hashes pinned.

**Not accepted:** first host/client join, 2-player shared session, multiplayer battle, full a11y, playable game, release. No private Studio code, original retail data, privileged GitHub triggers, auto merge or paid runners.