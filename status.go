package main

import (
	"os/exec"
	"strings"
)

type TunnelStatus int

const (
	Unknown TunnelStatus = iota
	Up      TunnelStatus = iota
	Down    TunnelStatus = iota
)

func (s TunnelStatus) String() string {
	switch s {
	case Up:
		return "up"
	case Down:
		return "down"
	case Unknown:
		return "unknown"
	}
	return "unknown"
}

type WgStatus struct {
	interfaceName string
}

func (w *WgStatus) GetStatus() TunnelStatus {
	//nolint G204
	cmd := exec.Command("wg", "show", w.interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return Down
	}

	if strings.TrimSpace(string(output)) != "" {
		return Up
	}

	return Down
}
