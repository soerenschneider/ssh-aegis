package main

import (
	"reflect"
	"testing"
)

func Test_getListenAddressIndices(t *testing.T) {
	type args struct {
		lines []string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "happy case",
			args: args{
				lines: []string{
					"Some value",
					"ListenAddress 1.2.3.4",
					"Some value",
					"Some value",
					"ListenAddress 4.3.2.1",
					"Some value",
					"Some value",
					"Some value",
					"ListenAddress ::",
				},
			},
			want: []int{1, 4, 8},
		},
		{
			name: "no listen address",
			args: args{
				lines: []string{
					"Some value",
					"Some value",
					"Some value",
					"Some value",
					"Some value",
					"Some value",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getListenAddressIndices(tt.args.lines); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getListenAddressIndices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getConfiguredListenAddresses(t *testing.T) {
	type args struct {
		data []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "happy case",
			args: args{
				data: []string{
					"Some value",
					"ListenAddress 1.2.3.4",
					"Some value",
					"Some value",
					"ListenAddress 4.3.2.1",
					"Some value",
					"Some value",
					"Some value",
					"ListenAddress ::",
				},
			},
			want: []string{
				"1.2.3.4",
				"4.3.2.1",
				"::",
			},
		},
		{
			name: "no listen addresses",
			args: args{
				data: []string{
					"Some value",
					"Some value",
					"Some value",
					"Some value",
					"Some value",
					"Some value",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getConfiguredListenAddresses(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfiguredListenAddresses() = %v, want %v", got, tt.want)
			}
		})
	}
}

type dummyConfigWrapper struct {
	config []string
}

func (d *dummyConfigWrapper) GetConfig() ([]string, error) {
	return d.config, nil
}

func (d *dummyConfigWrapper) WriteConfig(data []string) error {
	d.config = data
	return nil
}

func TestSshAegis_setConfiguredListenAddresses(t *testing.T) {
	type fields struct {
		configWrapper        ConfigWrapper
		tunnelStatusSource   TunnelStatusSource
		serviceProvider      ServiceReloader
		addressConfiguration map[TunnelStatus][]string
		oldStatus            TunnelStatus
	}
	type args struct {
		wanted []string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantConfig []string
	}{
		{
			name: "replace listen addresses, append at beginning",
			fields: fields{
				configWrapper: &dummyConfigWrapper{
					config: []string{
						"ListenAddress 1.2.3.4",
						"Unrelated",
						"ListenAddress 4.3.2.1",
						"Other option",
					},
				},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wanted: []string{
					"9.9.9.9",
					"1.2.3.4",
				},
			},
			wantErr: false,
			wantConfig: []string{
				"ListenAddress 9.9.9.9",
				"ListenAddress 1.2.3.4",
				"Unrelated",
				"Other option",
			},
		},
		{
			name: "replace listen addresses, append at end",
			fields: fields{
				configWrapper: &dummyConfigWrapper{
					config: []string{
						"Unrelated",
						"Other option",
						"ListenAddress 1.2.3.4",
						"ListenAddress 4.3.2.1",
					},
				},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wanted: []string{
					"9.9.9.9",
					"1.2.3.4",
				},
			},
			wantErr: false,
			wantConfig: []string{
				"Unrelated",
				"Other option",
				"ListenAddress 9.9.9.9",
				"ListenAddress 1.2.3.4",
			},
		},
		{
			name: "add listen addresses",
			fields: fields{
				configWrapper: &dummyConfigWrapper{
					config: []string{
						"Unrelated",
						"Other option",
					},
				},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wanted: []string{
					"1.2.3.4",
				},
			},
			wantErr: false,
			wantConfig: []string{
				"ListenAddress 1.2.3.4",
				"Unrelated",
				"Other option",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SshAegis{
				configWrapper:        tt.fields.configWrapper,
				tunnelStatusSource:   tt.fields.tunnelStatusSource,
				serviceProvider:      tt.fields.serviceProvider,
				addressConfiguration: tt.fields.addressConfiguration,
				oldStatus:            tt.fields.oldStatus,
			}
			if err := s.setConfiguredListenAddresses(tt.args.wanted); (err != nil) != tt.wantErr {
				t.Errorf("setConfiguredListenAddresses() error = %v, wantErr %v", err, tt.wantErr)
			}

			config, err := s.configWrapper.GetConfig()
			if err != nil {
				t.Errorf("unexpected error while reading config")
			}

			if !reflect.DeepEqual(config, tt.wantConfig) {
				t.Errorf("setConfiguredListenAddresses() config = %v, wantConfig %v", config, tt.wantConfig)
			}
		})
	}
}

func TestSshAegis_isUpdateNeeded(t *testing.T) {
	type fields struct {
		configWrapper        ConfigWrapper
		tunnelStatusSource   TunnelStatusSource
		serviceProvider      ServiceReloader
		addressConfiguration map[TunnelStatus][]string
		oldStatus            TunnelStatus
	}
	type args struct {
		wantedListenAddresses []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "nothing to do, no wanted addresses",
			fields: fields{
				configWrapper:        &dummyConfigWrapper{config: []string{""}},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wantedListenAddresses: nil,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "needs adding, no listenaddress configured",
			fields: fields{
				configWrapper: &dummyConfigWrapper{config: []string{
					"Some option",
					"Other option",
				}},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wantedListenAddresses: []string{"1.2.3.4"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "needs adding, 1 different listenaddress configured",
			fields: fields{
				configWrapper: &dummyConfigWrapper{config: []string{
					"Some option",
					"Other option",
					"ListenAddress 4.3.2.1",
				}},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wantedListenAddresses: []string{"1.2.3.4"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no adding needed, listenaddress already configured",
			fields: fields{
				configWrapper: &dummyConfigWrapper{config: []string{
					"Some option",
					"Other option",
					"ListenAddress 1.2.3.4",
				}},
				tunnelStatusSource:   nil,
				serviceProvider:      nil,
				addressConfiguration: nil,
				oldStatus:            0,
			},
			args: args{
				wantedListenAddresses: []string{"1.2.3.4"},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SshAegis{
				configWrapper:        tt.fields.configWrapper,
				tunnelStatusSource:   tt.fields.tunnelStatusSource,
				serviceProvider:      tt.fields.serviceProvider,
				addressConfiguration: tt.fields.addressConfiguration,
				oldStatus:            tt.fields.oldStatus,
			}
			got, err := s.isUpdateNeeded(tt.args.wantedListenAddresses)
			if (err != nil) != tt.wantErr {
				t.Errorf("isUpdateNeeded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isUpdateNeeded() got = %v, want %v", got, tt.want)
			}
		})
	}
}
