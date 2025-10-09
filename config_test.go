package main

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testFile = "./testFileJustForTesting"
var testFolder = "./testFolderJustForTesting"
var testSubFolder = testFolder + "/Folder2/Folder3"

func createTestFile(t *testing.T) {
	create, err := os.Create(testFile)
	if err != nil {
		t.Errorf("File not existing and unable to create.")
		return
	}
	err = create.Close()
	if err != nil {
		t.Errorf("Error closing new file.")
		return
	}
	if !fileExists(testFile) {
		t.Errorf("File not existing and unable to create.")
		return
	}
}

func readLogFile(t *testing.T) string {
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading test log file: %v\n", err)
	}
	return string(data)
}

func removeTestFolder(t *testing.T) {
	err := os.RemoveAll(testFolder)
	if err != nil {
		t.Errorf("Failed to remove folder '%s': %v", testFolder, err)
	}
}

func folderExists(t *testing.T, path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		t.Errorf("Error while checking folder '%s': %v", path, err)
		return false
	}
	if info.IsDir() {
		return true
	}
	t.Errorf("Found file instead of folder: %v", path)
	return false
}

func isFolderEmpty(t *testing.T) bool {
	dir, err := os.Open(testSubFolder)
	if err != nil {
		t.Errorf("Error opening folder '%s': %v", testSubFolder, err)
		return false
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			t.Errorf("Error closing folder '%s': %v", testSubFolder, err)
		}
	}(dir)
	files, err := dir.Readdirnames(-1)
	if err != nil {
		t.Errorf("Error listing content '%s': %v", testSubFolder, err)
		return false
	}
	return len(files) == 0
}

func TestConfig_commandExecution(t *testing.T) {
	type args struct {
		command Command
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{
			name: "Successful",
			args: args{command: Command{
				Command: "ls",
				Args:    nil,
				Late:    false,
				USB:     false,
				Ping:    false,
				Web:     false,
			}},
			want: ExecSuc,
		},
		{
			name: "Unsuccessful: Not found",
			args: args{command: Command{
				Command: "gehfwahkagkhgrbjavrhgkabjgrashjklbgahjslksglhbvklsd", // Command not found
				Args:    nil,
				Late:    false,
				USB:     false,
				Ping:    false,
				Web:     false,
			}},
			want: ExecErr,
		},
		{
			name: "Unsuccessful: empty",
			args: args{command: Command{
				Command: "",
				Args:    nil,
				Late:    false,
				USB:     false,
				Ping:    false,
				Web:     false,
			}},
			want: ExecErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.LogFile = ""
			if got := c.commandExecution(tt.args.command); got != tt.want {
				t.Errorf("commandExecution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_deleteFileIfExisting(t *testing.T) {
	type args struct {
		create bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Unclear state, delete file",
			args: args{create: true},
		},
		{
			name: "Create & delete file",
			args: args{create: true},
		},
		{
			name: "No file to delete",
			args: args{create: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.LogFile = ""
			if !fileExists(testFile) && tt.args.create {
				createTestFile(t)
			}
			c.deleteFileIfExisting(testFile)
			if fileExists(testFile) {
				t.Errorf("File not deleted")
			}
		})
	}
}

func TestConfig_exec(t *testing.T) {
	lsAllCommand := Command{Command: "ls", Args: nil, Late: false, USB: true, Ping: true, Web: true}
	lsUSBCommand := Command{Command: "ls", Args: nil, Late: false, USB: true, Ping: false, Web: false}
	lsPingCommand := Command{Command: "ls", Args: nil, Late: false, USB: false, Ping: true, Web: false}
	lsWebCommand := Command{Command: "ls", Args: nil, Late: false, USB: false, Ping: false, Web: true}
	lsLateUSBCommand := Command{Command: "ls", Args: nil, Late: true, USB: true, Ping: false, Web: false}
	lsLatePingCommand := Command{Command: "ls", Args: nil, Late: true, USB: false, Ping: true, Web: false}
	lsLateWebCommand := Command{Command: "ls", Args: nil, Late: true, USB: false, Ping: false, Web: true}
	type fields struct {
		FileLock        bool
		FileLockPresent bool
		Commands        []Command
	}
	type args struct {
		callee uint8
		noExec bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint8
		late   bool
	}{
		{
			name: "NoExec active, Ping, FileLock present",
			args: args{callee: CalleePing, noExec: true},
			fields: fields{
				FileLock:        true,
				FileLockPresent: true,
				Commands:        []Command{lsAllCommand}},
			want: NoExec,
			late: false,
		},
		{
			name:   "NoExec active, Web, FileLock not present",
			args:   args{callee: CalleeWeb, noExec: true},
			fields: fields{FileLock: true, FileLockPresent: false, Commands: []Command{lsAllCommand}},
			want:   NoExec,
			late:   false,
		},
		{
			name:   "NoExec active, USB",
			args:   args{callee: CalleeUSB, noExec: true},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsAllCommand}},
			want:   NoExec,
			late:   false,
		},
		{
			name:   "FileLock, present, USB",
			args:   args{callee: CalleeUSB, noExec: false},
			fields: fields{FileLock: true, FileLockPresent: true, Commands: []Command{lsAllCommand}},
			want:   FileLock,
			late:   false,
		},
		{
			name:   "FileLock, present, Web",
			args:   args{callee: CalleeWeb, noExec: false},
			fields: fields{FileLock: true, FileLockPresent: true, Commands: []Command{lsAllCommand}},
			want:   FileLock,
			late:   false,
		},
		{
			name:   "FileLock, present, Ping",
			args:   args{callee: CalleePing, noExec: false},
			fields: fields{FileLock: true, FileLockPresent: true, Commands: []Command{lsAllCommand}},
			want:   FileLock,
			late:   false,
		},
		{
			name:   "FileLock, not present",
			args:   args{callee: CalleeUSB, noExec: false},
			fields: fields{FileLock: true, FileLockPresent: false, Commands: []Command{lsAllCommand}},
			want:   ExecSuc,
			late:   false,
		},
		{
			name:   "no FileLock, but present",
			args:   args{callee: CalleePing, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: true, Commands: []Command{lsAllCommand}},
			want:   ExecSuc,
			late:   false,
		},
		{
			name:   "no FileLock, not present",
			args:   args{callee: CalleeWeb, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsAllCommand}},
			want:   ExecSuc,
			late:   false,
		},
		{
			name:   "Execution, USB",
			args:   args{callee: CalleeUSB, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsUSBCommand}},
			want:   ExecSuc,
			late:   false,
		},
		{
			name:   "Late Execution, USB",
			args:   args{callee: CalleeUSB, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsLateUSBCommand}},
			want:   ExecSuc,
			late:   true,
		},
		{
			name:   "Execution, Ping",
			args:   args{callee: CalleePing, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsPingCommand}},
			want:   ExecSuc,
			late:   false,
		},
		{
			name:   "Late Execution, Ping",
			args:   args{callee: CalleePing, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsLatePingCommand}},
			want:   ExecSuc,
			late:   true,
		},
		{
			name:   "Execution, Web",
			args:   args{callee: CalleeWeb, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsWebCommand}},
			want:   ExecSuc,
			late:   false,
		},
		{
			name:   "Late Execution, Web",
			args:   args{callee: CalleeWeb, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsLateWebCommand}},
			want:   ExecSuc,
			late:   true,
		},
		{
			name:   "No Execution, Wrong Callee: USB",
			args:   args{callee: CalleeUSB, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsWebCommand}},
			want:   NoExec,
			late:   false,
		},
		{
			name:   "No Execution, Wrong Callee: Ping",
			args:   args{callee: CalleePing, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsUSBCommand}},
			want:   NoExec,
			late:   false,
		},
		{
			name:   "No Execution, Wrong Callee: Web",
			args:   args{callee: CalleeWeb, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsPingCommand}},
			want:   NoExec,
		},
		{
			name:   "No Execution, Wrong Callee: USB, late",
			args:   args{callee: CalleeUSB, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsLatePingCommand}},
			want:   NoExec,
			late:   false,
		},
		{
			name:   "No Execution, Wrong Callee: Ping, late",
			args:   args{callee: CalleePing, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsLateWebCommand}},
			want:   NoExec,
			late:   false,
		},
		{
			name:   "No Execution, Wrong Callee: Web, late",
			args:   args{callee: CalleeWeb, noExec: false},
			fields: fields{FileLock: false, FileLockPresent: false, Commands: []Command{lsLateUSBCommand}},
			want:   NoExec,
			late:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.LogFile = ""
			c.FileLock = tt.fields.FileLock
			c.Commands = tt.fields.Commands
			c.FileLockPath = testFile
			if tt.fields.FileLockPresent {
				createTestFile(t)
			} else {
				c.deleteFileIfExisting(testFile)
			}
			got, late := c.exec(tt.args.callee, tt.args.noExec)
			if got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}
			if late != tt.late {
				t.Errorf("exec() late = %v, want %v", late, tt.late)
			}
		})
	}
}

func TestConfig_log(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name string
		args args
		file string
	}{
		{
			name: "No LogFile",
			args: args{message: "No LogFile"},
			file: "",
		},
		{
			name: "LogFile",
			args: args{message: "LogFile"},
			file: testFile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.LogFile = tt.file
			if len(tt.file) > 0 {
				c.deleteFileIfExisting(tt.file)
				c.log(tt.args.message)
				content := readLogFile(t)
				c.deleteFileIfExisting(tt.file)
				if !strings.Contains(content, tt.args.message) {
					t.Errorf("Got \"%v\" on stdout, want to include %v", content, tt.args.message)
				}
			} else {
				c.log(tt.args.message)
			}
		})
	}
}

func TestConfig_logErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Log Error",
			args: args{err: errors.New("TestError")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.LogFile = testFile
			c.deleteFileIfExisting(c.LogFile)
			c.logErr(tt.args.err)
			content := readLogFile(t)
			c.deleteFileIfExisting(c.LogFile)
			if !strings.Contains(content, tt.args.err.Error()) {
				t.Errorf("Got \"%v\" on stdout, want to include %v", content, tt.args.err.Error())
			}
		})
	}
}

func TestConfig_printAndLog(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Log to file",
			args: args{message: "Log to file"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.LogFile = testFile
			c.deleteFileIfExisting(c.LogFile)
			c.printAndLog(tt.args.message)
			content := readLogFile(t)
			c.deleteFileIfExisting(c.LogFile)
			if !strings.Contains(content, tt.args.message) {
				t.Errorf("Got \"%v\" on stdout, want to include %v", content, tt.args.message)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "NewConfig arrays not nil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(); got.PingTrackingConfigs == nil || got.WebTrackingConfigs == nil || got.Commands == nil {
				t.Errorf("NewConfig() = %v, want non nil arrays", got)
			}
		})
	}
}

func TestNewConfigFromFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "Load test config",
			args: args{filename: "./testing.yaml"},
			want: &Config{
				FileLock:         true,
				FileLockPath:     " ",
				FileLockDeletion: true,
				StartDelay:       1 * time.Hour,
				LogFile:          " ",
				OldLogs:          9,
				USBTracking:      true,
				USBInterval:      1 * time.Hour,
				IgnoredIDs:       []string{" "},
				PingTracking:     true,
				PingInterval:     1 * time.Hour,
				PingTrackingConfigs: []PingTarget{{
					Target:      " ",
					PingTimeout: 1 * time.Hour,
					OnSuccess:   true,
					RetryCount:  9,
					RetryDelay:  1 * time.Hour,
				}},
				WebTracking: true,
				WebInterval: 1 * time.Hour,
				WebTrackingConfigs: []WebTarget{{
					Target:          " ",
					Content:         " ",
					ContentIsExact:  true,
					StatusCode:      9,
					OnCodeIdentical: true,
					OnHTTPSFails:    true,
					RetryCount:      9,
					RetryDelay:      1 * time.Hour,
				}},
				Commands: []Command{{
					Command: " ",
					Args:    []string{" "},
					Late:    true,
					USB:     true,
					Ping:    true,
					Web:     true,
				}},
			},
			wantErr: false,
		},
		{
			name:    "Invalid filepath",
			args:    args{filename: "ghdshjgkldfslhgfsjdhudfhjkldsf"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigFromFile(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigFromFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consume(t *testing.T) {
	type args struct {
		u uint8
		v uint8
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{
			name: "u == NoExec",
			args: args{u: NoExec, v: ExecSuc},
			want: ExecSuc, // Sollte v zurückgeben, wenn u == NoExec
		},
		{
			name: "u == NoExec",
			args: args{u: NoExec, v: ExecErr},
			want: ExecErr, // Sollte v zurückgeben, wenn u == NoExec
		},
		{
			name: "u == NoExec",
			args: args{u: NoExec, v: NoExec},
			want: NoExec, // Sollte v zurückgeben, wenn u == NoExec
		},
		{
			name: "u == ExecErr",
			args: args{u: ExecErr, v: ExecSuc},
			want: ExecErr, // Sollte ExecErr zurückgeben, wenn u == ExecErr
		},
		{
			name: "u == ExecErr",
			args: args{u: ExecErr, v: ExecErr},
			want: ExecErr, // Sollte ExecErr zurückgeben, wenn u == ExecErr
		},
		{
			name: "u == ExecErr",
			args: args{u: ExecErr, v: NoExec},
			want: ExecErr, // Sollte ExecErr zurückgeben, wenn u == ExecErr
		},
		{
			name: "u == ExecSuc",
			args: args{u: ExecSuc, v: ExecErr},
			want: ExecErr, // Sollte ExecErr zurückgeben, wenn v == ExecErr
		},
		{
			name: "u == ExecSuc",
			args: args{u: ExecSuc, v: ExecSuc},
			want: ExecSuc, // Sollte ExecSuc zurückgeben, wenn v == ExecSuc
		},
		{
			name: "u == ExecSuc",
			args: args{u: ExecSuc, v: NoExec},
			want: ExecSuc, // Sollte ExecSuc zurückgeben, wenn v == NoExec
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := consume(tt.args.u, tt.args.v); got != tt.want {
				t.Errorf("consume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createPath(t *testing.T) {
	tests := []struct {
		name   string
		double bool
	}{
		{name: "Test folder creation", double: false},
		{name: "Test with existing folder", double: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//if folderExists(t, testFolder) {
			//	removeTestFolder(t)
			//}
			createPath(testSubFolder + "/")
			if tt.double {
				createPath(testSubFolder + "/")
			}
			if !folderExists(t, testSubFolder) {
				t.Errorf("Folder not created")
			}
			if !isFolderEmpty(t) {
				t.Errorf("Folder not empty")
			}
			removeTestFolder(t)
		})
	}
}

func Test_fileExists(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "Create test file and check", want: true},
		{name: "No test file", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want {
				createTestFile(t)
			}
			if got := fileExists(testFile); got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
			c := NewConfig()
			c.LogFile = ""
			c.deleteFileIfExisting(testFile)
		})
	}
}
