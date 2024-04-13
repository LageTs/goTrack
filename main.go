package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

const defaultConfigPath = "/etc/goTrack.yaml"

func main() {
	// Define command-line flags
	intervalFlag := pflag.IntP("interval", "i", 0, "Interval for polling USB devices in milliseconds")
	debugFlag := pflag.BoolP("debug", "d", false, "Debug mode: Print debugging notes")
	verboseFlag := pflag.BoolP("verbose", "v", false, "Print state at start")
	noExecFlag := pflag.BoolP("noExec", "n", false, "Do not execute on device detection")
	helpFlag := pflag.BoolP("help", "h", false, "Show help text")
	commandFlag := pflag.StringP("command", "c", "", "Command (chain) to be executed on device change detection")
	commandArgFlag := pflag.StringP("arguments", "a", "", "Command arguments")
	configPathFlag := pflag.StringP("configPath", "p", "", "Path to config")

	// Parse command-line flags
	pflag.Parse()

	// Show help text if help flag is provided
	if *helpFlag {
		showHelp()
		return
	}

	var configFilePath string
	if len(*configPathFlag) != 0 {
		tempPath := *configPathFlag
		if fileExists(tempPath) {
			configFilePath = tempPath
		}
	}

	if len(configFilePath) == 0 && fileExists(defaultConfigPath) {
		configFilePath = defaultConfigPath
	}

	var config *Config
	var err error
	debug := *debugFlag
	if len(configFilePath) != 0 {
		// Load configuration from file
		config, err = NewConfigFromFile(configFilePath)
		if err != nil {
			NewConfig().logErr(err)
			return
		}

		//Move old log
		if fileExists(config.LogFile) {
			if config.OldLogs < 1 {
				err := os.Remove(config.LogFile)
				if err != nil {
					NewConfig().logErr(err)
				}
			} else {
				for i := config.OldLogs - 1; i >= 1; i-- {
					oldName := fmt.Sprintf("%s.%d", config.LogFile, i)
					if fileExists(oldName) {
						newName := fmt.Sprintf("%s.%d", config.LogFile, i+1)
						err := os.Rename(oldName, newName)
						if err != nil {
							NewConfig().logErr(err)
						}
					}
				}
				if fileExists(config.LogFile) {
					newName := fmt.Sprintf("%s.%d", config.LogFile, 1)
					err := os.Rename(config.LogFile, newName)
					if err != nil {
						NewConfig().logErr(err)
					}
				}
			}
		}
	} else {
		config = NewConfig()
		if debug {
			config.log("Using default config")
		}
	}

	// Override interval with command-line flag if provided
	if *intervalFlag != 0 {
		config.Interval = *intervalFlag
	}

	// Overwrite command with command-line flag if provided
	if len(*commandFlag) != 0 {
		config.Commands = nil
		config.Commands = append(config.Commands, Command{
			Command: *commandFlag,
			Args:    nil,
		})

		// Overwrite command with command-line flag if provided
		if len(*commandArgFlag) != 0 {
			config.Commands[0].Args = append(config.Commands[0].Args, *commandArgFlag)
		}
	}

	// Do not execute if true
	noExec := *noExecFlag

	// Print Config for debugging
	if debug {
		yamlData, err := yaml.Marshal(&config)
		if err != nil {
			config.logErr(err)
		}
		config.log(string(yamlData))
	}

	// Delay execution before start
	time.Sleep(time.Duration(config.StartDelay) * time.Second)

	// Create file lock if activated
	if config.FileLock {
		createEmptyFileIfMissing(config.FileLockPath)
	}

	// Start ticker
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Millisecond)
	defer ticker.Stop()

	// Create USB tracker with loaded configuration
	var usbTracker *USBTracker
	if config.USBTracking {
		usbTracker = NewUSBTracker(config)
	}

	// Print Current state for verbose mode and Init usbTracker
	verbose := *verboseFlag
	if config.USBTracking {
		usbTracker.InitUSBDevices(verbose)
	}

	for {
		select {
		case <-ticker.C:
			if debug {
				// Format the time using the desired layout
				timeString := time.Now().Format("15:04:05.00") // hh:mm:ss,ss
				config.log("Tick at: " + timeString)
			}
			if config.USBTracking {
				usbTracker.TrackUSBDevices(noExec, debug)
			}
		}
	}
}

func showHelp() {
	fmt.Println("Usage: goTrack [OPTIONS]")
	pflag.PrintDefaults()
}
