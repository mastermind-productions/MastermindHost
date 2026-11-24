package backend

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type BackupInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Created int64  `json:"created"`
}

var datedArkRe = regexp.MustCompile(`_\d{2}\.\d{2}\.\d{4}_\d{2}\.\d{2}\.\d{2}\.ark(rbf)?$`)

func (h *Host) backupDir(sv Server) string {
	if strings.TrimSpace(sv.Backup.Dir) != "" {
		return sv.Backup.Dir
	}
	return filepath.Join(h.store.Settings().BackupsRoot, filepath.Base(sv.Dir))
}

func (h *Host) SetBackupSettings(id string, b BackupSettings) error {
	if b.Keep < 1 {
		b.Keep = 1
	}
	return h.store.Update(id, func(s *Server) error { s.Backup = b; return nil })
}

func (h *Host) Backup(id string) (string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return "", err
	}
	if h.isRunning(id) {
		h.rcon(sv, saveCmd(sv), 60*time.Second)
		time.Sleep(3 * time.Second)
	}
	path, err := h.createBackup(sv)
	h.logAction(id, "backup", statusOf(err), errStr(err))
	if err == nil && sv.Backup.AutoRemove {
		h.pruneBackups(sv)
	}
	return path, err
}

func (h *Host) createBackup(sv Server) (string, error) {
	g := Games[sv.Game]
	dir := h.backupDir(sv)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := sanitizeFolderName(sv.Name) + "_" + time.Now().Format("2006-01-02_15-04-05") + ".zip"
	out := filepath.Join(dir, name)
	f, err := os.Create(out + ".part")
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(f)
	roots := []string{g.SavesRel, arksaCfgDir}
	for _, c := range g.ConfigFiles {
		if c.RelPath != "" && !strings.HasPrefix(c.RelPath, arksaCfgDir) {
			roots = append(roots, c.RelPath)
		}
	}
	count := 0
	for _, rel := range roots {
		root := filepath.Join(sv.Dir, rel)
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if sv.Backup.SkipDatedWorldArks && datedArkRe.MatchString(d.Name()) {
				return nil
			}
			r, _ := filepath.Rel(sv.Dir, p)
			if err := addZipFile(zw, p, filepath.ToSlash(r)); err != nil {
				return err
			}
			count++
			return nil
		})
		if err != nil {
			break
		}
	}
	cerr := zw.Close()
	f.Close()
	if err == nil {
		err = cerr
	}
	if err == nil && count == 0 {
		err = errors.New("nothing to back up - no save or config files found")
	}
	if err != nil {
		os.Remove(out + ".part")
		return "", err
	}
	return out, os.Rename(out+".part", out)
}

func addZipFile(zw *zip.Writer, path, name string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()
	st, err := src.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(st)
	if err != nil {
		return err
	}
	hdr.Name, hdr.Method = name, zip.Deflate
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}

func (h *Host) ListBackups(id string) ([]BackupInfo, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	out := []BackupInfo{}
	entries, err := os.ReadDir(h.backupDir(sv))
	if err != nil {
		return out, nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".zip") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, BackupInfo{Name: e.Name(), Path: filepath.Join(h.backupDir(sv), e.Name()), Size: fi.Size(), Created: fi.ModTime().UnixMilli()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created > out[j].Created })
	return out, nil
}

func (h *Host) pruneBackups(sv Server) {
	list, _ := h.ListBackups(sv.ID)
	keep := sv.Backup.Keep
	if keep < 1 {
		keep = 1
	}
	for i := keep; i < len(list); i++ {
		os.Remove(list[i].Path)
	}
}

func (h *Host) backupPath(sv Server, name string) (string, error) {
	if name != filepath.Base(name) || !strings.HasSuffix(strings.ToLower(name), ".zip") {
		return "", errors.New("invalid backup name")
	}
	return filepath.Join(h.backupDir(sv), name), nil
}

func (h *Host) DeleteBackup(id, name string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	p, err := h.backupPath(sv, name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func (h *Host) RestoreBackup(id, name string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if h.isRunning(id) || h.state(id).busy {
		return errors.New("stop the server before restoring a backup")
	}
	p, err := h.backupPath(sv, name)
	if err != nil {
		return err
	}
	err = unzip(p, sv.Dir)
	h.logAction(id, "restore "+name, statusOf(err), errStr(err))
	return err
}

func (h *Host) OpenBackupFolder(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	d := h.backupDir(sv)
	os.MkdirAll(d, 0o755)
	return openInExplorer(d)
}
