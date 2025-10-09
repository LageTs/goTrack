package main

import (
	"reflect"
	"testing"
	"time"
)

func TestNewPingTracker(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name string
		args args
		want *PingTracker
	}{
		{
			name: "simple config",
			args: args{
				config: &Config{
					FileLock:   true,
					StartDelay: 2 * time.Second,
				},
			},
			want: &PingTracker{
				Config: &Config{
					FileLock:   true,
					StartDelay: 2 * time.Second,
				},
			},
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: &PingTracker{
				Config: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPingTracker(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPingTracker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPingTracker_TrackPingTargets(t *testing.T) {
	timeout := 0 * time.Second
	delay := 30 * time.Second
	type fields struct {
		Config *Config
	}
	tests := []struct {
		name     string
		fields   fields
		expected uint
	}{
		{
			name: "no ping targets",
			fields: fields{
				Config: &Config{PingTrackingConfigs: []PingTarget{}},
			},
			expected: 0,
		},
		{
			name: "single target with noExec",
			fields: fields{
				Config: &Config{
					PingTrackingConfigs: []PingTarget{
						{Target: "127.0.0.1", PingTimeout: timeout, RetryCount: 0, RetryDelay: delay, OnSuccess: true},
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple targets normal run",
			fields: fields{
				Config: &Config{
					PingTrackingConfigs: []PingTarget{
						{Target: "127.0.0.1", PingTimeout: timeout, RetryCount: 0, RetryDelay: delay, OnSuccess: false},
						{Target: "localhost", PingTimeout: timeout, RetryCount: 0, RetryDelay: delay, OnSuccess: true},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PingTracker{Config: tt.fields.Config}
			res := p.TrackPingTargets(true, false)
			if res != tt.expected {
				t.Errorf("TrackPingTargets() = %d, expected %d", res, tt.expected)
			}
		})
	}
}

func TestPingTracker_ping(t *testing.T) {
	var timeout = 1 * time.Second
	var delay = 1 * time.Second
	type fields struct {
		Config *Config
	}
	type args struct {
		debug       bool
		pingTargets *PingTarget
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected uint8
	}{
		{
			name:   "successful ping",
			fields: fields{Config: &Config{}},
			args: args{
				debug:       true,
				pingTargets: &PingTarget{Target: "127.0.0.1", PingTimeout: timeout, OnSuccess: false, RetryCount: 0, RetryDelay: 0},
			},
			expected: PingSuc,
		},
		{
			name:   "successful ping; OnSuccess",
			fields: fields{Config: &Config{}},
			args: args{
				debug:       false,
				pingTargets: &PingTarget{Target: "127.0.0.1", PingTimeout: timeout, OnSuccess: true, RetryCount: 1, RetryDelay: 0},
			},
			expected: PingExec,
		},
		{
			name:   "failed ping with retries",
			fields: fields{Config: &Config{}},
			args: args{
				debug:       true,
				pingTargets: &PingTarget{Target: "192.0.2.1", PingTimeout: timeout, OnSuccess: false, RetryCount: 2, RetryDelay: delay},
			},
			expected: PingExec,
		},
		{
			name:   "failed ping with retries; OnSuccess",
			fields: fields{Config: &Config{}},
			args: args{
				debug:       true,
				pingTargets: &PingTarget{Target: "192.0.2.1", PingTimeout: timeout, OnSuccess: true, RetryCount: 5, RetryDelay: delay},
			},
			expected: PingNoSuc,
		},
		{
			name:   "Invalid PingTimeout",
			fields: fields{Config: &Config{}},
			args: args{
				debug:       false,
				pingTargets: &PingTarget{Target: "127.0.0.1", PingTimeout: 0, OnSuccess: true, RetryCount: 1, RetryDelay: 0},
			},
			expected: PingTimeoutErr,
		},
		{
			name:   "Invalid Target", // Will be converted to nil which will be interpreted as localhost
			fields: fields{Config: &Config{}},
			args: args{
				debug:       false,
				pingTargets: &PingTarget{Target: "999.0.0.1", PingTimeout: timeout, OnSuccess: false, RetryCount: 1, RetryDelay: 0},
			},
			expected: PingSuc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PingTracker{Config: tt.fields.Config}
			// darf nicht crashen
			res := p.ping(true, tt.args.debug, tt.args.pingTargets)
			if res != tt.expected {
				t.Errorf("Ping() = %d, expected %d", res, tt.expected)
			}
		})
	}
}
