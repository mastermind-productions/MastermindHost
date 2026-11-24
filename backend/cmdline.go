package backend

import (
	"strconv"
	"strings"
)

type KV struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	HasValue bool   `json:"hasValue"`
}

type ArkCmd struct {
	Map     string `json:"map"`
	Options []KV   `json:"options"`
	Flags   []KV   `json:"flags"`
}

func tokenize(s string) []string {
	var out []string
	var cur strings.Builder
	inQ := false
	for _, r := range s {
		switch {
		case r == '"':
			inQ = !inQ
			cur.WriteRune(r)
		case (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inQ:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func splitKV(s string) KV {
	if i := strings.Index(s, "="); i >= 0 {
		return KV{Key: s[:i], Value: s[i+1:], HasValue: true}
	}
	return KV{Key: s}
}

func ParseArkCmd(s string) ArkCmd {
	c := ArkCmd{}
	for _, t := range tokenize(s) {
		if strings.HasPrefix(t, "-") {
			c.Flags = append(c.Flags, splitKV(t[1:]))
			continue
		}
		parts := strings.Split(t, "?")
		if c.Map == "" && parts[0] != "" {
			c.Map = parts[0]
		}
		for _, p := range parts[1:] {
			if p != "" {
				c.Options = append(c.Options, splitKV(p))
			}
		}
	}
	return c
}

func (c ArkCmd) String() string {
	var b strings.Builder
	b.WriteString(c.Map)
	for _, o := range c.Options {
		b.WriteString("?" + o.Key)
		if o.HasValue {
			b.WriteString("=" + o.Value)
		}
	}
	for _, f := range c.Flags {
		b.WriteString(" -" + f.Key)
		if f.HasValue {
			b.WriteString("=" + f.Value)
		}
	}
	return b.String()
}

func findKV(list []KV, key string) int {
	for i, kv := range list {
		if strings.EqualFold(kv.Key, key) {
			return i
		}
	}
	return -1
}

func (c ArkCmd) Opt(key string) (string, bool) {
	if i := findKV(c.Options, key); i >= 0 {
		return strings.Trim(c.Options[i].Value, `"`), true
	}
	return "", false
}

func (c ArkCmd) Flag(key string) (string, bool) {
	if i := findKV(c.Flags, key); i >= 0 {
		return strings.Trim(c.Flags[i].Value, `"`), true
	}
	return "", false
}

func (c *ArkCmd) SetOpt(key, value string) {
	kv := KV{Key: key, Value: value, HasValue: true}
	if i := findKV(c.Options, key); i >= 0 {
		c.Options[i] = kv
		return
	}
	c.Options = append(c.Options, kv)
}

func (c *ArkCmd) SetFlag(key, value string, hasValue bool) {
	kv := KV{Key: key, Value: value, HasValue: hasValue}
	if i := findKV(c.Flags, key); i >= 0 {
		c.Flags[i] = kv
		return
	}
	c.Flags = append(c.Flags, kv)
}

func (c *ArkCmd) RemoveFlag(key string) {
	if i := findKV(c.Flags, key); i >= 0 {
		c.Flags = append(c.Flags[:i], c.Flags[i+1:]...)
	}
}

func (c ArkCmd) IntOpt(key string, def int) int {
	if v, ok := c.Opt(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func (c ArkCmd) Mods() []string {
	v, ok := c.Flag("mods")
	if !ok || strings.TrimSpace(v) == "" {
		return nil
	}
	var out []string
	for _, m := range strings.Split(v, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

func (c *ArkCmd) SetMods(ids []string) {
	if len(ids) == 0 {
		c.RemoveFlag("mods")
		return
	}
	c.SetFlag("mods", strings.Join(ids, ","), true)
}
