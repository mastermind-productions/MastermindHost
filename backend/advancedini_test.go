package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMetaForStatArrays(t *testing.T) {
	m, ok := metaFor("gameIni", "PerLevelStatsMultiplier_Player[7]")
	if !ok || !strings.Contains(m.desc, "Weight") || m.kind != "number" {
		t.Fatalf("player[7] meta = %+v ok=%v", m, ok)
	}
	m, ok = metaFor("gameIni", "PerLevelStatsMultiplier_DinoWild[9]")
	if !ok || !strings.Contains(m.desc, "Movement Speed") {
		t.Fatalf("dinowild[9] meta = %+v ok=%v", m, ok)
	}
	m, ok = metaFor("gameIni", "PerLevelStatsMultiplier_DinoTamed_Add[3]")
	if !ok || !strings.Contains(m.desc, "index 3") {
		t.Fatalf("add[3] meta = %+v ok=%v", m, ok)
	}
	if _, ok := metaFor("gameIni", "NotARealKey"); ok {
		t.Fatal("unknown key should not match catalog")
	}
}

func TestNatLess(t *testing.T) {
	if !natLess("perlevelstatsmultiplier_player[2]", "perlevelstatsmultiplier_player[10]") {
		t.Fatal("[2] should sort before [10]")
	}
	if natLess("perlevelstatsmultiplier_player[10]", "perlevelstatsmultiplier_player[2]") {
		t.Fatal("[10] should not sort before [2]")
	}
}

func TestAdvancedIniListAndApply(t *testing.T) {
	t.Setenv("MASTERMINDHOST_DATA", t.TempDir())
	h, err := NewHost()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := h.store.Add(&Server{ID: "adv", Name: "Adv", Game: "arksa", Dir: dir, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	entries, err := h.AdvancedIniList("adv")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected catalog entries for a fresh server")
	}
	foundUnset := false
	for _, e := range entries {
		if e.Key == "XPMultiplier" && !e.Set {
			foundUnset = true
		}
		if e.Set {
			t.Fatalf("no files exist yet, entry %s should be unset", e.Key)
		}
	}
	if !foundUnset {
		t.Fatal("catalog should offer XPMultiplier")
	}

	gusDir := filepath.Join(dir, "ShooterGame", "Saved", "Config", "WindowsServer")
	if err := os.MkdirAll(gusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	gus := filepath.Join(gusDir, "GameUserSettings.ini")
	if err := os.WriteFile(gus, []byte("[ServerSettings]\r\nServerPVE=True\r\nXPMultiplier=1.0\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err = h.AdvancedIniList("adv")
	if err != nil {
		t.Fatal(err)
	}
	gotPVE, gotXP := false, false
	for _, e := range entries {
		switch e.Key {
		case "ServerPVE":
			gotPVE = e.Set && e.Value == "True" && e.Kind == "bool" && e.Group == "PvP / PvE"
		case "XPMultiplier":
			gotXP = e.Set && e.Value == "1.0" && e.Kind == "number"
		}
	}
	if !gotPVE || !gotXP {
		t.Fatalf("parsed entries wrong: ServerPVE ok=%v XPMultiplier ok=%v", gotPVE, gotXP)
	}

	if err := h.AdvancedIniApply("adv", []AdvancedIniChange{
		{File: "gusIni", Section: "ServerSettings", Key: "XPMultiplier", Value: "2.5"},
		{File: "gusIni", Section: "ServerSettings", Key: "ServerPVE", Remove: true},
		{File: "gameIni", Section: shootGameMode, Key: "PerLevelStatsMultiplier_Player[7]", Value: "3.0"},
	}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(gus)
	if err != nil {
		t.Fatal(err)
	}
	c := string(b)
	if !strings.Contains(c, "XPMultiplier=2.5") || strings.Contains(c, "ServerPVE") {
		t.Fatalf("GUS apply wrong:\n%s", c)
	}
	game := filepath.Join(gusDir, "Game.ini")
	gb, err := os.ReadFile(game)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gb), "PerLevelStatsMultiplier_Player[7]=3.0") || !strings.Contains(string(gb), "["+shootGameMode+"]") {
		t.Fatalf("Game.ini apply wrong:\n%s", string(gb))
	}

	if err := h.AdvancedIniApply("adv", []AdvancedIniChange{{File: "bogus", Section: "x", Key: "y"}}); err == nil {
		t.Fatal("bogus file key should be rejected")
	}
}
