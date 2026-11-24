package backend

import (
	"strings"
	"testing"
)

func TestCurseForgeResultPatternExtractsLiveModID(t *testing.T) {
	page := `{"title":"Go to Ark Supreme Project Page","projectId":1442825}`
	match := curseForgeResultPattern.FindStringSubmatch(page)
	if len(match) != 3 {
		t.Fatalf("expected to parse a CurseForge result, got %v", match)
	}
	if got, want := strings.TrimSpace(match[1]), "Ark Supreme"; got != want {
		t.Errorf("mod name = %q, want %q", got, want)
	}
	if got, want := match[2], "1442825"; got != want {
		t.Errorf("mod ID = %q, want %q", got, want)
	}
}
