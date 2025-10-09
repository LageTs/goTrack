package main

import (
	"reflect"
	"testing"
)

func TestNewUSBTracker(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "Test creation with config",
			want: &Config{LogFile: "123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUSBTracker(tt.want); !reflect.DeepEqual(got.Config, tt.want) {
				t.Errorf("NewUSBTracker() = %v, want %v", got.Config, tt.want)
			}
		})
	}
}

func TestUSBDevice_getBusSum(t *testing.T) {
	var emptyBus = map[uint]uint{}
	var singleItemOneBus = map[uint]uint{}
	singleItemOneBus[1] = 1
	var singleItemMultipleBus = map[uint]uint{}
	singleItemMultipleBus[1] = 1
	singleItemMultipleBus[2] = 1
	singleItemMultipleBus[4] = 1
	var multipleItemMultipleBus = map[uint]uint{}
	multipleItemMultipleBus[1] = 1
	multipleItemMultipleBus[2] = 2
	multipleItemMultipleBus[5] = 3
	var multipleItemSingleBus = map[uint]uint{}
	multipleItemSingleBus[5] = 3

	type fields struct {
		ID       string
		Name     string
		BusCount map[uint]uint
	}
	tests := []struct {
		name   string
		fields fields
		want   uint
	}{
		{
			name: "Empty",
			fields: fields{
				ID:       "1234:5678",
				Name:     "Test",
				BusCount: emptyBus,
			},
			want: 0,
		},
		{
			name: "singleItemOneBus",
			fields: fields{
				ID:       "",
				Name:     "Test",
				BusCount: singleItemOneBus,
			},
			want: 1,
		},
		{
			name: "singleItemMultipleBus",
			fields: fields{
				ID:       "1234",
				Name:     " ",
				BusCount: singleItemMultipleBus,
			},
			want: 3,
		},
		{
			name: "multipleItemMultipleBus",
			fields: fields{
				ID:       "1234:5678",
				Name:     "",
				BusCount: multipleItemMultipleBus,
			},
			want: 6,
		},
		{
			name: "multipleItemSingleBus",
			fields: fields{
				ID:       "ABCD:5678",
				Name:     "Testing",
				BusCount: multipleItemSingleBus,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &USBDevice{
				ID:       tt.fields.ID,
				Name:     tt.fields.Name,
				BusCount: tt.fields.BusCount,
			}
			if got := u.getBusSum(); got != tt.want {
				t.Errorf("getBusSum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUSBDevice_isBusCountEqual(t *testing.T) {
	var emptyBus = map[uint]uint{}
	var singleItemOneBus = map[uint]uint{}
	singleItemOneBus[1] = 1
	var singleItemMultipleBus = map[uint]uint{}
	singleItemMultipleBus[1] = 1
	singleItemMultipleBus[2] = 1
	singleItemMultipleBus[4] = 1
	var multipleItemMultipleBus = map[uint]uint{}
	multipleItemMultipleBus[1] = 1
	multipleItemMultipleBus[2] = 2
	multipleItemMultipleBus[5] = 3
	type fields struct {
		ID       string
		Name     string
		BusCount map[uint]uint
	}
	tests := []struct {
		name    string
		fields  fields
		fields2 fields
		want    bool
	}{
		{
			name:    "Empty",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: emptyBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: emptyBus},
			want:    true,
		},
		{
			name:    "singleItemOneBus",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: singleItemOneBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: singleItemOneBus},
			want:    true,
		},
		{
			name:    "singleItemMultipleBus",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: singleItemMultipleBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: singleItemMultipleBus},
			want:    true,
		},
		{
			name:    "multipleItemMultipleBus",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: multipleItemMultipleBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: multipleItemMultipleBus},
			want:    true,
		},
		{
			name:    "Empty wrong",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: emptyBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: multipleItemMultipleBus},
			want:    false,
		},
		{
			name:    "singleItemOneBus wrong",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: singleItemOneBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: emptyBus},
			want:    false,
		},
		{
			name:    "singleItemMultipleBus wrong",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: singleItemMultipleBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: singleItemOneBus},
			want:    false,
		},
		{
			name:    "multipleItemMultipleBus wrong",
			fields:  fields{ID: "1234:5678", Name: "Test", BusCount: multipleItemMultipleBus},
			fields2: fields{ID: "1234:5678", Name: "Test", BusCount: singleItemMultipleBus},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &USBDevice{
				ID:       tt.fields.ID,
				Name:     tt.fields.Name,
				BusCount: tt.fields.BusCount,
			}
			u2 := USBDevice{
				ID:       tt.fields2.ID,
				Name:     tt.fields2.Name,
				BusCount: tt.fields2.BusCount,
			}
			if got := u.isBusCountEqual(u2); got != tt.want {
				t.Errorf("isBusCountEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUSBTracker_deviceIDExists(t *testing.T) {
	var singleDevice = map[string]USBDevice{}
	singleDevice["1234:5678"] = USBDevice{}
	var multipleDevice = map[string]USBDevice{}
	multipleDevice["1234:5678"] = USBDevice{}
	multipleDevice["ABCD:1234"] = USBDevice{}
	tests := []struct {
		name          string
		cachedDevices map[string]USBDevice
		id            string
		want          bool
	}{
		{
			name:          "Empty",
			cachedDevices: map[string]USBDevice{},
			id:            "1234:5678",
			want:          false,
		},
		{
			name:          "Single found",
			cachedDevices: singleDevice,
			id:            "1234:5678",
			want:          true,
		},
		{
			name:          "Single missing",
			cachedDevices: singleDevice,
			id:            "1234:1234",
			want:          false,
		},
		{
			name:          "Multiple found",
			cachedDevices: multipleDevice,
			id:            "1234:5678",
			want:          true,
		},
		{
			name:          "Multiple missing",
			cachedDevices: multipleDevice,
			id:            "5678:5678",
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &USBTracker{
				cachedDevices: tt.cachedDevices,
			}
			if got := u.deviceIDExists(tt.id); got != tt.want {
				t.Errorf("deviceIDExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUSBTracker_deviceIDIgnored(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		id     string
		want   bool
	}{
		{
			name:   "Empty",
			config: &Config{IgnoredIDs: []string{}},
			id:     "1234:5678",
			want:   false,
		},
		{
			name:   "Single, found",
			config: &Config{IgnoredIDs: []string{"1234:5678"}},
			id:     "1234:5678",
			want:   true,
		},
		{
			name:   "Single, not found",
			config: &Config{IgnoredIDs: []string{"1234:5678"}},
			id:     "1234:1234",
			want:   false,
		},
		{
			name:   "Multiple, found",
			config: &Config{IgnoredIDs: []string{"1234:5678", "ABCD:1234"}},
			id:     "1234:5678",
			want:   true,
		},
		{
			name:   "Multiple, not found",
			config: &Config{IgnoredIDs: []string{"1234:5678", "ABCD:1234"}},
			id:     "1234:1234",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &USBTracker{
				Config: tt.config,
			}
			if got := u.deviceIDIgnored(tt.id); got != tt.want {
				t.Errorf("deviceIDIgnored() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_combineFields(t *testing.T) {
	type args struct {
		array      []string
		startIndex int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty array",
			args: args{array: []string{}, startIndex: 0},
			want: "",
		},
		{
			name: "Empty string",
			args: args{array: []string{""}, startIndex: 0},
			want: "",
		},
		{
			name: "Test full",
			args: args{array: []string{"T", "es", "t"}, startIndex: 0},
			want: "T es t",
		},
		{
			name: "Test partial",
			args: args{array: []string{"Th", "is", " ", "is a", "T", "es", "t"}, startIndex: 4},
			want: "T es t",
		},
		{
			name: "Test partial leading space",
			args: args{array: []string{"Th", "is", " ", "is a", "  T", "es", "t"}, startIndex: 4},
			want: "T es t",
		},
		{
			name: "Empty high start",
			args: args{array: []string{}, startIndex: 3},
			want: "",
		},
		{
			name: "Test only spaces left",
			args: args{array: []string{"T", "es", "t", "", "   ", " ", " "}, startIndex: 3},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := combineFields(tt.args.array, tt.args.startIndex); got != tt.want {
				t.Errorf("combineFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deviceIDMissing(t *testing.T) {
	var empty = map[string]USBDevice{}
	var singleDevice = map[string]USBDevice{}
	singleDevice["1234:5678"] = USBDevice{}
	var multipleDevice = map[string]USBDevice{}
	multipleDevice["1234:5678"] = USBDevice{}
	multipleDevice["ABCD:1234"] = USBDevice{}
	type args struct {
		current map[string]USBDevice
		id      string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Empty",
			args: args{current: empty, id: "1234:5678"},
			want: true,
		},
		{
			name: "Single, found",
			args: args{current: singleDevice, id: "1234:5678"},
			want: false,
		},
		{
			name: "Single, not found",
			args: args{current: singleDevice, id: "1234:1234"},
			want: true,
		},
		{
			name: "Multiple, found",
			args: args{current: multipleDevice, id: "1234:5678"},
			want: false,
		},
		{
			name: "Multiple, not found",
			args: args{current: multipleDevice, id: "1234:1234"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deviceIDMissing(tt.args.current, tt.args.id); got != tt.want {
				t.Errorf("deviceIDMissing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_has(t *testing.T) {
	var empty []string
	var single = []string{"1234:5678"}
	var multiple = []string{"1234:5678", "ABCD:1234"}
	type args struct {
		array []string
		id    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Empty",
			args: args{array: empty, id: "1234:5678"},
			want: false,
		},
		{
			name: "Single, found",
			args: args{array: single, id: "1234:5678"},
			want: true,
		},
		{
			name: "Single, not found",
			args: args{array: single, id: "1234:1234"},
			want: false,
		},
		{
			name: "Multiple, found",
			args: args{array: multiple, id: "1234:5678"},
			want: true,
		},
		{
			name: "Multiple, not found",
			args: args{array: multiple, id: "1234:1234"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := has(tt.args.array, tt.args.id); got != tt.want {
				t.Errorf("has() = %v, want %v", got, tt.want)
			}
		})
	}
}
