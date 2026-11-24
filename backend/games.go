package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ConfigFileDef struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	RelPath     string `json:"relPath"`
	Description string `json:"description"`
}

type GameDef struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	AppID       string          `json:"appId"`
	Exe         string          `json:"exe"`
	ProcessName string          `json:"processName"`
	SavesRel    string          `json:"savesRel"`
	LogRel      string          `json:"logRel"`
	Header      string          `json:"header"`
	ConfigFiles []ConfigFileDef `json:"configFiles"`
	Enabled     bool            `json:"enabled"`
	Kind        string          `json:"kind"`
	DefaultPort int             `json:"defaultPort"`
	DefaultRcon int             `json:"defaultRcon"`
	Maps        []string        `json:"maps"`
	Group       string          `json:"group"`
	DefaultDir  string          `json:"defaultDir"`
}

const arksaCfgDir = `ShooterGame\Saved\Config\WindowsServer`
const palCfgDir = `Pal\Saved\Config\WindowsServer`
const palSection = "/Script/Pal.PalGameWorldSettings"

var arkseMaps = []string{"TheIsland", "TheCenter", "ScorchedEarth_P", "Ragnarok", "Aberration_P", "Extinction", "Valguero_P", "Genesis", "CrystalIsles", "Gen2", "LostIsland", "Fjordur"}
var arksaMaps = []string{"TheIsland_WP", "ScorchedEarth_WP", "TheCenter_WP", "Aberration_WP", "Extinction_WP", "Astraeos_WP", "Ragnarok_WP", "Valguero_WP", "Genesis_WP", "LostColony_WP", "BobsMissions_WP"}

var Games = map[string]GameDef{
	"arksa": {
		Group:       "ASA",
		ID:          "arksa",
		Name:        "ARK: Survival Ascended",
		AppID:       "2430930",
		Exe:         `ShooterGame\Binaries\Win64\ArkAscendedServer.exe`,
		ProcessName: "ArkAscendedServer.exe",
		SavesRel:    `ShooterGame\Saved\SavedArks`,
		LogRel:      `ShooterGame\Saved\Logs\ShooterGame.log`,
		Header:      "https://steamcdn-a.akamaihd.net/steam/apps/2399830/header.jpg",
		Enabled:     true,
		Kind:        "ark",
		DefaultPort: 7777,
		DefaultRcon: 27020,
		Maps:        arksaMaps,
		ConfigFiles: []ConfigFileDef{
			{"cmd", "Command Line", "", "Launch arguments: Map?Option=Value -Flag"},
			{"gusIni", "GameUserSettings.ini", arksaCfgDir + `\GameUserSettings.ini`, "Main server settings"},
			{"gameIni", "Game.ini", arksaCfgDir + `\Game.ini`, "Advanced gameplay settings"},
			{"adminList", "AllowedCheaterAccountIDs.txt", `ShooterGame\Saved\AllowedCheaterAccountIDs.txt`, "Admin player IDs (one per line)"},
			{"exclusiveList", "PlayersExclusiveJoinList.txt", `ShooterGame\Binaries\Win64\PlayersExclusiveJoinList.txt`, "Whitelist used with -exclusivejoin"},
			{"exclusiveListNoCheckList", "PlayersJoinNoCheckList.txt", `ShooterGame\Binaries\Win64\PlayersJoinNoCheckList.txt`, "Players that skip the queue / join checks"},
			{"banList", "BanList.txt", `ShooterGame\Saved\BanList.txt`, "Banned player IDs"},
		},
	},
	"arkse": {
		Group:       "ASE",
		ID:          "arkse",
		Name:        "ARK: Survival Evolved",
		AppID:       "376030",
		Exe:         `ShooterGame\Binaries\Win64\ShooterGameServer.exe`,
		ProcessName: "ShooterGameServer.exe",
		SavesRel:    `ShooterGame\Saved\SavedArks`,
		LogRel:      `ShooterGame\Saved\Logs\ShooterGame.log`,
		Header:      "https://steamcdn-a.akamaihd.net/steam/apps/346110/header.jpg",
		Enabled:     true,
		Kind:        "ark",
		DefaultPort: 7777,
		DefaultRcon: 27020,
		Maps:        arkseMaps,
		ConfigFiles: []ConfigFileDef{
			{"cmd", "Command Line", "", "Launch arguments: Map?listen?Option=Value -Flag"},
			{"gusIni", "GameUserSettings.ini", arksaCfgDir + `\GameUserSettings.ini`, "Main server settings"},
			{"gameIni", "Game.ini", arksaCfgDir + `\Game.ini`, "Advanced gameplay settings"},
			{"adminList", "AllowedCheaterSteamIDs.txt", `ShooterGame\Saved\AllowedCheaterSteamIDs.txt`, "Admin SteamID64s (one per line)"},
			{"exclusiveList", "PlayersExclusiveJoinList.txt", `ShooterGame\Binaries\Win64\PlayersExclusiveJoinList.txt`, "Whitelist used with -exclusivejoin"},
			{"exclusiveListNoCheckList", "PlayersJoinNoCheckList.txt", `ShooterGame\Binaries\Win64\PlayersJoinNoCheckList.txt`, "Players that skip join checks"},
			{"banList", "BanList.txt", `ShooterGame\Saved\BanList.txt`, "Banned SteamID64s"},
		},
	},
	"palworld": {
		Group:       "Palworld",
		ID:          "palworld",
		Name:        "Palworld",
		AppID:       "2394010",
		Exe:         `Pal\Binaries\Win64\PalServer-Win64-Shipping-Cmd.exe`,
		ProcessName: "PalServer-Win64-Shipping-Cmd.exe",
		SavesRel:    `Pal\Saved\SaveGames`,
		LogRel:      ``,
		Header:      "https://steamcdn-a.akamaihd.net/steam/apps/1623730/header.jpg",
		Enabled:     true,
		Kind:        "palworld",
		DefaultPort: 8211,
		DefaultRcon: 25575,
		ConfigFiles: []ConfigFileDef{
			{"cmd", "Command Line", "", "Launch arguments: Pal -port=8211 -players=32 ..."},
			{"gusIni", "PalWorldSettings.ini", palCfgDir + `\PalWorldSettings.ini`, "Main server settings (OptionSettings)"},
			{"engineIni", "Engine.ini", palCfgDir + `\Engine.ini`, "Engine tuning"},
			{"banList", "banlist.txt", `Pal\Saved\SaveGames\banlist.txt`, "Banned player IDs"},
		},
	},
	"rappelz": {
		ID:         "rappelz",
		Name:       "Rappelz",
		Group:      "MMOs",
		Kind:       "script",
		Enabled:    true,
		DefaultDir: `F:\MMRappelz`,
	},
	"hero": {
		ID:         "hero",
		Name:       "Hero Online",
		Group:      "MMOs",
		Kind:       "script",
		Enabled:    true,
		DefaultDir: `F:\DragonzOnline\HeroRebirthFork`,
	},
	"scions": {
		ID:         "scions",
		Name:       "Scions of Fate",
		Group:      "MMOs",
		Kind:       "script",
		Enabled:    true,
		DefaultDir: `F:\HollowFateOnline`,
	},
}

var gameOrder = []string{"arksa", "arkse", "palworld", "rappelz", "hero", "scions"}

func isScript(sv Server) bool { return Games[sv.Game].Kind == "script" }

func gameDef(id string) (GameDef, error) {
	g, ok := Games[id]
	if !ok || !g.Enabled {
		return GameDef{}, fmt.Errorf("unsupported game %q", id)
	}
	return g, nil
}

func (g GameDef) configFile(key string) (ConfigFileDef, error) {
	for _, c := range g.ConfigFiles {
		if c.Key == key {
			return c, nil
		}
	}
	return ConfigFileDef{}, fmt.Errorf("unknown config file %q", key)
}

var invalidNameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
var reservedNames = map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true}

func sanitizeFolderName(name string) string {
	n := invalidNameChars.ReplaceAllString(strings.TrimSpace(name), "_")
	n = strings.TrimRight(n, ". ")
	if n == "" {
		n = "GameServer"
	}
	if reservedNames[strings.ToUpper(n)] {
		n = "_" + n
	}
	if len(n) > 100 {
		n = strings.TrimRight(n[:100], ". ")
	}
	return n
}

func uniqueDir(root, name, except string) string {
	base := sanitizeFolderName(name)
	candidate := filepath.Join(root, base)
	for i := 2; ; i++ {
		if strings.EqualFold(candidate, except) {
			return candidate
		}
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
		candidate = filepath.Join(root, fmt.Sprintf("%s (%d)", base, i))
	}
}
