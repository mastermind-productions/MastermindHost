package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreRecoversFromCorruptFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MASTERMINDHOST_DATA", dir)
	s, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetSettings(Settings{ServersRoot: "X:/a", BackupsRoot: "X:/b"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetSettings(Settings{ServersRoot: "X:/a", BackupsRoot: "X:/b"}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "store.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	s2, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if s2.Settings().ServersRoot != "X:/a" {
		t.Fatalf("not recovered: %q", s2.Settings().ServersRoot)
	}
	if _, err := os.Stat(filepath.Join(dir, "store.json.corrupt")); err != nil {
		t.Fatal(err)
	}
}
