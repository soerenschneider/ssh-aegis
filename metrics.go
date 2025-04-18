package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"
)

const templateData = `# HELP ssh_aegis_version version information for the running binary
# TYPE ssh_aegis_version gauge
ssh_aegis_version{app="{{ index .Version "app" }}",go="{{ index .Version "go" }}"} 1
# HELP ssh_aegis_timestamp_seconds the timestamp of the invocation
# TYPE ssh_aegis_timestamp_seconds gauge
ssh_aegis_timestamp_seconds {{ .Now }}
# HELP ssh_aegis_status represents the status of the tunnel
# TYPE ssh_aegis_status gauge
ssh_aegis_status{status="{{ .Status }}"} 1
# HELP ssh_aegis_last_status_change_timestamp_seconds represents the status of the tunnel
# TYPE ssh_aegis_last_status_change_timestamp_seconds gauge
ssh_aegis_last_status_change_timestamp_seconds {{ .LastStatusChange }}
# HELP ssh_aegis_restart_ssh_errors Number of SSH restart errors encountered.
# TYPE ssh_aegis_restart_ssh_errors counter
ssh_aegis_restart_ssh_errors {{ .RestartSshErrors }}
# HELP ssh_aegis_config_read_errors Number of errors encountered while reading the config.
# TYPE ssh_aegis_config_read_errors counter
ssh_aegis_config_read_errors {{ .ConfigReadErrors }}
# HELP ssh_aegis_config_write_errors Number of errors encountered while writing the config.
# TYPE ssh_aegis_config_write_errors counter
ssh_aegis_config_write_errors {{ .ConfigWriteErrors }}
`

var metrics = Metrics{
	Version: map[string]string{
		"go":  GoVersion,
		"app": BuildVersion,
	},
	Now:    time.Now().Unix(),
	Status: Unknown,
}

type Metrics struct {
	Version           map[string]string
	Now               int64
	Status            TunnelStatus
	LastStatusChange  int64
	RestartSshErrors  int
	ConfigReadErrors  int
	ConfigWriteErrors int
}

type MetricsWriter struct {
	tmpl        *template.Template
	metricsFile string
}

func NewMetricsWriter(metricsFile string) (*MetricsWriter, error) {
	tmpl, err := template.New("metrics").Parse(templateData)
	if err != nil {
		return nil, err
	}

	return &MetricsWriter{
		tmpl:        tmpl,
		metricsFile: metricsFile,
	}, nil
}

func (m *MetricsWriter) Dump() error {
	metrics.Now = time.Now().Unix()

	tmpFile := fmt.Sprintf("%s.tmp", m.metricsFile)
	file, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	if err := m.tmpl.Execute(file, metrics); err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	return os.Rename(tmpFile, m.metricsFile)
}
