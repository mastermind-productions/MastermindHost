package backend

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (h *Host) configPath(sv Server, key string) (string, error) {
	g, err := gameDef(sv.Game)
	if err != nil {
		return "", err
	}
	c, err := g.configFile(key)
	if err != nil {
		return "", err
	}
	if c.RelPath == "" {
		return "", errors.New("config is not file based")
	}
	return filepath.Join(sv.Dir, c.RelPath), nil
}

func (h *Host) readConfigSilently(sv Server, key string) string {
	p, err := h.configPath(sv, key)
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(string(b), "\ufeff")
}

func (h *Host) GetConfigFiles(id string) ([]ConfigFileDef, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	return Games[sv.Game].ConfigFiles, nil
}

func (h *Host) ReadConfig(id, key string) (string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return "", err
	}
	if key == "cmd" {
		return sv.Cmd, nil
	}
	return h.readConfigSilently(sv, key), nil
}

func (h *Host) WriteConfig(id, key, content string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if key == "cmd" {
		return h.SetCmd(id, content)
	}
	p, err := h.configPath(sv, key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if strings.HasSuffix(strings.ToLower(p), ".txt") {
		content = strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\n", "\r\n")
	}
	err = os.WriteFile(p, []byte(content), 0o644)
	h.logAction(id, "edit "+filepath.Base(p), statusOf(err), errStr(err))
	h.emit(id)
	return err
}

func (h *Host) SetCmd(id, cmd string) error {
	cmd = strings.Join(strings.Fields(strings.ReplaceAll(cmd, "\r\n", " ")), " ")
	err := h.store.Update(id, func(s *Server) error { s.Cmd = cmd; return nil })
	if err == nil {
		h.logAction(id, "edit command line", "success", "")
		h.emit(id)
	}
	return err
}

func (h *Host) GetParsedCmd(id string) (ArkCmd, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return ArkCmd{}, err
	}
	return ParseArkCmd(sv.Cmd), nil
}

func (h *Host) SetParsedCmd(id string, c ArkCmd) error {
	return h.SetCmd(id, c.String())
}

func (h *Host) GetIni(id, key string) (map[string]map[string]string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	return ParseIni(h.readConfigSilently(sv, key)), nil
}

func (h *Host) ApplyIni(id, key string, changes []IniChange) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	return h.WriteConfig(id, key, ApplyIniChanges(h.readConfigSilently(sv, key), changes))
}

func (h *Host) ReadList(id, key string) ([]string, error) {
	c, err := h.ReadConfig(id, key)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, l := range splitLines(c) {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out, nil
}

func (h *Host) WriteList(id, key string, items []string) error {
	var clean []string
	seen := map[string]bool{}
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it != "" && !seen[it] {
			seen[it] = true
			clean = append(clean, it)
		}
	}
	return h.WriteConfig(id, key, strings.Join(clean, "\r\n"))
}

type ModInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Host) GetMods(id string) ([]ModInfo, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	out := []ModInfo{}
	for _, m := range ParseArkCmd(sv.Cmd).Mods() {
		out = append(out, ModInfo{ID: m, Name: sv.ModNames[m]})
	}
	return out, nil
}

func (h *Host) SetMods(id string, mods []ModInfo) error {
	ids := []string{}
	err := h.store.Update(id, func(s *Server) error {
		if s.ModNames == nil {
			s.ModNames = map[string]string{}
		}
		seen := map[string]bool{}
		for _, m := range mods {
			mid := strings.TrimSpace(m.ID)
			if mid == "" || seen[mid] {
				continue
			}
			seen[mid] = true
			ids = append(ids, mid)
			if m.Name != "" {
				s.ModNames[mid] = m.Name
			}
		}
		c := ParseArkCmd(s.Cmd)
		c.SetMods(ids)
		s.Cmd = c.String()
		return nil
	})
	if err == nil {
		h.logAction(id, "update mods", "success", strings.Join(ids, ","))
		h.emit(id)
	}
	return err
}

func (h *Host) GetSettings() Settings { return h.store.Settings() }

func (h *Host) SetSettings(s Settings) error {
	if strings.TrimSpace(s.ServersRoot) == "" || strings.TrimSpace(s.BackupsRoot) == "" {
		return errors.New("folders cannot be empty")
	}
	return h.store.SetSettings(s)
}

func (h *Host) SetBeta(id, branch, password string) error {
	return h.store.Update(id, func(s *Server) error {
		s.BetaBranch, s.BetaPassword = strings.TrimSpace(branch), strings.TrimSpace(password)
		return nil
	})
}

func (h *Host) GetServerRaw(id string) (Server, error) {
	sv, err := h.store.Get(id)
	sv.Events = nil
	return sv, err
}

func (h *Host) OpenFolder(id, sub string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	p := filepath.Join(sv.Dir, sub)
	if _, err := os.Stat(p); err != nil {
		p = sv.Dir
	}
	return openInExplorer(p)
}

func (h *Host) OpenPath(p string) error { return openInExplorer(p) }

func (h *Host) SelectFolder(title string) (string, error) {
	return wruntime.OpenDirectoryDialog(h.ctx, wruntime.OpenDialogOptions{Title: title})
}
