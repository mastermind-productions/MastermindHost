package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (h *Host) firewallRuleName(sv Server) string {
	return "MastermindHost - " + sanitizeFolderName(sv.Name)
}

func (h *Host) FirewallStatus(id string) (bool, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return false, err
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name="+h.firewallRuleName(sv))
	cmd.SysProcAttr = hideWindowAttr()
	out, err := cmd.CombinedOutput()
	return err == nil && strings.Contains(string(out), h.firewallRuleName(sv)), nil
}

func (h *Host) OpenFirewall(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	ports := []string{}
	for _, p := range h.runtime(sv).udpPorts(Games[sv.Game].Kind) {
		ports = append(ports, fmt.Sprint(p))
	}
	name := h.firewallRuleName(sv)
	exe := h.exePath(sv)
	lines := []string{
		"@echo off",
		fmt.Sprintf(`netsh advfirewall firewall delete rule name="%s" >nul 2>&1`, name),
		fmt.Sprintf(`netsh advfirewall firewall add rule name="%s" dir=in action=allow protocol=UDP localport=%s`, name, strings.Join(ports, ",")),
		fmt.Sprintf(`netsh advfirewall firewall add rule name="%s" dir=in action=allow program="%s" enable=yes`, name, exe),
	}
	bat := filepath.Join(h.store.DataDir, "firewall.bat")
	if err := os.WriteFile(bat, []byte(strings.Join(lines, "\r\n")+"\r\n"), 0o644); err != nil {
		return err
	}
	err = runElevated(bat)
	h.logAction(id, "firewall open", statusOf(err), errStr(err))
	return err
}

func (h *Host) CloseFirewall(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	bat := filepath.Join(h.store.DataDir, "firewall.bat")
	content := fmt.Sprintf("@echo off\r\nnetsh advfirewall firewall delete rule name=\"%s\"\r\n", h.firewallRuleName(sv))
	if err := os.WriteFile(bat, []byte(content), 0o644); err != nil {
		return err
	}
	err = runElevated(bat)
	h.logAction(id, "firewall close", statusOf(err), errStr(err))
	return err
}
