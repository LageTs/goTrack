package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
)

const CalleeUSB uint8 = 1

type Command struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Late    bool     `yaml:"late"`
	USB     bool     `yaml:"usb"`
}

// Config represents the configuration for the USB tracker
type Config struct {
	FileLock     bool      `yaml:"fileLock"`
	FileLockPath string    `yaml:"fileLockPath"`
	StartDelay   int       `yaml:"startDelay"`
	Interval     int       `yaml:"interval"`
	LogFile      string    `yaml:"logFile"`
	OldLogs      int       `yaml:"oldLogs"`
	IgnoredIDs   []string  `yaml:"ignoredIDs"`
	USBTracking  bool      `yaml:"usb_tracking"`
	Commands     []Command `yaml:"commands"`
}

// NewConfig Constructor for Config
func NewConfig() *Config {
	commands := []Command{{
		Command: "shutdown",
		Args:    []string{"0"},
		Late:    true,
		USB:     true,
	}}

	return &Config{
		FileLock:     false,
		FileLockPath: "",
		StartDelay:   0,
		Interval:     1000,
		LogFile:      "",
		OldLogs:      1,
		IgnoredIDs:   nil,
		USBTracking:  true,
		Commands:     commands,
	}
}

// NewConfigFromFile creates a new configuration from a JSON file
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

// exec executes all commands that are enabled for the callee
func (c Config) exec(callee uint8) {
	if !c.FileLock || c.FileLock && fileExists(c.FileLockPath) {
		var lateCommands []Command
		for _, command := range c.Commands {
			if command.USB && callee == CalleeUSB {
				if command.Late {
					lateCommands = append(lateCommands, command)
					continue
				}
				output, err := exec.Command(command.Command, command.Args...).Output()
				fmt.Println(string(output))
				fmt.Println(err)
			}
		}
		for _, command := range lateCommands {
			output, err := exec.Command(command.Command, command.Args...).Output()
			fmt.Println(string(output))
			fmt.Println(err)
		}
	} else {
		fmt.Println("Execution skipped as file lock is activated but not present")
	}
}

func (c Config) logErr(err error) {
	c.log(err.Error())
}

// log handles logging with log file path from config
func (c Config) log(message string) {
	if len(c.LogFile) > 0 {
		// Log into file
		file, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			println("Could not open log file: " + c.LogFile)
			println("Error: ", err.Error())
		} else {
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					println("Could not close log file: " + c.LogFile)
					println("Error: ", err.Error())
				}
			}(file)

			// Write log file
			_, err = file.WriteString(message + "\n")
			if err != nil {
				println("Could not write to log file: " + c.LogFile)
				println("Error: ", err.Error())
			}
		}
	} else {
		println(message)
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		print("File error", err)
		os.Exit(-1)
	}
	return true
}

func createEmptyFileIfMissing(path string) bool {
	if fileExists(path) {
		return true
	}
	file, err := os.Create(path)
	if err != nil {
		print("File error: ", err)
		os.Exit(-2)
	}
	err = file.Close()
	if err != nil {
		fmt.Println("File error: ", err)
		os.Exit(-3)
	}
	return true
}
