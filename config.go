package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"slices"
)

const (
	configDefaultMetricsFile        = "/var/lib/node_exporter/ssh_aegis.prom"
	configDefaultSshServiceName     = "sshd"
	configDefaultWireguardInterface = "wg0"
	configDefaultSshdConfigFile     = "/etc/ssh/sshd_config"
)

type SshAegisConfig struct {
	ListenAddressesUp      []string `json:"up"`
	ListenAddressesDown    []string `json:"down"`
	ListenAddressesUnknown []string `json:"unknown,omitempty"`
	SshdConfigFile         string   `json:"sshd_config_file,omitempty"`
	WireguardInterface     string   `json:"wg,omitempty"`
	SshServiceName         string   `json:"ssh_service_name"`
	MetricsFile            string   `json:"metrics_file"`
}

func (c *SshAegisConfig) Validate() error {
	if slices.Equal(c.ListenAddressesUp, c.ListenAddressesDown) {
		return errors.New("addresses for up and down are equal")
	}

	if len(c.ListenAddressesUp) == 0 {
		return errors.New("no addresses configured for tunnel status 'up'")
	}

	if len(c.ListenAddressesDown) == 0 {
		return errors.New("no addresses configured for tunnel status 'down'")
	}

	for _, addr := range append(c.ListenAddressesUp, c.ListenAddressesDown...) {
		if net.ParseIP(addr) == nil {
			return fmt.Errorf("invalid address supplied: %s", addr)
		}
	}

	if len(c.ListenAddressesUnknown) > 0 {
		for _, addr := range c.ListenAddressesUnknown {
			if net.ParseIP(addr) == nil {
				return fmt.Errorf("invalid address supplied: %s", addr)
			}
		}
	}

	_, err := os.Stat(c.SshdConfigFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("sshd config file does not exist: %s", c.SshdConfigFile)
	}

	if c.SshServiceName == "" {
		return errors.New("empty ssh service name provided")
	}

	if c.WireguardInterface == "" {
		return errors.New("empty wg interface name provided")
	}

	return nil
}

func (c *SshAegisConfig) printConfig() {
	slog.Info("Using config", "wg_interface", c.WireguardInterface)
	slog.Info("Using config", "sshd_config", c.SshdConfigFile)
	slog.Info("Using config", "status", "up", "addresses", c.ListenAddressesUp)
	slog.Info("Using config", "status", "down", "addresses", c.ListenAddressesDown)
	if len(c.ListenAddressesUnknown) > 0 {
		slog.Info("Using config", "status", "unknown", "addresses", c.ListenAddressesUnknown)
	}
}

func getDefault() SshAegisConfig {
	return SshAegisConfig{
		ListenAddressesDown: []string{"0.0.0.0"},
		SshdConfigFile:      configDefaultSshdConfigFile,
		WireguardInterface:  configDefaultWireguardInterface,
		SshServiceName:      configDefaultSshServiceName,
		MetricsFile:         configDefaultMetricsFile,
	}
}

func readConfig(file string) (*SshAegisConfig, error) {
	conf := getDefault()

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &conf)
	return &conf, err
}
