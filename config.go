package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const CalleeUSB uint8 = 1
const CalleePing uint8 = 2
const CalleeWeb uint8 = 3

// WebConfig represents the configuration struct for web content to be tracked. TODO: Not introduced in default config
type WebConfig struct {
	Target string `yaml:"target"`
	// Content and ContentIsExact will be ignored if this is the empty string.
	Content string `yaml:"content"`
	// If ContentIsExact is true the configured commands will be executed if the received content is exactly Content. If false they will be executed if Content is a substring.
	ContentIsExact bool `yaml:"content_is_exact"`
	// StatusCode and OnCodeIdentical will be ignored if this is 0.
	StatusCode int `yaml:"status_code"`
	// If OnCodeIdentical is true the configured commands will be executed if the received HTTP Status Code is StatusCode. If false they will be executed if the Status Code differs.
	OnCodeIdentical bool `yaml:"on_code_identical"`
	// If OnHTTPSFails is true the configured commands will be executed if the HTTPS connection can not be established.
	OnHTTPSFails bool `yaml:"on_https_fails"`
	// If OnHTTPSFails is false RetryCount defines the number of retries before command execution.
	RetryCount int `yaml:"retry_count"`
	// If OnHTTPSFails is false RetryDelay defines the time in milliseconds to wait between two tries to curl.
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// PingConfig represents the configuration struct for pings to be tracked. TODO: Not introduced in default config
type PingConfig struct {
	Target      string        `yaml:"target"`
	PingTimeout time.Duration `yaml:"ping_timeout"`
	// If OnSuccess is true the configured commands will be executed if the ping is successfully received back. If false they will be executed if ping fails.
	OnSuccess bool `yaml:"on_success"`
	// If OnSuccess is false RetryCount defines the number of retries before command execution.
	RetryCount int `yaml:"retry_count"`
	// If OnSuccess is false RetryDelay defines the time in milliseconds to wait between two tries to ping.
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// Command represents the configuration struct for commands to be executed
type Command struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Late    bool     `yaml:"late"`
	// Is this command executed on USB activation?
	USB bool `yaml:"usb"`
	// Is this command executed on Ping activation?
	Ping bool `yaml:"ping"`
	// Is this command executed on Web activation?
	Web bool `yaml:"web"`
}

// Config represents the configuration for goTrack
type Config struct {
	FileLock            bool          `yaml:"fileLock"`
	FileLockPath        string        `yaml:"fileLockPath"`
	StartDelay          time.Duration `yaml:"startDelay"`
	Interval            time.Duration `yaml:"interval"`
	LogFile             string        `yaml:"logFile"`
	OldLogs             int           `yaml:"oldLogs"`
	IgnoredIDs          []string      `yaml:"ignoredIDs"`
	USBTracking         bool          `yaml:"usb_tracking"`
	PingTracking        bool          `yaml:"ping_tracking"`
	WebTracking         bool          `yaml:"web_tracking"`
	Commands            []Command     `yaml:"commands"`
	PingTrackingConfigs []PingConfig  `yaml:"pingConfig"`
	WebTrackingConfigs  []WebConfig   `yaml:"webConfig"`
}

// NewConfig Constructor for Config
func NewConfig() *Config {
	commands := []Command{{}}
	pingTrackingConfig := []PingConfig{{}}
	webTrackingConfig := []WebConfig{{}}

	return &Config{
		FileLock:            false,
		FileLockPath:        "",
		StartDelay:          0 * time.Second,
		Interval:            1000 * time.Millisecond,
		LogFile:             "",
		OldLogs:             1,
		IgnoredIDs:          nil,
		USBTracking:         true,
		PingTracking:        false,
		WebTracking:         false,
		Commands:            commands,
		PingTrackingConfigs: pingTrackingConfig,
		WebTrackingConfigs:  webTrackingConfig,
	}
}

// NewConfigFromFile creates a new configuration from a yaml file
func NewConfigFromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	config := &Config{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

// commandExecution runs any given command without any validation
func (c Config) commandExecution(command Command) {
	output, err := exec.Command(command.Command, command.Args...).Output()
	c.log(string(output))
	if err != nil {
		c.logErr(err)
	}
}

// exec executes all commands that are enabled for the callee
func (c Config) exec(callee uint8, noExec bool) {
	// If noExec is set nothing will be executed
	if noExec {
		c.log("Execution aborted due to \"NoExec\"")
	} else if !c.FileLock || c.FileLock && fileExists(c.FileLockPath) {
		// Execution will be started
		// lateCommands holds all commands that shall be executed after others
		var lateCommands []Command
		for _, command := range c.Commands {
			if command.USB && callee == CalleeUSB {
				if command.Late {
					lateCommands = append(lateCommands, command)
					continue
				}
				c.commandExecution(command)

			} else if command.Ping && callee == CalleePing {
				if command.Late {
					lateCommands = append(lateCommands, command)
					continue
				}
				c.commandExecution(command)

			} else if command.Web && callee == CalleeWeb {
				if command.Late {
					lateCommands = append(lateCommands, command)
					continue
				}
				c.commandExecution(command)
			}

		}
		// Execute late commands
		for _, command := range lateCommands {
			c.commandExecution(command)
		}
	} else {
		c.log("Execution skipped as file lock is activated but not present")
	}
}

// logErr is a little bit shorter and can be adapted in future
func (c Config) logErr(err error) {
	c.log(err.Error())
}

// printAndLog is a function to print the log message to stdout and also to log normally. Returns false if only stdout is used
func (c Config) printAndLog(message string) bool {
	if len(c.LogFile) > 0 {
		// Print and log
		println(message)
		c.log(message)
		return true
	} else {
		// Only print
		println(message)
		return false
	}
}

// log handles logging with log file path from config
func (c Config) log(message string) {
	if len(c.LogFile) > 0 {
		// Log into file
		if !fileExists(c.LogFile) {
			createPath(c.LogFile)
		}
		file, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		// Handle file system errors
		if err != nil {
			println("Could not open log file: " + c.LogFile)
			println("Error: ", err.Error())
		} else {
			// file is present -> start logging
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					println("Could not close log file: " + c.LogFile)
					println("Error: ", err.Error())
				}
			}(file)

			// Write log file
			timeString := "[" + time.Now().Format("15:04:05.00") + "]" // hh:mm:ss,ss
			_, err = file.WriteString(timeString + " " + message + "\n")
			if err != nil {
				println("Could not write to log file: " + c.LogFile)
				println("Error: ", err.Error())
			}
		}
	} else {
		println(message)
	}
}

// fileExists checks if a file is present in file system
func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		println("File error", err)
		os.Exit(-1)
	}
	return true
}

// createEmptyFileIfMissing creates an empty file is the path does not exist else does nothing.
func createEmptyFileIfMissing(path string) {
	if !fileExists(path) {
		createPath(path)
		file, err := os.Create(path)
		if err != nil {
			println("File error: ", err)
			os.Exit(-2)
		}
		err = file.Close()
		if err != nil {
			println("File error: ", err)
			os.Exit(-3)
		}
	}
}

// createPath creates als dirs that are part of the given path
func createPath(path string) {
	err := os.MkdirAll(filepath.Dir(path), 0600)
	if err != nil {
		println("File error: ", err)
	}
}
