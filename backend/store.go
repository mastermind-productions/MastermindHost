package backend

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Settings struct {
	ServersRoot string      `json:"serversRoot"`
	BackupsRoot string      `json:"backupsRoot"`
	Window      *WindowRect `json:"window,omitempty"`
}

type DiscordHook struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type BackupSettings struct {
	Dir                string `json:"dir"`
	AutoRemove         bool   `json:"autoRemove"`
	Keep               int    `json:"keep"`
	SkipDatedWorldArks bool   `json:"skipDatedWorldArks"`
}

type Mod struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Server struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Game         string            `json:"game"`
	Dir          string            `json:"dir"`
	Cmd          string            `json:"cmd"`
	CreatedAt    time.Time         `json:"createdAt"`
	ForceRunning bool              `json:"forceRunning"`
	BetaBranch   string            `json:"betaBranch"`
	BetaPassword string            `json:"betaPassword"`
	Discord      DiscordHook       `json:"discord"`
	Backup       BackupSettings    `json:"backup"`
	ModNames     map[string]string `json:"modNames"`
	Events       []Event           `json:"events"`
	Schedules    []Schedule        `json:"schedules"`
	Affinity     []int             `json:"affinity"`
	WindowRect   *WindowRect       `json:"windowRect,omitempty"`
	Script       *ScriptSpec       `json:"script,omitempty"`
}

type storeFile struct {
	Settings Settings  `json:"settings"`
	Servers  []*Server `json:"servers"`
}

type Store struct {
	mu      sync.RWMutex
	path    string
	DataDir string
	data    storeFile
}

func defaultDataDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "MastermindHost-data"
	}
	return filepath.Join(filepath.Dir(exe), "MastermindHost-data")
}

func NewStore() (*Store, error) {
	dir := os.Getenv("MASTERMINDHOST_DATA")
	if dir == "" {
		dir = defaultDataDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "store.json"), DataDir: dir}
	b, err := os.ReadFile(s.path)
	if err == nil {
		if err := json.Unmarshal(b, &s.data); err != nil {
			bb, berr := os.ReadFile(s.path + ".bak")
			if berr != nil {
				return nil, err
			}
			s.data = storeFile{}
			if err := json.Unmarshal(bb, &s.data); err != nil {
				return nil, err
			}
			_ = os.WriteFile(s.path+".corrupt", b, 0o644)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if s.data.Settings.ServersRoot == "" {
		s.data.Settings.ServersRoot = filepath.Join(dir, "servers")
	}
	if s.data.Settings.BackupsRoot == "" {
		s.data.Settings.BackupsRoot = filepath.Join(dir, "backups")
	}
	for _, sv := range s.data.Servers {
		if sv.ModNames == nil {
			sv.ModNames = map[string]string{}
		}
	}
	return s, s.saveLocked()
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if old, err := os.ReadFile(s.path); err == nil && len(old) > 0 {
		_ = os.WriteFile(s.path+".bak", old, 0o644)
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Settings() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Settings
}

func (s *Store) SetSettings(v Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v.Window = s.data.Settings.Window
	s.data.Settings = v
	return s.saveLocked()
}

func (s *Store) SetWindow(w WindowRect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Settings.Window = &w
	s.saveLocked()
}

func (s *Store) List() []Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Server, 0, len(s.data.Servers))
	for _, sv := range s.data.Servers {
		out = append(out, *sv)
	}
	return out
}

func (s *Store) Get(id string) (Server, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sv := range s.data.Servers {
		if sv.ID == id {
			return *sv, nil
		}
	}
	return Server{}, errors.New("game server not found")
}

func (s *Store) Add(sv *Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Servers = append(s.data.Servers, sv)
	return s.saveLocked()
}

func (s *Store) Update(id string, fn func(*Server) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sv := range s.data.Servers {
		if sv.ID == id {
			if err := fn(sv); err != nil {
				return err
			}
			return s.saveLocked()
		}
	}
	return errors.New("game server not found")
}

func (s *Store) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sv := range s.data.Servers {
		if sv.ID == id {
			s.data.Servers = append(s.data.Servers[:i], s.data.Servers[i+1:]...)
			return s.saveLocked()
		}
	}
	return errors.New("game server not found")
}

func (s *Store) Reorder(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos := map[string]int{}
	for i, sv := range s.data.Servers {
		pos[sv.ID] = i
	}
	slots := make([]int, 0, len(ids))
	byID := map[string]*Server{}
	for _, id := range ids {
		i, ok := pos[id]
		if !ok {
			return errors.New("game server not found")
		}
		if _, dup := byID[id]; dup {
			return errors.New("duplicate server in order list")
		}
		slots = append(slots, i)
		byID[id] = s.data.Servers[i]
	}
	sort.Ints(slots)
	for n, i := range slots {
		s.data.Servers[i] = byID[ids[n]]
	}
	return s.saveLocked()
}
