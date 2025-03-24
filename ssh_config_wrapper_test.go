package main

import (
	"os"
	"reflect"
	"testing"
)

func TestSshConfigWrapper_GetConfig(t *testing.T) {
	type fields struct {
		sshConfigFile string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "happy case",
			fields: fields{
				sshConfigFile: "contrib/configs/sshd_config1.txt",
			},
			want: []string{
				"Some option",
				"ListenAddress 1.2.3.4",
				"Some other option",
			},
			wantErr: false,
		},
		{
			name: "non existent config",
			fields: fields{
				sshConfigFile: "contrib/configs/non-existent",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SshConfigWrapper{
				sshConfigFile: tt.fields.sshConfigFile,
			}
			got, err := s.GetConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSshConfigWrapper_WriteConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Errorf("could not create tmp file: %v", err)
	}

	defer os.Remove(tmpFile.Name())

	s := &SshConfigWrapper{
		sshConfigFile: tmpFile.Name(),
	}

	data := []string{
		"Line 1",
		"Line 2",
		"Line 3",
	}

	if err := s.WriteConfig(data); err != nil {
		t.Errorf("WriteConfig() error = %v", err)
	}

	read, err := s.GetConfig()
	if err != nil {
		t.Errorf("reading config back from GetConfig() error = %v", err)
	}

	if !reflect.DeepEqual(read, data) {
		t.Errorf("written data does not match expected data: written=%v, read=%v", data, read)
	}
}
