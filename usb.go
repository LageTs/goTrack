package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// USBDevice represents a connected USB device
type USBDevice struct {
	ID   string
	Name string
}

// USBTracker represents the USB tracking service
type USBTracker struct {
	Config        *Config
	cachedDevices map[string]USBDevice
}

// NewUSBTracker creates a new USBTracker instance
func NewUSBTracker(config *Config) *USBTracker {
	return &USBTracker{
		Config: config,
	}
}

// InitUSBDevices initializes the USB devices list
func (u *USBTracker) InitUSBDevices(verbose bool) {
	u.cachedDevices = u.getConnectedUSBDevices()
	if verbose {
		fmt.Println("Connected at start:")
		for id, device := range u.cachedDevices {
			fmt.Println(id, " ", device.Name)
		}
	}
}

// TrackUSBDevices periodically tracks connected USB devices
func (u *USBTracker) TrackUSBDevices(noExec, debug bool) {
	// Get list of currently connected USB devices
	currentDevices := u.getConnectedUSBDevices()

	// Check for new devices
	for id, device := range currentDevices {
		if !u.deviceIDExists(id) {
			if !u.deviceIDIgnored(id) {
				u.Config.log("New ID: " + id + " Name: " + device.Name)
				if !noExec {
					u.Config.exec(CalleeUSB)
				} else {
					u.Config.log("Execution aborted due to \"NoExec\"")
				}
			} else if debug {
				u.Config.log("New device from ignored IDs: " + id + " Name: " + device.Name)
			}
			u.cachedDevices[id] = device
		}
	}

	// Check for missing devices
	for id, device := range u.cachedDevices {
		if deviceIDMissing(currentDevices, id) {
			if !u.deviceIDIgnored(id) {
				u.Config.log("Old missing ID: " + id + " Name: " + device.Name)
				if !noExec {
					u.Config.exec(CalleeUSB)
				} else {
					u.Config.log("Execution aborted due to \"NoExec\"")
				}
			} else {
				if debug {
					u.Config.log("Ignored missing device ID: " + id + " Name: " + device.Name)
				}
			}
			delete(u.cachedDevices, id)
		}
	}
}

// getConnectedUSBDevices retrieves currently connected USB devices
func (u *USBTracker) getConnectedUSBDevices() map[string]USBDevice {
	output, err := exec.Command("lsusb").Output()
	if err != nil {
		u.Config.logErr(err)
		return nil
	}

	devices := make(map[string]USBDevice)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			id := parts[5]
			name := combineFields(parts)
			devices[id] = USBDevice{ID: id, Name: name}
		}
	}

	return devices
}

// deviceIDIgnored checks if a device ID is ignored
func (u *USBTracker) deviceIDIgnored(id string) bool {
	return has(u.Config.IgnoredIDs, id)
}

// deviceIDExists checks if a device ID already exists in the cache
func (u *USBTracker) deviceIDExists(id string) bool {
	_, exists := u.cachedDevices[id]
	return exists
}

func deviceIDMissing(current map[string]USBDevice, id string) bool {
	_, exists := current[id]
	return !exists
}

func combineFields(array []string) string {
	// Initialize an empty string to store the combined fields
	var combined string

	// Iterate over the array starting from the second element
	for i := 1; i < len(array); i++ {
		// Append each field to the combined string with a whitespace delimiter
		combined += array[i] + " "
	}

	// Trim any trailing whitespace
	combined = strings.TrimSpace(combined)

	return combined
}

func has(array []string, id string) bool {
	for _, s := range array {
		if s == id {
			return true
		}
	}
	return false
}
