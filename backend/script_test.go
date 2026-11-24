package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverRealFolders(t *testing.T) {
	for game, dir := range map[string]string{"rappelz": `F:\MMRappelz`, "hero": `F:\DragonzOnline\HeroRebirthFork`, "scions": `F:\HollowFateOnline`} {
		if _, err := os.Stat(dir); err != nil {
			t.Skip("server folders not present")
		}
		sp, err := discoverScript(game, dir)
		if err != nil {
			t.Fatalf("%s: %v", game, err)
		}
		for _, p := range sp.Procs {
			if _, err := os.Stat(sp.exe(p)); err != nil {
				t.Errorf("%s: %s missing: %v", game, p.Name, err)
			}
		}
		t.Logf("%s: start=%q procs=%d", game, sp.StartCmd, len(sp.Procs))
	}
}

func TestRappelzStartCmd(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "Start-Servers.bat"), []byte("@echo off"), 0o644)
	sp, err := discoverScript("rappelz", d)
	if err != nil || !strings.HasSuffix(sp.StartCmd, `Start-Servers.bat" -NoPause`) {
		t.Fatalf("%v %q", err, sp.StartCmd)
	}
}

func TestReorder(t *testing.T) {
	t.Setenv("MASTERMINDHOST_DATA", t.TempDir())
	st, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b", "c", "d"} {
		st.Add(&Server{ID: id})
	}
	if err := st.Reorder([]string{"d", "b"}); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range st.List() {
		got = append(got, s.ID)
	}
	if strings.Join(got, "") != "adcb" {
		t.Fatalf("order %v", got)
	}
}
