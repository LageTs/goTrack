package main

import (
	"reflect"
	"testing"
	"time"
)

func TestWebTracker_TrackWebSources(t *testing.T) {
	timeout := 30 * time.Second
	delay := 3 * time.Second
	type fields struct {
		Config *Config
	}
	type args struct {
		noExec bool
		debug  bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "no web targets",
			fields: fields{
				Config: &Config{PingTrackingConfigs: []PingTarget{}},
			},
		},
		{
			name: "single web target with noExec",
			fields: fields{
				Config: &Config{PingTrackingConfigs: []PingTarget{
					{Target: "https://networkcheck.kde.org/", PingTimeout: timeout, OnSuccess: false, RetryCount: 1, RetryDelay: delay},
				}},
			},
		},
		{
			name: "multiple web target with noExec",
			fields: fields{
				Config: &Config{PingTrackingConfigs: []PingTarget{
					{Target: "https://networkcheck.kde.org/", PingTimeout: timeout, OnSuccess: false, RetryCount: 1, RetryDelay: delay},
					{Target: "https://kde.org/", PingTimeout: timeout, OnSuccess: true, RetryCount: 0, RetryDelay: delay},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WebTracker{
				Config: tt.fields.Config,
			}
			res := w.TrackWebSources(true, false)
			if res != uint(len(tt.fields.Config.WebTrackingConfigs)) {
				t.Errorf("WebTracker.TrackWebSources() = %v, want %v", res, len(tt.fields.Config.WebTrackingConfigs))
			}
		})
	}
}

func TestWebTracker_curl(t *testing.T) {
	type fields struct {
		Config *Config
	}
	type args struct {
		config WebTarget
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   webCurl
	}{
		{
			name:   "Unreachable",
			fields: fields{Config: &Config{}},
			args: args{WebTarget{
				Target:     "https://dfgszukkdbfzhyfkbjdhvzy",
				RetryCount: 0,
				RetryDelay: 0,
			}},
			want: webCurl{
				status:  -1,
				content: "",
				isHttps: false,
			},
		},
		{
			name:   "Reachable https",
			fields: fields{Config: &Config{}},
			args: args{WebTarget{
				Target:     "https://networkcheck.kde.org/",
				RetryCount: 1,
				RetryDelay: 3 * time.Second,
			}},
			want: webCurl{
				status:  200,
				content: "OK",
				isHttps: true,
			},
		},
		{
			name:   "Reachable http",
			fields: fields{Config: &Config{}},
			args: args{WebTarget{
				Target:     "http://networkcheck.kde.org/",
				RetryCount: 1,
				RetryDelay: 3 * time.Second,
			}},
			want: webCurl{
				status:  200,
				content: "OK",
				isHttps: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WebTracker{
				Config: tt.fields.Config,
			}
			if got := w.curl(false, tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("curl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebTracker_trackWebSource(t *testing.T) {
	config := Config{LogFile: ""}
	delay := 3 * time.Second
	type args struct {
		config WebTarget
	}
	tests := []struct {
		name  string
		args  args
		wants uint8
	}{
		{
			name: "Status code Identical",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "",
				ContentIsExact:  false,
				StatusCode:      200,
				OnCodeIdentical: true,
				OnHTTPSFails:    false,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebCode,
		},
		{
			name: "Status code differs",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "",
				ContentIsExact:  false,
				StatusCode:      500,
				OnCodeIdentical: false,
				OnHTTPSFails:    false,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebNotCode,
		},
		{
			name: "https fails",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "",
				ContentIsExact:  false,
				StatusCode:      0,
				OnCodeIdentical: false,
				OnHTTPSFails:    true,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebHTTPS,
		},
		{
			name: "Content is Exact",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "OK\n",
				ContentIsExact:  true,
				StatusCode:      0,
				OnCodeIdentical: false,
				OnHTTPSFails:    false,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebContExact,
		},
		{
			name: "Content contains",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "OK",
				ContentIsExact:  false,
				StatusCode:      0,
				OnCodeIdentical: false,
				OnHTTPSFails:    false,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebCont,
		},
		{
			name: "No execution",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "OK",
				ContentIsExact:  true,
				StatusCode:      500,
				OnCodeIdentical: true,
				OnHTTPSFails:    true,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebNoExec,
		},
		{
			name: "Multiple reasons",
			args: args{config: WebTarget{
				Target:          "https://networkcheck.kde.org/",
				Content:         "OK",
				ContentIsExact:  false,
				StatusCode:      500,
				OnCodeIdentical: false,
				OnHTTPSFails:    false,
				RetryCount:      1,
				RetryDelay:      delay,
			}},
			wants: WebCont,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WebTracker{
				Config: &config,
			}
			w.trackWebSource(true, false, tt.args.config)
		})
	}
}
