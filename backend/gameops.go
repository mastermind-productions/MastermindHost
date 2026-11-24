package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type gameRuntime struct {
	Port        int
	QueryPort   int
	RconPort    int
	RconPass    string
	RconEnabled bool
	MaxPlayers  int
	Session     string
	Map         string
	HasPassword bool
	ClusterID   string
}

func isPal(sv Server) bool { return Games[sv.Game].Kind == "palworld" }

func parsePalOptions(s string) []KV {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.TrimPrefix(s, "("), ")")
	var out []KV
	var cur strings.Builder
	inQ := false
	flush := func() {
		t := strings.TrimSpace(cur.String())
		cur.Reset()
		if t != "" {
			out = append(out, splitKV(t))
		}
	}
	for _, r := range s {
		switch {
		case r == '"':
			inQ = !inQ
			cur.WriteRune(r)
		case r == ',' && !inQ:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

func palOptionsString(opts []KV) string {
	parts := make([]string, 0, len(opts))
	for _, o := range opts {
		if o.HasValue {
			parts = append(parts, o.Key+"="+o.Value)
		} else {
			parts = append(parts, o.Key)
		}
	}
	return "(" + strings.Join(parts, ",") + ")"
}

func palGet(opts []KV, key string) string {
	if i := findKV(opts, key); i >= 0 {
		return strings.Trim(opts[i].Value, `"`)
	}
	return ""
}

func palSet(opts []KV, key, value string) []KV {
	kv := KV{Key: key, Value: value, HasValue: true}
	if i := findKV(opts, key); i >= 0 {
		opts[i] = kv
		return opts
	}
	return append(opts, kv)
}

func palQuote(v string) string { return `"` + strings.ReplaceAll(v, `"`, "") + `"` }

func (h *Host) palOptions(sv Server) []KV {
	ini := ParseIni(h.readConfigSilently(sv, "gusIni"))
	return parsePalOptions(ini[palSection]["OptionSettings"])
}

func flagInt(c ArkCmd, key string, def int) int {
	if v, ok := c.Flag(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func (h *Host) runtime(sv Server) gameRuntime {
	g := Games[sv.Game]
	cmd := ParseArkCmd(sv.Cmd)
	if g.Kind == "palworld" {
		o := h.palOptions(sv)
		r := gameRuntime{
			Port:        flagInt(cmd, "port", g.DefaultPort),
			RconPort:    g.DefaultRcon,
			RconPass:    palGet(o, "AdminPassword"),
			RconEnabled: strings.EqualFold(palGet(o, "RCONEnabled"), "true"),
			MaxPlayers:  32,
			Session:     palGet(o, "ServerName"),
			HasPassword: palGet(o, "ServerPassword") != "",
		}
		if v := palGet(o, "RCONPort"); v != "" {
			fmt.Sscan(v, &r.RconPort)
		}
		if v := palGet(o, "ServerPlayerMaxNum"); v != "" {
			fmt.Sscan(v, &r.MaxPlayers)
		}
		r.MaxPlayers = flagInt(cmd, "players", r.MaxPlayers)
		return r
	}
	gus := ParseIni(h.readConfigSilently(sv, "gusIni"))
	r := gameRuntime{
		Port:      cmd.IntOpt("Port", g.DefaultPort),
		QueryPort: cmd.IntOpt("QueryPort", 0),
		RconPort:  rconPort(cmd, gus),
		Map:       cmd.Map,
		Session:   strings.Trim(gus["SessionSettings"]["SessionName"], `"`),
	}
	if v, ok := cmd.Opt("SessionName"); ok {
		r.Session = v
	}
	r.RconEnabled = strings.EqualFold(firstNonEmpty(optOr(cmd, "RCONEnabled"), gus["ServerSettings"]["RCONEnabled"]), "true")
	pass, ok := cmd.Opt("ServerAdminPassword")
	if !ok || pass == "" {
		pass = gus["ServerSettings"]["ServerAdminPassword"]
	}
	r.RconPass = pass
	r.MaxPlayers = cmd.IntOpt("MaxPlayers", 70)
	if v, ok := cmd.Flag("WinLiveMaxPlayers"); ok {
		fmt.Sscan(v, &r.MaxPlayers)
	}
	sp, _ := cmd.Opt("ServerPassword")
	r.HasPassword = firstNonEmpty(sp, gus["ServerSettings"]["ServerPassword"]) != ""
	r.ClusterID, _ = cmd.Flag("clusterid")
	return r
}

func (r gameRuntime) ports() []int {
	p := []int{r.Port, r.Port + 1, r.RconPort}
	if r.QueryPort > 0 {
		p = append(p, r.QueryPort)
	}
	return p
}

func (r gameRuntime) udpPorts(kind string) []int {
	if kind == "palworld" {
		return []int{r.Port}
	}
	p := []int{r.Port, r.Port + 1}
	if r.QueryPort > 0 {
		p = append(p, r.QueryPort)
	}
	return p
}

func saveCmd(sv Server) string {
	if isPal(sv) {
		return "Save"
	}
	return "SaveWorld"
}

func playersCmd(sv Server) string {
	if isPal(sv) {
		return "ShowPlayers"
	}
	return "ListPlayers"
}

func broadcastCmd(sv Server, msg string) string {
	if isPal(sv) {
		return "Broadcast " + strings.ReplaceAll(msg, " ", "_")
	}
	return "ServerChat " + msg
}

func unbanCmd(sv Server) string {
	if isPal(sv) {
		return "UnBanPlayer "
	}
	return "UnbanPlayer "
}

func parsePalPlayers(out string) []Player {
	list := []Player{}
	for i, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if i == 0 && strings.HasPrefix(strings.ToLower(line), "name,") || line == "" {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 3 {
			continue
		}
		list = append(list, Player{Index: f[1], Name: strings.Join(f[:len(f)-2], ","), ID: f[len(f)-1]})
	}
	return list
}

func parsePlayersFor(sv Server, out string) []Player {
	if isPal(sv) {
		return parsePalPlayers(out)
	}
	return parsePlayers(out)
}

func defaultCmd(g GameDef, port, query, rcon int, pass string) string {
	switch g.ID {
	case "arkse":
		return fmt.Sprintf("TheIsland?listen?Port=%d?QueryPort=%d?RCONPort=%d?RCONEnabled=True?ServerAdminPassword=%s?MaxPlayers=70 -servergamelog -NoBattlEye",
			port, query, rcon, pass)
	case "palworld":
		return fmt.Sprintf("Pal -port=%d -players=32 -useperfthreads -NoAsyncLoadingThread -UseMultithreadForDS", port)
	}
	return fmt.Sprintf("TheIsland_WP?Port=%d?RCONPort=%d?RCONEnabled=True?ServerAdminPassword=%s -WinLiveMaxPlayers=70 -servergamelog",
		port, rcon, pass)
}

func initialConfig(g GameDef, name string, rcon int, pass string) string {
	if g.Kind == "palworld" {
		opts := []KV{
			{Key: "ServerName", Value: palQuote(name), HasValue: true},
			{Key: "ServerPlayerMaxNum", Value: "32", HasValue: true},
			{Key: "AdminPassword", Value: palQuote(pass), HasValue: true},
			{Key: "RCONEnabled", Value: "True", HasValue: true},
			{Key: "RCONPort", Value: fmt.Sprint(rcon), HasValue: true},
		}
		return ApplyIniChanges("", []IniChange{{Section: palSection, Key: "OptionSettings", Value: palOptionsString(opts)}})
	}
	return ApplyIniChanges("", []IniChange{
		{Section: "ServerSettings", Key: "RCONEnabled", Value: "True"},
		{Section: "ServerSettings", Key: "RCONPort", Value: fmt.Sprint(rcon)},
		{Section: "SessionSettings", Key: "SessionName", Value: name},
	})
}

func (h *Host) GetPalSettings(id string) ([]KV, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !isPal(sv) {
		return nil, fmt.Errorf("not a Palworld server")
	}
	o := []KV{}
	if b, err := os.ReadFile(filepath.Join(sv.Dir, "DefaultPalWorldSettings.ini")); err == nil {
		o = parsePalOptions(ParseIni(string(b))[palSection]["OptionSettings"])
	}
	for _, kv := range h.palOptions(sv) {
		o = palSet(o, kv.Key, kv.Value)
	}
	return o, nil
}

func (h *Host) SetPalSettings(id string, opts []KV) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if !isPal(sv) {
		return fmt.Errorf("not a Palworld server")
	}
	content := h.readConfigSilently(sv, "gusIni")
	content = ApplyIniChanges(content, []IniChange{{Section: palSection, Key: "OptionSettings", Value: palOptionsString(opts)}})
	if err := h.WriteConfig(id, "gusIni", content); err != nil {
		return err
	}
	h.logAction(id, "update settings", "success", "PalWorldSettings.ini")
	h.emit(id)
	return nil
}
