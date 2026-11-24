package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var discordColors = map[string]int{
	StatusOnline: 0x2ecc71, StatusOffline: 0xe74c3c, StatusStarting: 0xf1c40f,
	StatusRestarting: 0xf39c12, StatusUpdating: 0x3498db, StatusInstalling: 0x3498db,
	StatusShuttingDown: 0xe67e22, StatusStopping: 0xe67e22,
}

func (h *Host) SetDiscord(id string, d DiscordHook) error {
	d.URL = strings.TrimSpace(d.URL)
	if d.URL != "" && !strings.HasPrefix(d.URL, "https://discord.com/api/webhooks/") && !strings.HasPrefix(d.URL, "https://discordapp.com/api/webhooks/") {
		return errors.New("not a Discord webhook URL")
	}
	return h.store.Update(id, func(s *Server) error { s.Discord = d; return nil })
}

func (h *Host) TestDiscord(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if sv.Discord.URL == "" {
		return errors.New("webhook URL is not set")
	}
	return postDiscord(sv.Discord.URL, sv.Name, "Test message from MastermindHost", 0x1f8b4c)
}

func (h *Host) discordNotify(id, status string) {
	color, ok := discordColors[status]
	if !ok {
		return
	}
	sv, err := h.store.Get(id)
	if err != nil || !sv.Discord.Enabled || sv.Discord.URL == "" {
		return
	}
	postDiscord(sv.Discord.URL, sv.Name, fmt.Sprintf("Server is now **%s**", status), color)
}

func postDiscord(url, title, desc string, color int) error {
	payload := map[string]any{
		"username": "MastermindHost",
		"embeds": []map[string]any{{
			"title": title, "description": desc, "color": color,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}},
	}
	b, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned %s", resp.Status)
	}
	return nil
}
