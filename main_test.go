package main

import (
	"testing"
)

func Test_buildMetricsWriter(t *testing.T) {
	type args struct {
		config *SshAegisConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *MetricsWriter
		wantErr bool
	}{
		{
			name: "no metricswriter wanted",
			args: args{
				config: &SshAegisConfig{
					MetricsFile: "",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "metricswriter wanted but invalid path",
			args: args{
				config: &SshAegisConfig{
					MetricsFile: "/nonexistent/metrics.prom",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "metricswriter wanted but invalid path",
			args: args{
				config: &SshAegisConfig{
					MetricsFile: configDefaultMetricsFile,
				},
			},
			want: &MetricsWriter{
				metricsFile: configDefaultMetricsFile,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildMetricsWriter(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildMetricsWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotNil := got == nil
			wantNil := tt.want == nil
			if gotNil != wantNil {
				t.Errorf("buildMetricsWriter() got = %v, want %v", got, tt.want)
			}
		})
	}
}
