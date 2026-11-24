# MastermindHost — Basic Guide

A short how-to for day-to-day use of MastermindHost, the black-and-dark-green game server manager for ARK: Survival Ascended, ARK: Survival Evolved and Palworld (plus imported MMO script servers).

---

## 1. First run

MastermindHost is a portable single exe — there is no installer.

1. Run `MasterMindHost.exe`. Everything the app creates lives in a `MastermindHost-data` folder next to the exe (server folders, `store.json`, backups, SteamCMD). You can move the whole exe + data folder pair anywhere together.
2. Open **Settings** (gear icon, top right) and check two folders:
   - **Servers folder (new servers)** — where *newly created* servers are placed. Existing/imported servers keep their own folders.
   - **Default backups folder** — where backups go unless a server overrides it.
3. SteamCMD is downloaded automatically the first time you install a server — nothing to set up.

The top bar always shows host **CPU**, **RAM** and **Disk** usage, plus a **Check for game updates** button (refresh icon) that queries Steam for every installed server at once and flags cards with **Update available**.

## 2. The home screen

Servers are shown as cards in collapsible groups: **ASA**, **ASE**, **Palworld** and **MMOs** (any other group appears if you create one). Each card shows:

- The game name and a coloured status chip (STOPPED / STARTING / ONLINE / …).
- Live fields: players, map, port, CPU, RAM, uptime, build ID, mod count and whether **Auto-restart** is on. MMO cards show components up (e.g. `2 / 3`) instead.
- Action buttons (Install / Start / Shutdown / Restart / Stop / Update — see below).
- **Configure** to open the server's page.

Reordering: drag a card onto another card in the same group, or use the **< / >** arrows on the card. The order is saved.

## 3. Adding a server

**Create Server** (top right) opens a dialog with two tabs:

- **Create new** — pick the game, type a **Server name**, press **Create**. The dialog shows the folder it will create. The folder name always equals the server name (illegal characters are replaced with `_`); renaming the server later renames the folder too.
- **Import existing** — pick the game and **Browse** to the existing server folder, then **Import**. The server is registered where it is; files are not moved.

MMO servers (Rappelz, Hero Online, Scions of Fate) are **import-only**: picking one pre-fills its default folder (e.g. `F:\MMRappelz`) and the app discovers the start script and component executables itself.

## 4. Installing, updating, validating (Steam games)

- **Install** — downloads the server via SteamCMD. Progress and the full SteamCMD log are shown in the server's Overview tab. Cancel appears while it runs.
- **Update** — runs a SteamCMD update. If the server is running it is shut down first, then started again afterwards. The button turns amber and reads **Update available** after an update check finds a newer build.
- **Validate files** — Overview tab; runs a SteamCMD `validate` pass if files look broken.

## 5. Starting and stopping

On the card (and on the server page header):

- **Start** — boots the server.
- **Shutdown** — saves the world via RCON and exits cleanly (MMO: asks every component to close in reverse start order, then force-stops survivors).
- **Restart** — shutdown + start.
- **Stop** — force-kills the process immediately (unsaved progress is lost). Destructive actions ask for confirmation first.

**Force running** (Overview tab switch) automatically restarts the server if it crashes.

## 6. Renaming, removing, deleting

All in the Overview tab:

- **Rename** — changes the server name *and* its folder name. For MMO servers it is display-only (the folder stays put).
- **Remove from list** — forgets the server but keeps every file on disk.
- **Delete server** — permanently deletes the server folder. Asked twice; this cannot be undone.

## 7. Tabs reference

The vertical tab list on a server's page depends on the game:

| Tab | What it does |
|---|---|
| **Overview** | Session info, ports, build IDs, rename, auto-restart, open folders, validate, remove/delete, SteamCMD log |
| **Command Line** | Extra command-line arguments (key/value rows or raw string), map, RCON enabled |
| **Game Settings** | Friendly editor for the main settings (ARK: GameUserSettings-based; Palworld: the ~125 PalWorldSettings options) |
| **Advanced Config** | Searchable key/value editor over `GameUserSettings.ini` and `Game.ini` with a built-in catalog of known ARK keys (rates, per-level stat multipliers with labelled indices, breeding, PvP/PvE…) — including keys not yet present in the file |
| **Config Files** | Direct raw editing of the config files |
| **Mods** (ARK only) | Mod IDs (comma separated or **Add by ID**), **Browse CurseForge** name search, names auto-resolved from IDs, **Save mods** |
| **Players & RCON** | Live player list with **Kick / Ban / Copy ID**, broadcast to all, and a free-form RCON console (**Run**) |
| **Admins & Lists** | Admins, whitelists and ban lists |
| **Backups** | Backup folder override, keep-newest / auto-remove settings, run a backup now, **Restore** from a backup, **Open folder** |
| **Performance** | Per-server and host CPU/RAM graphs (5m / 15m / 1h history) and CPU affinity control (applied on start, and when adopting an already-running server) |
| **Ports & A2S** | Port manager showing what is bound and conflict detection (via netstat), **Find free TCP/UDP port**, and an A2S query tool — **Query this server** or any **Remote address** for live server info and player list |
| **Scheduler** | Scheduled **Actions** (start / stop / restart / shutdown / update / backup) by interval or day/time, each with an Enable switch, **Run now**, edit and delete |
| **Discord / Beta / Firewall** | Discord webhook (with test), Steam beta branch (name + password), Windows firewall rules (**Open ports / Remove rules** — needs admin rights) |
| **Logs** | The server's own log output |
| **Events** | Recent server events (started, stopped, crashes, backups…) |

Palworld servers drop the **Mods** tab; MMO script servers only show **Overview**, **Launch Config**, **Scheduler** and **Events**.

## 8. MMO script servers (Rappelz, Hero Online, Scions of Fate)

These are managed as scripts, not Steam installs:

- **Import** the folder; the app finds the start script (`.bat`) and the component executables and their ports. Hero Online has no start script, so its exe is launched directly.
- **Launch Config** tab edits the working folder, start command, optional stop command, and lists the discovered components.
- Status is derived from the component processes (matched by exe path) and listening ports — a card shows `2 / 3` components up, for example.
- **Stop** closes components gracefully in reverse start order, then force-kills any that remain.
- Rename is display-only and **Delete server** is unavailable — use **Remove from list** (files are never touched).
- Scheduler works with start / stop / restart / shutdown.

## 9. Where everything lives

```
MasterMindHost.exe
MastermindHost-data\
  store.json          <- all server definitions and settings
  <server folders>\   <- newly created servers (unless overridden in Settings)
  backups\
  steamcmd\
```

Set the `MASTERMINDHOST_DATA` environment variable to point the app at a different data folder (useful for testing).

## 10. Quick troubleshooting

- **Server shows a conflict or will not bind** — open **Ports & A2S**, **Rescan**, and use **Find free UDP/TCP port** to pick a free one.
- **Files seem corrupt or the build ID is odd** — Overview → **Validate files**.
- **Update check finds nothing but you know there is an update** — top-bar refresh (**Check for game updates**) refreshes every installed server; also check the Steam beta branch on **Discord / Beta / Firewall**.
- **Crashed server** — with **Force running** on it comes back by itself; otherwise just press **Start**.
- **Deployed exe will not update** — close the running app before copying a newly built exe over it (Windows locks running exes).
