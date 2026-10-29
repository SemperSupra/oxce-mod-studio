# OXCE 8.6.5 — public original-asset-free native engine smoke

**Purpose:** establish with an *actual* public GitHub-hosted free-tier Ubuntu runner whether the pinned **real** OXCE C++ engine can compile and load a tiny independently authored synthetic master without licensed X-COM game assets.

This test is explicitly **not** the private Mod Studio source, its synthetic-core generator, private test fixtures, or source sanitizer. Everything in this PR is newly authored for public exposure; the engine comes from public upstream GPLv3 source. No private-source material, secrets, or commercial original assets should enter public CI.

- Engine: `MeridianOXC/OpenXcom@668880bb756852f0734cfe92e438e400080a2f4d`, pinned 8.6.5 project baseline.
- Runner: free **public repository** `ubuntu-22.04`, contents-read-only, no secrets, no paid runners.
- Generated master: `rdte-synthetic-master`, `isMaster: true`, `master: ""`, fictional items/research/translation, no external assets.
- Source data: upstream `bin/common` and `bin/standard`, both public *engine* assets. **No** original `UFO` or `TFTD` game folders.
- Build: SDL1.2+CMake dependencies, native compilation in a scratch workdir.
- Launch: `xvfb-run`, 25-second watchdog, isolated `-data`, `-user`, `-config`; original-free status enforced.
- Evidence: native workflow run and job IDs, exact public source and engine SHAs, compiler and engine logs and attempted exit. Build and launch may **fail** as anticipated; evidence is still useful. Engine boot/gameplay are **not passed** solely because the workflow exits zero.

## Interpretation

**BUILD_PASS** means the pinned OXCE C++ executable actually compiled on public GHA.
**ENGINE_LAUNCH_ATTEMPTED** means that executable ran with a self-contained generated master, not that it loaded valid gameplay.
**RESOURCE_GAP** means the native log shows an original asset dependency.
**ENGINE_RULES_PASS** requires engine validation past resource load and a known valid ruleset execution; it cannot be assumed from tests here.
**GAMEPLAY_PASS** additionally requires scripted menu/mission/battle assertions and independent evidence.

The initial engine startup may reject missing GEODATA, GEOGRAPH, UFOGRAPH and TERRAIN legacy assets. This failure is the first falsification target; follow-on work should generate minimal valid original-free legacy binary placeholders and rerun. **Never** present placeholders as authentic X-COM game data or copy proprietary records/graphics.

Private authority: [OXCE DLE #79](https://github.com/SemperSupra/oxce-mod-studio-private/issues/79). This public-source-only PR is not a private-to-public code promotion and does not change #10 review gate, human/automation/agent product acceptance, private PR #88 or release holds. Do not merge/release without existing owner and independent gates.
