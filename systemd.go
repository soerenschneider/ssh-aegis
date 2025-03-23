package main

import (
	"cmp"
	"os/exec"
)

const defaultUnitName = "sshd"

type Systemd struct {
	unitName string
}

func (w *Systemd) UnitExists() error {
	//nolint G204
	cmd := exec.Command("systemctl", "status", cmp.Or(w.unitName, defaultUnitName))
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (w *Systemd) RestartSsh() error {
	//nolint G204
	cmd := exec.Command("systemctl", "restart", cmp.Or(w.unitName, defaultUnitName))
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
