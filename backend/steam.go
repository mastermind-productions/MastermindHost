package backend

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const steamcmdZipURL = "https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip"

func (h *Host) steamcmdPath() string {
	return filepath.Join(h.store.DataDir, "steamcmd", "steamcmd.exe")
}

func (h *Host) ensureSteamCmd(id string) error {
	exe := h.steamcmdPath()
	if _, err := os.Stat(exe); err == nil {
		return nil
	}
	h.appendLog(id, "Downloading SteamCMD...")
	dir := filepath.Dir(exe)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	resp, err := http.Get(steamcmdZipURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("steamcmd download failed: %s", resp.Status)
	}
	zipPath := filepath.Join(dir, "steamcmd.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)
	if err := unzip(zipPath, dir); err != nil {
		return err
	}
	h.appendLog(id, "SteamCMD downloaded")
	return nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		p := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(p), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path in zip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(p, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(p)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) appendLog(id, line string) {
	s := h.state(id)
	h.mu.Lock()
	s.installLog = append(s.installLog, line)
	if len(s.installLog) > 500 {
		s.installLog = s.installLog[len(s.installLog)-500:]
	}
	h.mu.Unlock()
	if h.ctx != nil {
		wruntime.EventsEmit(h.ctx, "gs-install-log", id, line)
	}
}

func (h *Host) GetInstallLog(id string) []string {
	s := h.state(id)
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string{}, s.installLog...)
}

var progressRe = regexp.MustCompile(`progress:\s*([\d.]+)`)

func (h *Host) runSteamCmd(ctx context.Context, id string, sv Server, validate bool) error {
	h.steamMu.Lock()
	defer h.steamMu.Unlock()
	if err := h.ensureSteamCmd(id); err != nil {
		return err
	}
	g := Games[sv.Game]
	args := []string{"+force_install_dir", sv.Dir, "+login", "anonymous", "+app_update", g.AppID}
	if sv.BetaBranch != "" {
		args = append(args, "-beta", sv.BetaBranch)
		if sv.BetaPassword != "" {
			args = append(args, "-betapassword", sv.BetaPassword)
		}
	}
	if validate {
		args = append(args, "validate")
	}
	args = append(args, "+quit")
	h.appendLog(id, "steamcmd "+strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, h.steamcmdPath(), args...)
	cmd.Dir = filepath.Dir(h.steamcmdPath())
	cmd.SysProcAttr = hideWindowAttr()
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	success := false
	sc := bufio.NewScanner(pipe)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	sc.Split(scanLinesCR)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if m := progressRe.FindStringSubmatch(line); m != nil {
			if p, err := strconv.ParseFloat(m[1], 64); err == nil {
				s := h.state(id)
				h.mu.Lock()
				s.progress = p
				h.mu.Unlock()
				h.emit(id)
			}
		}
		if strings.Contains(line, "Success! App") && strings.Contains(line, "fully installed") {
			success = true
		}
		h.appendLog(id, line)
	}
	werr := cmd.Wait()
	if ctx.Err() != nil {
		return errors.New("cancelled")
	}
	if !success {
		if werr != nil {
			return fmt.Errorf("steamcmd failed: %v", werr)
		}
		return errors.New("steamcmd did not report success - see install log")
	}
	return nil
}

func scanLinesCR(data []byte, atEOF bool) (int, []byte, error) {
	for i, b := range data {
		if b == '\n' || b == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (h *Host) beginBusy(id string) (context.Context, error) {
	s := h.state(id)
	h.mu.Lock()
	defer h.mu.Unlock()
	if s.busy {
		return nil, errors.New("another operation is already in progress")
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.busy, s.cancel, s.progress, s.installLog = true, cancel, 0, nil
	return ctx, nil
}

func (h *Host) endBusy(id string) {
	s := h.state(id)
	h.mu.Lock()
	s.busy, s.progress = false, 0
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	h.mu.Unlock()
}

func (h *Host) CancelOperation(id string) {
	s := h.state(id)
	h.mu.Lock()
	c := s.cancel
	h.mu.Unlock()
	if c != nil {
		c()
	}
}

func (h *Host) Install(id string, validate bool) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if h.isRunning(id) {
		return errors.New("stop the server before installing")
	}
	ctx, err := h.beginBusy(id)
	if err != nil {
		return err
	}
	status := StatusInstalling
	if h.installed(sv) {
		status = StatusUpdating
	}
	h.setStatus(id, status)
	go func() {
		os.MkdirAll(sv.Dir, 0o755)
		err := h.runSteamCmd(ctx, id, sv, validate)
		h.endBusy(id)
		h.setStatus(id, h.idleStatus(sv))
		h.logAction(id, status, statusOf(err), errStr(err))
		if err == nil {
			h.appendLog(id, "Done.")
		} else {
			h.appendLog(id, "ERROR: "+err.Error())
		}
	}()
	return nil
}

func (h *Host) Update(id string, validate bool) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	wasRunning := h.isRunning(id)
	if !wasRunning {
		return h.Install(id, validate)
	}
	go func() {
		h.setStatus(id, StatusShuttingDown)
		if err := h.shutdownInternal(id); err != nil {
			h.stopInternal(id)
		}
		h.setStatus(id, StatusOffline)
		if err := h.Install(id, validate); err != nil {
			h.logAction(id, "update", "error", err.Error())
			return
		}
		for h.state(id).busy {
			time.Sleep(time.Second)
		}
		if h.installed(sv) {
			h.Start(id)
		}
	}()
	return nil
}

var buildIDRe = regexp.MustCompile(`"buildid"\s+"(\d+)"`)

func readBuildID(dir, appID string) string {
	b, err := os.ReadFile(filepath.Join(dir, "steamapps", "appmanifest_"+appID+".acf"))
	if err != nil {
		return ""
	}
	if m := buildIDRe.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	return ""
}

func (h *Host) latestBuild(appID, branch string) string {
	if branch == "" {
		branch = "public"
	}
	h.latestMu.Lock()
	defer h.latestMu.Unlock()
	return h.latest[appID+"|"+branch]
}

func (h *Host) fetchLatest(appID string) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get("https://api.steamcmd.net/v1/info/" + appID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body struct {
		Data map[string]struct {
			Depots struct {
				Branches map[string]struct {
					BuildID string `json:"buildid"`
				} `json:"branches"`
			} `json:"depots"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	app, ok := body.Data[appID]
	if !ok {
		return errors.New("app not found")
	}
	h.latestMu.Lock()
	for name, b := range app.Depots.Branches {
		h.latest[appID+"|"+name] = b.BuildID
	}
	h.latestMu.Unlock()
	return nil
}

func (h *Host) CheckForUpdates() error {
	seen := map[string]bool{}
	var last error
	for _, sv := range h.store.List() {
		app := Games[sv.Game].AppID
		if seen[app] {
			continue
		}
		seen[app] = true
		if err := h.fetchLatest(app); err != nil {
			last = err
		}
	}
	for _, sv := range h.store.List() {
		h.emit(sv.ID)
	}
	return last
}

func (h *Host) latestBuildLoop() {
	for {
		h.CheckForUpdates()
		time.Sleep(15 * time.Minute)
	}
}
