package backend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	StatusCreated      = "created"
	StatusInstalling   = "installing"
	StatusUpdating     = "updating"
	StatusOffline      = "offline"
	StatusStarting     = "starting"
	StatusRunning      = "running"
	StatusOnline       = "online"
	StatusStopping     = "stopping"
	StatusShuttingDown = "shutting down"
	StatusRestarting   = "restarting"
	StatusNotFound     = "not found"
)

type rtState struct {
	status     string
	pid        uint32
	startTime  time.Time
	cpu        float64
	ram        uint64
	lastCPU    time.Duration
	lastSample time.Time
	lastRcon   time.Time
	winPid     uint32
	players    int
	progress   float64
	installLog []string
	userStop   bool
	launching  bool
	compUp     int
	busy       bool
	cancel     context.CancelFunc
}

type Host struct {
	ctx      context.Context
	store    *Store
	mu       sync.Mutex
	rt       map[string]*rtState
	steamMu  sync.Mutex
	latest   map[string]string
	latestMu sync.Mutex
	metrics  metricsStore
}

type ServerInfo struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Game            string  `json:"game"`
	GameName        string  `json:"gameName"`
	Dir             string  `json:"dir"`
	CreatedAt       string  `json:"createdAt"`
	Status          string  `json:"status"`
	Installed       bool    `json:"installed"`
	Map             string  `json:"map"`
	SessionName     string  `json:"sessionName"`
	Port            int     `json:"port"`
	RconPort        int     `json:"rconPort"`
	RconEnabled     bool    `json:"rconEnabled"`
	MaxPlayers      int     `json:"maxPlayers"`
	PlayerCount     int     `json:"playerCount"`
	BuildID         string  `json:"buildId"`
	LatestBuildID   string  `json:"latestBuildId"`
	UpdateAvailable bool    `json:"updateAvailable"`
	CPU             float64 `json:"cpu"`
	RAM             uint64  `json:"ram"`
	StartTime       int64   `json:"startTime"`
	ForceRunning    bool    `json:"forceRunning"`
	Progress        float64 `json:"progress"`
	ModCount        int     `json:"modCount"`
	ClusterID       string  `json:"clusterId"`
	HasPassword     bool    `json:"hasPassword"`
	BetaBranch      string  `json:"betaBranch"`
	Header          string  `json:"header"`
	Group           string  `json:"group"`
	Script          bool    `json:"script"`
	CompTotal       int     `json:"compTotal"`
	CompUp          int     `json:"compUp"`
}

func NewHost() (*Host, error) {
	st, err := NewStore()
	if err != nil {
		return nil, err
	}
	return &Host{store: st, rt: map[string]*rtState{}, latest: map[string]string{}}, nil
}

func (h *Host) Startup(ctx context.Context) {
	h.ctx = ctx
	for _, sv := range h.store.List() {
		h.state(sv.ID)
	}
	go h.monitorLoop()
	go h.schedulerLoop()
	go h.latestBuildLoop()
	h.mainWindowRestore(ctx)
	go h.windowSaveLoop()
}

func (h *Host) OnShutdown(ctx context.Context) {}

func (h *Host) state(id string) *rtState {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.rt[id]
	if !ok {
		s = &rtState{status: ""}
		h.rt[id] = s
	}
	return s
}

func (h *Host) setStatus(id, status string) {
	s := h.state(id)
	h.mu.Lock()
	changed := s.status != status
	s.status = status
	h.mu.Unlock()
	if changed {
		h.emit(id)
		go h.discordNotify(id, status)
	}
}

func (h *Host) getStatus(id string) string {
	s := h.state(id)
	h.mu.Lock()
	defer h.mu.Unlock()
	return s.status
}

func (h *Host) emit(id string) {
	if h.ctx != nil {
		wruntime.EventsEmit(h.ctx, "gs-status-update", id)
	}
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func randomPassword(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func (h *Host) exePath(sv Server) string {
	g := Games[sv.Game]
	return filepath.Join(sv.Dir, g.Exe)
}

func (h *Host) installed(sv Server) bool {
	if isScript(sv) {
		return h.scriptInstalled(sv)
	}
	_, err := os.Stat(h.exePath(sv))
	return err == nil
}

func (h *Host) info(sv Server) ServerInfo {
	if isScript(sv) {
		return h.scriptInfo(sv)
	}
	g := Games[sv.Game]
	s := h.state(sv.ID)
	h.mu.Lock()
	status, cpu, ram, start, players, prog := s.status, s.cpu, s.ram, s.startTime, s.players, s.progress
	h.mu.Unlock()
	if status == "" {
		status = h.idleStatus(sv)
	}
	cmd := ParseArkCmd(sv.Cmd)
	r := h.runtime(sv)
	build := readBuildID(sv.Dir, g.AppID)
	latest := h.latestBuild(g.AppID, sv.BetaBranch)
	inf := ServerInfo{
		ID: sv.ID, Name: sv.Name, Game: sv.Game, GameName: g.Name, Dir: sv.Dir,
		CreatedAt: sv.CreatedAt.Format(time.RFC3339), Status: status, Installed: h.installed(sv),
		Map: r.Map, SessionName: r.Session, Port: r.Port, RconPort: r.RconPort,
		RconEnabled: r.RconEnabled, MaxPlayers: r.MaxPlayers, PlayerCount: players, BuildID: build,
		LatestBuildID: latest, UpdateAvailable: build != "" && latest != "" && build != latest,
		CPU: cpu, RAM: ram, ForceRunning: sv.ForceRunning, Progress: prog, ModCount: len(cmd.Mods()),
		ClusterID: r.ClusterID, HasPassword: r.HasPassword, BetaBranch: sv.BetaBranch, Header: g.Header, Group: g.Group,
	}
	if !start.IsZero() {
		inf.StartTime = start.UnixMilli()
	}
	return inf
}

func optOr(c ArkCmd, k string) string { v, _ := c.Opt(k); return v }

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func rconPort(cmd ArkCmd, gus map[string]map[string]string) int {
	p := 27020
	if v := gus["ServerSettings"]["RCONPort"]; v != "" {
		fmt.Sscan(v, &p)
	}
	return cmd.IntOpt("RCONPort", p)
}

func (h *Host) idleStatus(sv Server) string {
	if _, err := os.Stat(sv.Dir); err != nil {
		return StatusNotFound
	}
	if !h.installed(sv) {
		return StatusCreated
	}
	return StatusOffline
}

func (h *Host) List() []ServerInfo {
	list := h.store.List()
	out := make([]ServerInfo, 0, len(list))
	for _, sv := range list {
		out = append(out, h.info(sv))
	}
	return out
}

func (h *Host) Get(id string) (ServerInfo, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return ServerInfo{}, err
	}
	return h.info(sv), nil
}

func (h *Host) GetGames() []GameDef {
	var out []GameDef
	for _, id := range gameOrder {
		if g, ok := Games[id]; ok && g.Enabled {
			out = append(out, g)
		}
	}
	return out
}

func (h *Host) ReorderServers(ids []string) error {
	return h.store.Reorder(ids)
}

func (h *Host) HostStats() map[string]any {
	total, avail := hostMemory()
	dt, df := diskUsage(h.store.Settings().ServersRoot)
	return map[string]any{
		"cpu": hostCPU(), "ramTotal": total, "ramUsed": total - avail,
		"diskTotal": dt, "diskUsed": dt - df, "cores": runtime.NumCPU(),
		"serversRoot": h.store.Settings().ServersRoot,
	}
}

var cpuMu sync.Mutex
var lastIdle, lastTotal uint64
var lastHostCPU float64

func hostCPU() float64 {
	cpuMu.Lock()
	defer cpuMu.Unlock()
	idle, total := systemTimes()
	if lastTotal != 0 && total > lastTotal {
		lastHostCPU = 100 * (1 - float64(idle-lastIdle)/float64(total-lastTotal))
	}
	lastIdle, lastTotal = idle, total
	return lastHostCPU
}
