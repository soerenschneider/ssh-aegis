package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	defaultConfigFile = "/etc/ssh-aegis.json"
)

var (
	flagConfigFile   string
	flagDebug        bool
	flagPrintVersion bool

	BuildVersion string
	CommitHash   string
)

func parseFlags() {
	flag.StringVar(&flagConfigFile, "config", defaultConfigFile, "Path of config file")
	flag.BoolVar(&flagDebug, "debug", false, "Print debug logs")
	flag.BoolVar(&flagPrintVersion, "version", false, "Print version and exit")
	flag.Parse()
}

func main() {
	parseFlags()

	if flagPrintVersion {
		//nolint forbidigo
		fmt.Printf("%s %s\n", BuildVersion, CommitHash)
		os.Exit(0)
	}

	slog.Info("Starting ssh-aegis", "version", BuildVersion)
	slog.Info("Reading config", "file", flagConfigFile)
	config, err := readConfig(flagConfigFile)
	if err != nil {
		log.Fatal("could not read config: ", err)
	}

	slog.Info("Validating config")
	if err := config.Validate(); err != nil {
		log.Fatal("config is invalid: ", err)
	}
	config.printConfig()

	var statusSource TunnelStatusSource = &WgStatus{interfaceName: config.WireguardInterface}
	var serviceProvider ServiceReloader = &Systemd{}

	slog.Info("Checking if ssh service unit exists", "name", config.SshServiceName)
	if err := serviceProvider.UnitExists(); err != nil {
		log.Fatal("unit for ssh does not exist: ", err)
	}

	sshConfigWrapper := &SshConfigWrapper{config.SshdConfigFile}
	ssh, err := NewSshAegis(sshConfigWrapper, statusSource, serviceProvider, config)
	if err != nil {
		log.Fatal("could not build app: ", err)
	}

	metricsWriter, err := buildMetricsWriter(config)
	if err != nil {
		log.Fatal("could not build metrics writer: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		<-sigc
		slog.Info("Received signal")
		cancel()
	}()

	run(ctx, ssh, metricsWriter)
}

func run(ctx context.Context, ssh *SshAegis, metricsWriter *MetricsWriter) {
	ssh.Check()
	t := time.NewTicker(15 * time.Second)

	silenceMetricsWriterWarnLogs := false

	for {
		select {
		case <-t.C:
			ssh.Check()
			if metricsWriter != nil {
				if err := metricsWriter.Dump(); err != nil && !silenceMetricsWriterWarnLogs {
					silenceMetricsWriterWarnLogs = true
					slog.Warn("can not write metrics data", "err", err)
				} else {
					silenceMetricsWriterWarnLogs = false
				}
			}
		case <-ctx.Done():
			t.Stop()
			slog.Info("Bye")
			return
		}
	}
}

func buildMetricsWriter(config *SshAegisConfig) (*MetricsWriter, error) {
	if config.MetricsFile == "" {
		return nil, nil
	}

	basePath := filepath.Dir(config.MetricsFile)
	_, err := os.Stat(basePath)

	if err != nil && os.IsNotExist(err) {
		isUsingDefaultValue := config.MetricsFile == configDefaultMetricsFile
		if isUsingDefaultValue {
			slog.Warn("Disabling metrics writer, path does not exist", "path", basePath)
		} else {
			return nil, fmt.Errorf("base path for writing metrics does not exist: %w", err)
		}
	}

	return NewMetricsWriter(config.MetricsFile)
}
