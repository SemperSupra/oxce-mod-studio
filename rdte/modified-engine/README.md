# Signal Frontier modified OXCE evaluation candidate (Linux x86_64)

**Technical preview, not an accepted playable release.** Built from pinned public GPL OXCE 8.6.5, patched only for opt-in local file-mailbox SDL input and screen-class tracing. Includes public original-only synthetic fixture data and a small localhost Go companion.

To run on compatible Linux with SDL1.2 libraries installed:

1. Extract the technical-preview tarball into its own directory.
2. In that directory run `touch commands.txt; chmod 600 commands.txt; OXCE_RDTE_MAILBOX="$PWD/commands.txt" OXCE_RDTE_TRACE="$PWD/trace.jsonl" ./openxcom-sf-rdte -data "$PWD/data" -user "$PWD/user" -config "$PWD/config"`.
3. Separately launch `./oxce-rdte-broker "$PWD/commands.txt" "$PWD/trace.jsonl" "$PWD/index.html"`.
4. Open its reported `http://127.0.0.1:43819` local web UI and enter its randomly generated token. The page supports keyboard controls and a screen-class status announcement.
5. Use the game window for visual navigation. Do not expose the broker's port publicly, share the token or run it elevated.

**Scope and limitations:** Shared local control seats arbitrate access to one engine instance; they are NOT true multiplayer. The text observation is a C++ state *class name*, not a complete semantic accessible UI tree. Thus assistive-technology equivalence and full game playability remain **unqualified**. Neither HTTP controller nor program adds remote engine networking. The engine source is GPLv3; redistributions and derivative changes must preserve its license and notices. The sound, fonts and graphics in this candidate are independently authored synthetic test assets; production art, music, accessible descriptions and game content are separate work.

Build evidence and artifacts come from public GitHub Actions on a standard $0 runner, with exact upstream source SHA, source patch and hash reports. A green test job is not proof of all screens being usable, co-op sync, gameplay or accessibility. Multiplayer qualification proceeds using a separate pinned GPL OpenXcom-Coop source branch, rather than claiming equivalence.
