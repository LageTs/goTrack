package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const defaultConfigPath = "/etc/goTrack.yaml"
const version = "1.7"

func main() {
	// Define command-line flags
	intervalFlag := pflag.DurationP("interval", "i", time.Duration(0), "Interval for tracking")
	debugFlag := pflag.BoolP("debug", "d", false, "Debug mode: Print debugging notes")
	verboseFlag := pflag.BoolP("verbose", "x", false, "Print state at start")
	noExecFlag := pflag.BoolP("noExec", "n", false, "Do not execute on device detection")
	helpFlag := pflag.BoolP("help", "h", false, "Show help text")
	commandFlag := pflag.StringP("command", "c", "", "Command (chain) to be executed")
	commandArgFlag := pflag.StringP("arguments", "a", "", "Command arguments")
	configPathFlag := pflag.StringP("configPath", "p", "", "Path to config")
	versionFlag := pflag.BoolP("version", "v", false, "Print version text")

	// Parse command-line flags
	pflag.Parse()

	// Show help text if help flag is provided
	if *helpFlag {
		showHelp()
		return
	}

	// Show version text
	if *versionFlag {
		fmt.Println("Version: goTrack " + version)
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
		config.USBInterval = *intervalFlag
		config.PingInterval = *intervalFlag
		config.WebInterval = *intervalFlag
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

	// Delete file lock if FileLockDeletion is activated
	if config.FileLockDeletion {
		config.deleteFileIfExisting(config.FileLockPath)
	} else if config.FileLockCreation {
		// Create file lock if FileLockCreation is activated
		// else if as Deletion and Creation can't be activated both
		config.createEmptyFileIfMissing(config.FileLockPath)
	}

	// Convert verboseFlag to var
	verbose := *verboseFlag

	config.printAndLog("Waiting for " + config.StartDelay.String() + " at: " + time.Now().Format("15:04:05.00")) // hh:mm:ss,ss

	// Delay execution before start
	time.Sleep(config.StartDelay)
	config.printAndLog("Finished waiting at: " + time.Now().Format("15:04:05.00"))

	// Create USB tracker with loaded configuration
	if config.USBTracking {
		usbTracker := NewUSBTracker(config)
		usbTracker.InitUSBDevices(verbose)

		// Start ticker
		usbTicker := time.NewTicker(config.USBInterval)
		defer usbTicker.Stop()

		config.printAndLog("Started USB tracking at: " + time.Now().Format("15:04:05.00"))

		// Start tracking
		go func() {
			for {
				select {
				case <-usbTicker.C:
					go usbTracker.TrackUSBDevices(noExec, debug)
				}
			}
		}()
	}

	if config.PingTracking {
		pingTracker := NewPingTracker(config)

		// Start ticker
		pingTicker := time.NewTicker(config.PingInterval)
		defer pingTicker.Stop()

		config.printAndLog("Started Ping tracking at: " + time.Now().Format("15:04:05.00"))

		go func() {
			for {
				select {
				case <-pingTicker.C:
					go pingTracker.TrackPingTargets(noExec, debug)
				}
			}
		}()
	}

	if config.WebTracking {
		webTracker := NewWebTracker(config)

		// Start ticker
		webTicker := time.NewTicker(config.WebInterval)
		defer webTicker.Stop()

		config.printAndLog("Started Web tracking at: " + time.Now().Format("15:04:05.00"))

		go func() {
			for {
				select {
				case <-webTicker.C:
					go webTracker.TrackWebSources(noExec, debug)
				}
			}
		}()
	}

	select {} // keeps program running
}

func showHelp() {
	fmt.Println("Usage: goTrack [OPTIONS]")
	pflag.PrintDefaults()
}
