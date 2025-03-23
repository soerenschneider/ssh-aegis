package main

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	listenAddressConfiguration = "ListenAddress "
)

type ConfigWrapper interface {
	GetConfig() ([]string, error)
	WriteConfig(data []string) error
}

type TunnelStatusSource interface {
	GetStatus() TunnelStatus
}

type ServiceReloader interface {
	RestartSsh() error
	UnitExists() error
}

type SshAegis struct {
	configWrapper      ConfigWrapper
	tunnelStatusSource TunnelStatusSource
	serviceProvider    ServiceReloader

	addressConfiguration map[TunnelStatus][]string
	oldStatus            TunnelStatus
}

func NewSshAegis(configWrapper ConfigWrapper, tunnelStatusSource TunnelStatusSource, serviceProvider ServiceReloader, options *SshAegisConfig) (*SshAegis, error) {
	if configWrapper == nil {
		return nil, errors.New("no ssh config wrapper provided")
	}

	if options == nil {
		return nil, errors.New("nil options provided")
	}

	return &SshAegis{
		configWrapper:      configWrapper,
		tunnelStatusSource: tunnelStatusSource,
		serviceProvider:    serviceProvider,
		oldStatus:          Unknown,
		addressConfiguration: map[TunnelStatus][]string{
			Up:      options.ListenAddressesUp,
			Down:    options.ListenAddressesDown,
			Unknown: options.ListenAddressesUnknown,
		},
	}, nil
}

func (s *SshAegis) Check() {
	status := s.tunnelStatusSource.GetStatus()
	metrics.MetricStatus = status

	if s.oldStatus != status {
		slog.Info("Status changed", "from", s.oldStatus, "to", status)
		s.oldStatus = status
		if err := s.upsert(status); err != nil {
			slog.Error("could not upsert status", "err", err)
		}

		metrics.MetricLastStatusChange = time.Now().Unix()
	}
}

func (s *SshAegis) upsert(status TunnelStatus) error {
	if status == Unknown && len(s.addressConfiguration[Unknown]) == 0 {
		slog.Debug("Ignoring status 'unknown'")
		return nil
	}

	wanted := s.addressConfiguration[status]
	updateNeeded, err := s.isUpdateNeeded(wanted)
	if err != nil {
		return err
	}
	if updateNeeded {
		slog.Info("Updating ListenAddress configuration", "addresses", wanted)
		if err := s.SetConfiguredListenAddresses(wanted); err != nil {
			return err
		}

		if err := s.serviceProvider.RestartSsh(); err != nil {
			metrics.MetricRestartSshErrors++
		}
	} else {
		slog.Info("No updates needed")
	}

	return nil
}

func (s *SshAegis) isUpdateNeeded(wantedListenAddresses []string) (bool, error) {
	configuredListenAddresses, err := s.getConfiguredListenAddresses()
	if err != nil {
		return false, err
	}

	if len(configuredListenAddresses) != len(wantedListenAddresses) {
		return true, nil
	}

	for _, addr := range wantedListenAddresses {
		if !slices.Contains(configuredListenAddresses, addr) {
			return true, nil
		}
	}

	return false, nil
}

func (s *SshAegis) SetConfiguredListenAddresses(wanted []string) error {
	data, err := s.configWrapper.GetConfig()
	if err != nil {
		metrics.MetricReadConfigErrors++
		return err
	}

	// get indices of lines containing active ListenAddress configuration and then remove these indices from the slice
	listenAddressConfigLinesIndices := getListenAddressIndices(data)
	sort.Sort(sort.Reverse(sort.IntSlice(listenAddressConfigLinesIndices))) // Make sure indices to be removed are in descending order
	for _, index := range listenAddressConfigLinesIndices {
		if index >= 0 && index < len(data) {
			data = append(data[:index], data[index+1:]...)
		}
	}

	insertBlock := make([]string, len(wanted))
	for idx, wantedIp := range wanted {
		insertBlock[idx] = fmt.Sprintf("%s%s", listenAddressConfiguration, wantedIp)
	}

	index := 0
	if len(listenAddressConfigLinesIndices) > 0 {
		index = listenAddressConfigLinesIndices[len(listenAddressConfigLinesIndices)-1]
	}
	data = append(data[:index+1], append(insertBlock, data[index+1:]...)...)

	if err := s.configWrapper.WriteConfig(data); err != nil {
		metrics.MetricWriteConfigErrors++
		return err
	}

	return nil
}

func (s *SshAegis) getConfiguredListenAddresses() ([]string, error) {
	data, err := s.configWrapper.GetConfig()
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, line := range data {
		if strings.HasPrefix(line, listenAddressConfiguration) {
			addresses = append(addresses, strings.ReplaceAll(line, listenAddressConfiguration, ""))
		}
	}

	return addresses, nil
}

func getListenAddressIndices(lines []string) []int {
	var indices []int
	for i, line := range lines {
		if strings.HasPrefix(line, listenAddressConfiguration) {
			indices = append(indices, i)
		}
	}

	return indices
}
