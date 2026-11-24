package backend

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type AdvancedIniEntry struct {
	File     string `json:"file"`
	FileName string `json:"fileName"`
	Section  string `json:"section"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Set      bool   `json:"set"`
	Group    string `json:"group"`
	Desc     string `json:"desc"`
	Kind     string `json:"kind"`
}

type AdvancedIniChange struct {
	File    string `json:"file"`
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Remove  bool   `json:"remove"`
}

const shootGameMode = "/Script/ShooterGame.ShooterGameMode"

type setMeta struct {
	group string
	desc  string
	kind  string
}

var statLabels = [...]string{"Health", "Stamina", "Torpidity", "Oxygen", "Food", "Water", "Temperature", "Weight", "Melee Damage", "Movement Speed", "Fortitude", "Crafting Speed"}

var arrayFamilies = []struct {
	file, section, group, prefix, desc string
	expand                             bool
}{
	{"gameIni", shootGameMode, "Player & Dino Stats", "PerLevelStatsMultiplier_Player", "Player stat points gained per level-up multiplier (1 = official)", true},
	{"gameIni", shootGameMode, "Player & Dino Stats", "PerLevelStatsMultiplier_DinoWild", "Wild dino stat points gained per level-up multiplier (1 = official)", true},
	{"gameIni", shootGameMode, "Player & Dino Stats", "PerLevelStatsMultiplier_DinoTamed", "Tamed dino stat points gained per level-up multiplier (1 = official)", true},
	{"gameIni", shootGameMode, "Player & Dino Stats", "PerLevelStatsMultiplier_DinoTamed_Add", "Flat bonus added to tamed dino stat gains after multipliers", false},
	{"gameIni", shootGameMode, "Player & Dino Stats", "PerLevelStatsMultiplier_DinoTamed_Affinity", "Tamed dino stat gains from affinity (imprint / post-tame levels) multiplier", false},
	{"gameIni", shootGameMode, "Player & Dino Stats", "PerLevelStatsMultiplier_DinoTamed_Affinity_NonMult", "Affinity flat bonus variant", false},
}

var catTable = []struct {
	file, section, group, key, desc, kind string
}{
	{"gusIni", "SessionSettings", "Server", "SessionName", "Server name shown in the session browser", "text"},
	{"gusIni", "/Script/Engine.GameSession", "Server", "MaxPlayers", "Max players that can join", "number"},
	{"gusIni", "ServerSettings", "Server", "Port", "Game port (default 7777)", "number"},
	{"gusIni", "ServerSettings", "Server", "RCONEnabled", "Enable RCON remote admin", "bool"},
	{"gusIni", "ServerSettings", "Server", "RCONPort", "RCON port (default 27020)", "number"},
	{"gusIni", "ServerSettings", "Server", "RCONServerGameLogBuffer", "Lines of game log buffered for RCON", "number"},
	{"gusIni", "ServerSettings", "Server", "ServerAdminPassword", "Admin (cheat / RCON) password", "text"},
	{"gusIni", "ServerSettings", "Server", "ServerPassword", "Join password (empty = open)", "text"},
	{"gusIni", "ServerSettings", "Server", "AdminLogging", "Log all admin (cheat) usage", "bool"},
	{"gusIni", "ServerSettings", "Server", "ActiveMods", "Space-separated list of active mod IDs", "text"},
	{"gusIni", "ServerSettings", "Server", "AutoSavePeriodMinutes", "World autosave interval in minutes", "number"},
	{"gusIni", "ServerSettings", "Server", "KickIdlePlayersPeriod", "Kick idle players after this many seconds (0 = never)", "number"},
	{"gusIni", "ServerSettings", "Server", "NoTributeDownloads", "Block dino/item downloads (transfers) from other servers", "bool"},
	{"gusIni", "ServerSettings", "Server", "NPCNetworkStasisRangeScalePlayerCountStart", "Dino activation radius starts shrinking at this player count (performance)", "number"},
	{"gusIni", "ServerSettings", "Server", "NPCNetworkStasisRangeScalePlayerCountEnd", "Dino activation radius minimum player count (performance)", "number"},
	{"gusIni", "ServerSettings", "Server", "AutoDestroyOldStructuresMultiplier", "Auto-decay timer multiplier for old structures of absent players", "number"},
	{"gusIni", "ServerSettings", "Server", "AutoDestroyDecayedStructures", "Auto-destroy structures that finish decaying", "bool"},
	{"gusIni", "ServerSettings", "PvP / PvE", "ServerPVE", "PvE mode (no PvP damage). Restart required", "bool"},
	{"gusIni", "ServerSettings", "PvP / PvE", "ServerHardcore", "Hardcore mode (level reset on death)", "bool"},
	{"gusIni", "ServerSettings", "PvP / PvE", "bPvPDisableDuringBuildingDecay", "Disable PvP damage while a decay period is active", "bool"},
	{"gusIni", "ServerSettings", "PvP / PvE", "PvPZoneStructureDamageMultiplier", "Structure damage multiplier inside PvP zones (caves)", "number"},
	{"gusIni", "ServerSettings", "PvP / PvE", "bAutoPvETimer", "Auto-switch players between PvP and PvE by timer", "bool"},
	{"gusIni", "ServerSettings", "PvP / PvE", "bAutoPvEUseSystemTime", "Use real system time (not in-game time) for the auto PvP/PvE timer", "bool"},
	{"gusIni", "ServerSettings", "PvP / PvE", "AllowCaveBuildingPvE", "Allow building in caves on PvE servers", "bool"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DifficultyOffset", "Difficulty offset (0-1; 1 = max level 150 wild)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "OverrideOfficialDifficulty", "Force max wild dino level (e.g. 5 = 150)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "XPMultiplier", "XP gain multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "TamingSpeedMultiplier", "Taming speed multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "HarvestAmountMultiplier", "Harvest amount multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "HarvestHealthMultiplier", "Harvestable resource health (yield per hit) multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "ResourcesRespawnPeriodMultiplier", "Resource respawn period multiplier (lower = faster)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DinoCountMultiplier", "Wild dino spawn count multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DayCycleSpeedScale", "Full day cycle speed multiplier (lower = longer days)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DayTimeSpeedScale", "Daytime speed multiplier (lower = longer daylight)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "NightTimeSpeedScale", "Nighttime speed multiplier (lower = longer nights)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "PlayerCharacterFoodDrainMultiplier", "Player food drain multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "PlayerCharacterWaterDrainMultiplier", "Player water drain multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "PlayerCharacterStaminaDrainMultiplier", "Player stamina drain multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DinoCharacterFoodDrainMultiplier", "Dino food drain multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DinoCharacterStaminaDrainMultiplier", "Dino stamina drain multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "PlayerDamageMultiplier", "Player damage multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "PlayerResistanceMultiplier", "Player resistance multiplier (higher = takes more damage)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DinoDamageMultiplier", "Dino damage multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "DinoResistanceMultiplier", "Dino resistance multiplier (higher = takes more damage)", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "StructureDamageMultiplier", "Structure damage taken multiplier", "number"},
	{"gusIni", "ServerSettings", "Rates & Multipliers", "StructureResistanceMultiplier", "Structure resistance multiplier", "number"},
	{"gusIni", "ServerSettings", "Items & Resources", "ItemStackSizeMultiplier", "Item stack size multiplier", "number"},
	{"gusIni", "ServerSettings", "Items & Resources", "AllowHitMarkers", "Show hit markers in combat", "bool"},
	{"gusIni", "ServerSettings", "Misc", "ServerCrosshair", "Show crosshair", "bool"},
	{"gusIni", "ServerSettings", "Misc", "ShowMapPlayerLocation", "Show own position on the map", "bool"},
	{"gusIni", "ServerSettings", "Misc", "AllowThirdPersonPlayer", "Allow third-person view", "bool"},
	{"gusIni", "ServerSettings", "Misc", "AlwaysNotifyPlayerJoined", "Broadcast player join notifications", "bool"},
	{"gusIni", "ServerSettings", "Misc", "AlwaysNotifyPlayerLeft", "Broadcast player leave notifications", "bool"},
	{"gameIni", shootGameMode, "Breeding", "MatingIntervalMultiplier", "Mating interval multiplier (lower = mates again sooner)", "number"},
	{"gameIni", shootGameMode, "Breeding", "EggHatchSpeedMultiplier", "Egg hatch speed multiplier", "number"},
	{"gameIni", shootGameMode, "Breeding", "BabyMatureSpeedMultiplier", "Baby maturation speed multiplier", "number"},
	{"gameIni", shootGameMode, "Breeding", "BabyFoodConsumptionSpeedMultiplier", "Baby food consumption speed multiplier", "number"},
	{"gameIni", shootGameMode, "Breeding", "bDisableBabyFoodConsumption", "Babies consume no food (still need claiming)", "bool"},
	{"gameIni", shootGameMode, "Breeding", "BabyCuddleIntervalMultiplier", "Time between cuddles multiplier (lower = imprint faster)", "number"},
	{"gameIni", shootGameMode, "Breeding", "BabyCuddleGracePeriodMultiplier", "Cuddle grace period multiplier", "number"},
	{"gameIni", shootGameMode, "Breeding", "BabyCuddleLoseImprintQualitySpeedMultiplier", "Imprint quality decay speed multiplier", "number"},
	{"gameIni", shootGameMode, "Breeding", "BabyImprintAmountMultiplier", "Imprint amount per cuddle multiplier", "number"},
	{"gameIni", shootGameMode, "Breeding", "BabyImprintingStatScaleMultiplier", "How much of the imprint bonus is applied to stats", "number"},
	{"gameIni", shootGameMode, "Breeding", "LayEggIntervalMultiplier", "Egg laying interval multiplier", "number"},
	{"gameIni", shootGameMode, "World", "CropGrowthSpeedMultiplier", "Crop plot growth speed multiplier", "number"},
	{"gameIni", shootGameMode, "World", "SpoilingTimeMultiplier", "Item spoil time multiplier (higher = slower spoil)", "number"},
	{"gameIni", shootGameMode, "World", "GlobalItemDecayTimeMultiplier", "Item on-ground decay time multiplier", "number"},
	{"gameIni", shootGameMode, "Misc", "bAllowUnlimitedRespecs", "Allow unlimited mind-wipe respecs without cooldown", "bool"},
	{"gameIni", shootGameMode, "Misc", "bAllowFlyerSpeedLeveling", "Allow leveling flying speed on flyers", "bool"},
	{"gameIni", shootGameMode, "Misc", "bDisableStructurePlacementCollision", "Allow structures to ignore placement collision", "bool"},
	{"gameIni", shootGameMode, "Misc", "MaxTribesPerAlliance", "Max tribes in an alliance", "number"},
}

type catEntry struct {
	file, section, key, kind string
	meta                     setMeta
}

var catEntries []catEntry
var catIndex = map[string]setMeta{}

func init() {
	for _, r := range catTable {
		m := setMeta{group: r.group, desc: r.desc, kind: r.kind}
		catEntries = append(catEntries, catEntry{file: r.file, section: r.section, key: r.key, kind: r.kind, meta: m})
		catIndex[r.file+"|"+strings.ToLower(r.key)] = m
	}
	for _, f := range arrayFamilies {
		if f.expand {
			for i, l := range statLabels {
				k := fmt.Sprintf("%s[%d]", f.prefix, i)
				m := setMeta{group: f.group, desc: l + " - " + f.desc, kind: "number"}
				catEntries = append(catEntries, catEntry{file: f.file, section: f.section, key: k, kind: "number", meta: m})
				catIndex[f.file+"|"+strings.ToLower(k)] = m
			}
		} else {
			m := setMeta{group: f.group, desc: f.desc + " (append [0-11] to the key)", kind: "number"}
			catEntries = append(catEntries, catEntry{file: f.file, section: f.section, key: f.prefix, kind: "number", meta: m})
			catIndex[f.file+"|"+strings.ToLower(f.prefix)] = m
		}
	}
}

func metaFor(file, key string) (setMeta, bool) {
	if m, ok := catIndex[file+"|"+strings.ToLower(key)]; ok {
		return m, true
	}
	lk := strings.ToLower(key)
	for _, f := range arrayFamilies {
		if file != f.file || !strings.HasPrefix(lk, strings.ToLower(f.prefix)+"[") {
			continue
		}
		idx := strings.TrimSuffix(lk[len(f.prefix)+1:], "]")
		n, err := strconv.Atoi(idx)
		if err != nil || n < 0 || n >= len(statLabels) {
			continue
		}
		return setMeta{group: f.group, desc: fmt.Sprintf("%s (index %d: %s)", f.desc, n, statLabels[n]), kind: "number"}, true
	}
	return setMeta{}, false
}

func inferKind(v string) string {
	t := strings.TrimSpace(v)
	if strings.EqualFold(t, "true") || strings.EqualFold(t, "false") {
		return "bool"
	}
	if _, err := strconv.ParseFloat(t, 64); err == nil && t != "" {
		return "number"
	}
	if strings.HasPrefix(t, "(") {
		return "group"
	}
	return "text"
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func natLess(a, b string) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if isDigit(a[i]) && isDigit(b[j]) {
			ia, jb := i, j
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			na, _ := strconv.Atoi(a[ia:i])
			nb, _ := strconv.Atoi(b[jb:j])
			if na != nb {
				return na < nb
			}
			continue
		}
		if a[i] != b[j] {
			return a[i] < b[j]
		}
		i++
		j++
	}
	return len(a)-i < len(b)-j
}

func advGroupLabel(e AdvancedIniEntry) string {
	if e.Group != "" {
		return strings.ToLower(e.Group)
	}
	return strings.ToLower(e.Section)
}

func sortAdvancedEntries(entries []AdvancedIniEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		gi, gj := advGroupLabel(entries[i]), advGroupLabel(entries[j])
		if gi != gj {
			return gi < gj
		}
		if entries[i].File != entries[j].File {
			return entries[i].File < entries[j].File
		}
		return natLess(strings.ToLower(entries[i].Key), strings.ToLower(entries[j].Key))
	})
}

func (h *Host) AdvancedIniList(id string) ([]AdvancedIniEntry, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	g, ok := Games[sv.Game]
	if !ok {
		return nil, fmt.Errorf("unsupported game %q", sv.Game)
	}
	var out []AdvancedIniEntry
	seen := map[string]bool{}
	for _, cf := range g.ConfigFiles {
		if !strings.HasSuffix(strings.ToLower(cf.RelPath), ".ini") {
			continue
		}
		name := filepath.Base(cf.RelPath)
		parsed := ParseIni(h.readConfigSilently(sv, cf.Key))
		sections := make([]string, 0, len(parsed))
		for s := range parsed {
			sections = append(sections, s)
		}
		sort.Strings(sections)
		for _, s := range sections {
			keys := make([]string, 0, len(parsed[s]))
			for k := range parsed[s] {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool { return natLess(strings.ToLower(keys[i]), strings.ToLower(keys[j])) })
			for _, k := range keys {
				v := parsed[s][k]
				meta, _ := metaFor(cf.Key, k)
				kind := meta.kind
				if kind == "" {
					kind = inferKind(v)
				}
				out = append(out, AdvancedIniEntry{File: cf.Key, FileName: name, Section: s, Key: k, Value: v, Set: true, Group: meta.group, Desc: meta.desc, Kind: kind})
				seen[cf.Key+"|"+s+"|"+strings.ToLower(k)] = true
			}
		}
		for _, ce := range catEntries {
			if ce.file != cf.Key || seen[cf.Key+"|"+ce.section+"|"+strings.ToLower(ce.key)] {
				continue
			}
			out = append(out, AdvancedIniEntry{File: cf.Key, FileName: name, Section: ce.section, Key: ce.key, Set: false, Group: ce.meta.group, Desc: ce.meta.desc, Kind: ce.kind})
		}
	}
	sortAdvancedEntries(out)
	return out, nil
}

func (h *Host) AdvancedIniApply(id string, changes []AdvancedIniChange) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	g, ok := Games[sv.Game]
	if !ok {
		return fmt.Errorf("unsupported game %q", sv.Game)
	}
	byFile := map[string][]IniChange{}
	var order []string
	for _, c := range changes {
		c.Key = strings.TrimSpace(c.Key)
		if c.Key == "" {
			continue
		}
		if _, err := g.configFile(c.File); err != nil {
			return fmt.Errorf("unknown config file %q", c.File)
		}
		if _, ok := byFile[c.File]; !ok {
			order = append(order, c.File)
		}
		byFile[c.File] = append(byFile[c.File], IniChange{Section: c.Section, Key: c.Key, Value: c.Value, Remove: c.Remove})
	}
	for _, f := range order {
		content := ApplyIniChanges(h.readConfigSilently(sv, f), byFile[f])
		if err := h.WriteConfig(id, f, content); err != nil {
			return err
		}
	}
	return nil
}
