package backend

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LogFile struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

func (h *Host) logsDir(sv Server) string {
	return filepath.Dir(filepath.Join(sv.Dir, Games[sv.Game].LogRel))
}

func (h *Host) ListLogs(id string) ([]LogFile, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	out := []LogFile{}
	entries, err := os.ReadDir(h.logsDir(sv))
	if err != nil {
		return out, nil
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, LogFile{Name: e.Name(), Size: fi.Size(), Modified: fi.ModTime().UnixMilli()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Modified > out[j].Modified })
	return out, nil
}

func (h *Host) ReadLog(id, name string, maxBytes int64) (string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return "", err
	}
	if name == "" {
		name = filepath.Base(Games[sv.Game].LogRel)
	}
	if name != filepath.Base(name) {
		return "", os.ErrPermission
	}
	if maxBytes <= 0 {
		maxBytes = 256 * 1024
	}
	f, err := os.Open(filepath.Join(h.logsDir(sv), name))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()
	st, _ := f.Stat()
	off := int64(0)
	if st.Size() > maxBytes {
		off = st.Size() - maxBytes
	}
	f.Seek(off, io.SeekStart)
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	s := string(b)
	if off > 0 {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		}
	}
	return s, nil
}
