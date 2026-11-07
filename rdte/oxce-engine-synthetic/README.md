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

Historical native failures included missing GEODATA, GEOGRAPH, UFOGRAPH and TERRAIN placeholders, then an underpopulated original-only GEO.CAT sound set. Original-first-party synthetic resources now pass the **source/native loader gate** on the pinned build; this is not evidence of interactive gameplay or genuine commercial-game compatibility. **Never** present placeholders as authentic X-COM game data or copy proprietary records/graphics.

Private authority: [OXCE DLE #79](https://github.com/SemperSupra/oxce-mod-studio-private/issues/79). This public-source-only PR is not a private-to-public code promotion and does not change #10 review gate, human/automation/agent product acceptance, private PR #88 or release holds. Do not merge/release without existing owner and independent gates.

### VFS mount requirement (native first-failure correction)

The synthetic master declares `loadResources: [SYNTHETIC]`, with generated legacy files under `data/SYNTHETIC/GEODATA`, `GEOGRAPH`, `UFOGRAPH` and `TERRAIN`, rather than loose `data/GEODATA`. Pinned OXCE's `ModInfo::load` and `FileMap::mapExtResources` register named external resource roots in the VFS. Without this mount, `GEODATA/PALETTES.DAT` is invisible even when physically generated. No commercial game art is involved.

## Native ELF cache and evidence-grade separation (2026-11-07)

Native CI still runs the actual pinned OXCE loader against every admitted
fixture-only change. It may **reuse a byte-verified executable**, but it must
not report a cache hit as a new C++ compilation or as a playable game.

- **Cache miss:** build real pinned public C++, record `configure_exit=0`,
  `build_exit=0`, verify ELF64 x86-64 and SHA256, retain origin/source/run
  metadata, and record `seal_verify_exit=0`. GitHub saves the immutable
  branch-scoped cache only following the successful job.
- **Cache hit:** verify exact cache key/source+toolchain+runner-image closure,
  origin metadata, executable permissions, symlink refusal, SHA256, and x86-64
  ELF header. Record `cache_verify_exit=0` instead of a new build.
- **Both paths:** activate only the original-only synthetic master, require
  ruleset and engine-load completion without explicit engine `[ERROR]` or
  `[FATAL]`, and emit `SOURCE_NATIVE_LOAD_PASS` only after these checks.
  A 25-second watchdog exit 124 is a bounded observation, **not a successful
  game session or controlled engine shutdown**.
- **Next validation:** observe a fixture/documentation-only PR change that
  hits the exact unchanged cache and avoids recompilation while still
  executing the real loader. Cache eviction, image drift, manifest mutation,
  stale keys, or a changed workflow/build-source input must fail closed
  or force a fresh admitted build.
- **Non-gates:** synthetic font/menu visual legibility, suppressed/logged
  validation warnings, original UFO/TFTD gameplay, the owner-selected Signal
  Frontier mission, tactical actions, debrief, save/reload, Studio editing,
  human accessibility, and release authorization remain independent.

The first cache-seeding public native run is
[42316034695](https://github.com/SemperSupra/oxce-mod-studio/actions/runs/42316034695).
This documentation correction also provides a source-only same-key replay
without changing the pinned C++ input or workflow cache contract.
