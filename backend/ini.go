package backend

import (
	"strings"
)

type IniChange struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Remove  bool   `json:"remove"`
}

func splitLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if content == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(content, "\n"), "\n")
}

func sectionName(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
		return t[1 : len(t)-1], true
	}
	return "", false
}

func ParseIni(content string) map[string]map[string]string {
	out := map[string]map[string]string{}
	cur := ""
	for _, line := range splitLines(content) {
		if s, ok := sectionName(line); ok {
			cur = s
			if out[cur] == nil {
				out[cur] = map[string]string{}
			}
			continue
		}
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, ";") || strings.HasPrefix(t, "#") {
			continue
		}
		if i := strings.Index(t, "="); i > 0 {
			if out[cur] == nil {
				out[cur] = map[string]string{}
			}
			k := strings.TrimSpace(t[:i])
			v := t[i+1:]
			if prev, ok := out[cur][k]; ok {
				v = prev + "\n" + v
			}
			out[cur][k] = v
		}
	}
	return out
}

func ApplyIniChanges(content string, changes []IniChange) string {
	lines := splitLines(content)
	for _, ch := range changes {
		lines = applyIniChange(lines, ch)
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}

func applyIniChange(lines []string, ch IniChange) []string {
	cur := ""
	sectionStart, sectionEnd := -1, -1
	var hits []int
	for i, line := range lines {
		if s, ok := sectionName(line); ok {
			if cur == ch.Section && sectionStart >= 0 && sectionEnd < 0 {
				sectionEnd = i
			}
			cur = s
			if s == ch.Section && sectionStart < 0 {
				sectionStart = i
			}
			continue
		}
		if cur != ch.Section {
			continue
		}
		t := strings.TrimSpace(line)
		if j := strings.Index(t, "="); j > 0 && strings.EqualFold(strings.TrimSpace(t[:j]), ch.Key) {
			hits = append(hits, i)
		}
	}
	if sectionStart >= 0 && sectionEnd < 0 {
		sectionEnd = len(lines)
	}
	values := []string{}
	if !ch.Remove {
		values = strings.Split(ch.Value, "\n")
	}
	newLines := make([]string, 0, len(values))
	for _, v := range values {
		newLines = append(newLines, ch.Key+"="+v)
	}
	if len(hits) > 0 {
		out := make([]string, 0, len(lines)+len(newLines))
		hitSet := map[int]bool{}
		for _, h := range hits {
			hitSet[h] = true
		}
		for i, l := range lines {
			if i == hits[0] {
				out = append(out, newLines...)
			}
			if !hitSet[i] {
				out = append(out, l)
			}
		}
		return out
	}
	if ch.Remove {
		return lines
	}
	if sectionStart < 0 {
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "["+ch.Section+"]")
		return append(lines, newLines...)
	}
	insert := sectionEnd
	for insert > sectionStart+1 && strings.TrimSpace(lines[insert-1]) == "" {
		insert--
	}
	out := append([]string{}, lines[:insert]...)
	out = append(out, newLines...)
	return append(out, lines[insert:]...)
}
