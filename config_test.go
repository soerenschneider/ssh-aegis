package main

import "testing"

var (
	testValidAddressesMixed []string = []string{"0.0.0.0", "::"}
	testValidAddressIpv4    []string = []string{"0.0.0.0"}
	testValidAddressIpv6    []string = []string{"::"}
	testInvalidAddressIpv4  []string = []string{"127.0.0.0.1"}
)

const validSshConfigFile = "contrib/configs/sshd_config1.txt"

func TestSshAegisConfig_Validate(t *testing.T) {
	type fields struct {
		ListenAddressesUp      []string
		ListenAddressesDown    []string
		ListenAddressesUnknown []string
		SshdConfigFile         string
		WireguardInterface     string
		SshServiceName         string
		MetricsFile            string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "happy case, no unknown address",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    testValidAddressIpv6,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: false,
		},
		{
			name: "happy case, unknown address",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    testValidAddressIpv6,
				ListenAddressesUnknown: testValidAddressesMixed,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: false,
		},
		{
			name: "empty down address",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    nil,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
		{
			name: "empty up address",
			fields: fields{
				ListenAddressesUp:      nil,
				ListenAddressesDown:    testValidAddressIpv4,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
		{
			name: "invalid unknown address",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv6,
				ListenAddressesDown:    testValidAddressIpv4,
				ListenAddressesUnknown: testInvalidAddressIpv4,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
		{
			name: "using same addresses for up and down states",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    testValidAddressIpv4,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
		{
			name: "non-existing sshd config",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    testValidAddressIpv6,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         "nonexistent",
				WireguardInterface:     "wg0",
				SshServiceName:         "ssh",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
		{
			name: "empty ssh service name",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    testValidAddressIpv6,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "wg0",
				SshServiceName:         "",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
		{
			name: "empty wg interface name",
			fields: fields{
				ListenAddressesUp:      testValidAddressIpv4,
				ListenAddressesDown:    testValidAddressIpv6,
				ListenAddressesUnknown: nil,
				SshdConfigFile:         validSshConfigFile,
				WireguardInterface:     "",
				SshServiceName:         "sshd",
				MetricsFile:            "contrib/configs/test.prom",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SshAegisConfig{
				ListenAddressesUp:      tt.fields.ListenAddressesUp,
				ListenAddressesDown:    tt.fields.ListenAddressesDown,
				ListenAddressesUnknown: tt.fields.ListenAddressesUnknown,
				SshdConfigFile:         tt.fields.SshdConfigFile,
				WireguardInterface:     tt.fields.WireguardInterface,
				SshServiceName:         tt.fields.SshServiceName,
				MetricsFile:            tt.fields.MetricsFile,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
