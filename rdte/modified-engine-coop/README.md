# Modified OXCE Co-op engine — independent GPL evaluation track

**Source:** OpenXcom-Coop/OpenXcom-Coop-Mod pinned at `33f5c251d29b257130cfbb7634735d2e9c946299`, GPL-3.0. CMake builds actual C++ code on a standard free Ubuntu runner. The build applies the **same opt-in local SDL mailbox and screen-class trace patch** used by the single-player source track; the multiplayer engine also has its own pre-existing TCP/UDP co-op runtime and loopback `OXC_TEST_PORT` JSON TestServer.

This track is **not an accepted multiplayer release**. An actual host/client campaign, online/PvP/hotseat interaction, synchronized Battlescape, shared save, reconnect and accessibility must be qualified independently against **Signal Frontier synthetic content**, without original commercial X-COM files.

Relevant independent upstream test seams:
- `src/CoopMod/TestServer.cpp`: localhost JSON command injection on main game thread.
- `tools/coop_test/boot_check.py`: one-instance smoke.
- `tools/coop_test/test_geoscape_sync.py`: two-instance geoscape sync.
- `tools/coop_test/test_hotseat_start.py`: hotseat.
- `tools/coop_test/test_coop_parallel_turns_latch.py`: multiplayer turn ownership.
- `tools/coop_test/run_matrix.py`: native soak and evidence scoring.

Upstream harness currently bootstraps the `xcom1` master and some tests rely on non-free game data; do not run those unchanged or assert synthetic compatibility. First qualify original data-free startup, adapt fixture provisioning, and only then run isolated two-instance localhost tests with real match/state receipts. Current build lane only verifies source, compilation and compiled binary version. No private product source is copied publicly.

The original Go local control companion has keyboard-facing assistive controls but is **not** yet a full semantic accessibility adapter. Co-op TestServer requires its own permission, transport and evidence review; do not expose it to external networks or count built-in test commands as production network API.

Owner approval, source disclosure, signature, multi-player security, license and human accessibility acceptance are all HELD.
