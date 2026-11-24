# MastermindHost

Game-server manager modelled on HaruHost (non-GSM). Wails v2 (Go) + React/MUI, black and dark-green theme.

**Using the app?** See [GUIDE.md](GUIDE.md) for a basic how-to.

## Scope
- Games: ARK: Survival Ascended (arksa), ARK: Survival Evolved (arkse), Palworld.
- Server folder name always equals the Server Name; renaming a server renames its folder.
- MMO script servers (import-only): Rappelz, Hero Online, Scions of Fate.
- Later phase (not started): Cabal.

## Deployment
- Portable single exe, no installer: `F:\SteamLibrary\steamapps\common\MastermindHost\MasterMindHost.exe` (copied from `build\bin\MasterMindHost.exe`).
- Data (store.json, servers, backups, steamcmd) lives in `MastermindHost-data` next to the exe. Override with the `MASTERMINDHOST_DATA` env var.
- Deployed copy is NOT updated automatically: rebuild, then copy the exe over.

## Features
Create new / import existing servers, SteamCMD install/update, start/stop/restart/shutdown, RCON, players, command line, game settings (ARK ini; Palworld 125 options), Advanced Config (searchable key/value editor over GameUserSettings.ini / Game.ini with a built-in catalog of known ARK settings - rates, player/dino per-level stat multipliers with labelled indices, PvP/PvE, breeding - including keys not yet in the file), config files, mods (ARK, with CurseForge name search that scrapes the public CurseForge page and may break if its markup changes), lists, backups, scheduler, Discord integration, firewall rules, logs, events.

Window memory: the main window size, position and maximised state are saved (every 3 s and on close) and restored on launch; each running game server's own window is moved back to its last position/size shortly after it appears. Importing an ARK SA server resolves mod names from IDs via api.cfwidget.com; the Mods tab also fills any missing names.

Performance tab: per-server and host CPU/RAM graphs (5m/15m/1h, in-memory history, resets on restart) and CPU affinity control (applied on start and when adopting a running server). Ports & A2S tab: ports manager (bound/conflict detection via netstat, firewall status, find free port) and A2S query (server info and players) for a local server or any address.

MMO servers (Rappelz, Hero Online, Scions of Fate, group "MMOs"): import-only script servers, no SteamCMD. Import picks the folder (defaults F:\MMRappelz, F:\DragonzOnline\HeroRebirthFork, F:\HollowFateOnline), discovers the start script and component executables+ports, and the Launch Config tab edits working folder, start command, optional stop command and components. Status is derived from component processes (matched by exe path) and listening ports; Stop closes components gracefully in reverse order, then force-kills. Overview, Launch Config, Scheduler (start/stop/restart/shutdown) and Events tabs only; rename changes the display name, not the folder; Delete is blocked (Remove from list only). Hero has no start script, so its exe is launched directly. Home page groups servers as ASA, ASE, Palworld, MMOs (collapsible); reorder within a group by drag-and-drop or the arrow buttons, persisted in store.json. Cards for Rappelz, Hero and Scions use the artwork from LauncherHub (frontend/src/assets/images/hero.jpg, rappelz.jpg, scions.jpg, mapped by game id in ServerCard.tsx); other games use Steam header art.

ASA maps: TheIsland_WP, ScorchedEarth_WP, TheCenter_WP, Aberration_WP, Extinction_WP, Astraeos_WP, Ragnarok_WP, Valguero_WP, Genesis_WP, LostColony_WP, BobsMissions_WP (Club ARK, needs CurseForge mod 1005639). Dragontopia_WP excluded until released (announced 2026-12-10).

## Build / dev (Git-bash on Windows)
```bash
cd F:/AI/Projects/MasterMindHost
C:/Users/maste/go/bin/wails.exe build -clean
go vet ./... && go test ./...
cd frontend && node node_modules/typescript/bin/tsc --noEmit -p .
```
Dev (testing only): `MASTERMINDHOST_DATA=<temp dir> wails.exe dev -noreload -skipbindings` then http://localhost:34115. Restart dev after Go edits. Regenerate bindings with `wails.exe generate module`.

## Verification status
Verified in dev build: ARK SA and Palworld create/install/start/RCON/shutdown/backup/settings; ASA mods, lists, rename; CurseForge live search.
Verified in dev build (stopped server only): Ports report, find free port, host metrics, affinity core listing, Performance and Ports & A2S tabs render; unit tests for A2S, port scan, affinity mask.
Verified on the packaged exe (2026-11-21): deployment byte-identical to build output (`cmp`); main-window position/size/maximised memory (moved via PowerShell SetWindowPos to 120,90 1111x777, relaunched, restored); mod-name resolution via api.cfwidget.com (1442825 -> Ark Supreme, 1005639 -> Club ARK); app confirmed by the user to function as intended.
Checks green on the final build: gofmt, `go vet ./...`, `go test ./...` (A2S, port scan, affinity mask, advanced-ini list/apply and stat-array catalog), frontend `tsc --noEmit`, `wails build -clean`.
Advanced Config tab: backend list/apply covered by unit tests (round-trip edit, add, remove, Game.ini creation); tab UI verified 2026-11-21 against a real imported server (669 keys rendered, search filter narrows to labelled stat entries). Initial build crashed the whole app to a black screen on opening the tab (useMemo after an early return, "Rendered more hooks than during the previous render") - fixed by moving the hook above the early return; verified rendering in the browser against the dev server before redeploy. Game.ini catalog section corrected to /Script/ShooterGame.ShooterGameMode (was ShootGame, which ARK ignores).
Search performance (2026-11-21): typing in the search box re-rendered all ~670 field components per keystroke (measured 10 long tasks, ~3.1 s main-thread blocking for an 11-character query) and widening a filter back re-mounted them in one render (11.5 s block). Fixed with a 150 ms debounce on the query, a memoised row component with a stable edit callback, and progressive mounting (120 rows, extended by IntersectionObserver as the list scrolls; verified 120 -> 240 -> ... -> 669). After the fix: 0 long tasks while typing, ~1.2 s bounded block when clearing a filter (dev-mode numbers; production build is faster); edit/Revert/Save-enable re-verified in the browser. User-confirmed in the packaged exe 2026-11-21: functional, tested, complete.
Verified live by the user (2026-11-21, packaged exe): server-window repositioning on a running server, affinity applied to the running game process, A2S against the live server, graphs filling from a running server.
MMO servers + grouping + reordering (2026-11-24, dev build, scratch data): import of all three real folders (start command and component discovery; unit test), start via a start.bat with a stand-in executable -> Online, duplicate start rejected, Stop -> Offline with process gone, Delete blocked, reorder by API, arrows and drag-and-drop persisted, script tabs render. User confirmed the new MMO card artwork and card layout look correct in the running app (2026-11-24). NOT verified: starting/stopping the real Rappelz, Hero or Scions stacks, Scions assumed to be F:\HollowFateOnline, crash auto-restart for script servers (not implemented).
NOT verified: ARK SE install/start, Discord webhooks, firewall rules (needs admin), scheduler firing, backup restore.

## Git
GitHub: https://github.com/mastermind-productions/MastermindHost (branch `main`, tracks `origin/main`).
Initial commit `15cfaba`: "Initial commit: MastermindHost base (ARK SA/SE, Palworld) + MMO script servers", author Nathan Williams <mastermind2214@gmail.com> (set repo-local via `git config user.name/user.email`). `build/bin`, `frontend/dist` and `node_modules` are gitignored.
Push from Git-bash: `GIT_TERMINAL_PROMPT=0 timeout 60 git push` (a plain push can hang waiting on a credential prompt).
