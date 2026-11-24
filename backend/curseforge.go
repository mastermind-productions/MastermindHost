package backend

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CurseForgeMod is a mod returned by CurseForge's live ASA search.
type CurseForgeMod struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

var curseForgeResultPattern = regexp.MustCompile(`(?s)Go to (.{1,120}?) Project Page.{0,300}?projectId.{0,30}?([0-9]{4,})`)

// SearchCurseForgeMods queries CurseForge's public, live ARK: Survival Ascended
// search page. It does not rely on a bundled catalogue or an API key.
func (h *Host) SearchCurseForgeMods(serverID, query string) ([]CurseForgeMod, error) {
	sv, err := h.store.Get(serverID)
	if err != nil {
		return nil, err
	}
	if sv.Game != "arksa" {
		return nil, fmt.Errorf("CurseForge mod search is only available for ARK: Survival Ascended")
	}
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return nil, fmt.Errorf("enter at least 2 characters to search")
	}
	if len([]rune(query)) > 100 {
		return nil, fmt.Errorf("search text is too long")
	}

	u, _ := url.Parse("https://www.curseforge.com/ark-survival-ascended/search")
	q := u.Query()
	q.Set("class", "mods")
	q.Set("page", "1")
	q.Set("pageSize", "20")
	q.Set("sortBy", "relevancy")
	q.Set("search", query)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MastermindHost/1.0 (CurseForge ASA mod search)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CurseForge search failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CurseForge search returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("could not read CurseForge search results: %w", err)
	}

	matches := curseForgeResultPattern.FindAllSubmatch(body, -1)
	results := make([]CurseForgeMod, 0, 10)
	seen := make(map[string]bool)
	for _, match := range matches {
		name := strings.TrimSpace(html.UnescapeString(string(match[1])))
		id := string(match[2])
		if name == "" || seen[id] {
			continue
		}
		seen[id] = true
		results = append(results, CurseForgeMod{
			ID:   id,
			Name: name,
			URL:  "https://www.curseforge.com/ark-survival-ascended/search?search=" + url.QueryEscape(name),
		})
		if len(results) == 20 {
			break
		}
	}
	return results, nil
}

func lookupModName(id string) string {
	client := &http.Client{Timeout: 20 * time.Second}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodGet, "https://api.cfwidget.com/"+url.PathEscape(id), nil)
		if err != nil {
			return ""
		}
		req.Header.Set("User-Agent", "MastermindHost/1.0")
		resp, err := client.Do(req)
		if err != nil {
			return ""
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if resp.StatusCode == http.StatusAccepted {
			time.Sleep(3 * time.Second)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return ""
		}
		var v struct {
			Title string `json:"title"`
		}
		if json.Unmarshal(body, &v) != nil {
			return ""
		}
		return strings.TrimSpace(v.Title)
	}
	return ""
}

func (h *Host) ResolveModNames(id string) (int, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, m := range ParseArkCmd(sv.Cmd).Mods() {
		if sv.ModNames[m] != "" {
			continue
		}
		if name := lookupModName(m); name != "" {
			h.store.Update(id, func(s *Server) error {
				if s.ModNames == nil {
					s.ModNames = map[string]string{}
				}
				s.ModNames[m] = name
				return nil
			})
			n++
		}
	}
	if n > 0 {
		h.emit(id)
	}
	return n, nil
}
