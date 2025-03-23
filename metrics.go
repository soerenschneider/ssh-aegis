package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"
)

const templateData = `# HELP ssh_aegis_timestamp_seconds the timestamp of the invocation
# TYPE ssh_aegis_timestamp_seconds gauge
ssh_aegis_timestamp_seconds {{ .MetricNow }}
# HELP ssh_aegis_status represents the status of the tunnel
# TYPE ssh_aegis_status gauge
ssh_aegis_status{status="{{ .MetricStatus }}"} 1
# HELP ssh_aegis_last_status_change_timestamp_seconds represents the status of the tunnel
# TYPE ssh_aegis_last_status_change_timestamp_seconds gauge
ssh_aegis_last_status_change_timestamp_seconds {{ .MetricLastStatusChange }}
# HELP ssh_aegis_restart_ssh_errors Number of SSH restart errors encountered.
# TYPE ssh_aegis_restart_ssh_errors counter
ssh_aegis_restart_ssh_errors {{ .MetricRestartSshErrors }}
# HELP ssh_aegis_read_config_errors Number of errors encountered while reading the config.
# TYPE ssh_aegis_read_config_errors counter
ssh_aegis_read_config_errors {{ .MetricReadConfigErrors }}
# HELP ssh_aegis_write_config_errors Number of errors encountered while writing the config.
# TYPE ssh_aegis_write_config_errors counter
ssh_aegis_write_config_errors {{ .MetricWriteConfigErrors }}
`

var metrics = Metrics{
	MetricNow:    time.Now().Unix(),
	MetricStatus: Unknown,
}

type Metrics struct {
	MetricNow               int64
	MetricStatus            TunnelStatus
	MetricLastStatusChange  int64
	MetricRestartSshErrors  int
	MetricReadConfigErrors  int
	MetricWriteConfigErrors int
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
	metrics.MetricNow = time.Now().Unix()

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
