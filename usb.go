package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// USBDevice represents a connected USB device
type USBDevice struct {
	ID       string
	Name     string
	BusCount map[string]uint8
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
		fmt.Println("Connected at start:\nID        Count Name")
		for id, device := range u.cachedDevices {
			sum := uint8(0)
			for _, count := range device.BusCount {
				sum += count
			}
			s := string(sum)
			for i := uint8(0); len(s) < 4; i++ {
				s = s + " "
			}
			fmt.Println(id, sum, s, device.Name)
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
			// New device ID found
			if !u.deviceIDIgnored(id) {
				// ID not ignored -> execute commands
				u.Config.log("New ID: " + id + " Name: " + device.Name)
				u.Config.exec(CalleeUSB, noExec)
			} else if debug {
				u.Config.log("New device from ignored IDs: " + id + " Name: " + device.Name)
			}

			// Add device to cache as it is now registered
			u.cachedDevices[id] = device
		} else {
			// Device ID is known
			if !device.isBusCountEqual(u.cachedDevices[id]) {
				// Number of devices with same ID is not same as before
				if !u.deviceIDIgnored(id) {
					// ID not ignored -> execute commands
					cacheDevice := u.cachedDevices[id]
					u.Config.log("Device count differs for ID: " + id + " Name: " + device.Name)
					u.Config.log("Old count: " + strconv.Itoa(int(cacheDevice.getBusSum())) + " | New Count: " + strconv.Itoa(int(device.getBusSum())))
					u.Config.exec(CalleeUSB, noExec)
				} else if debug {
					u.Config.log("Device count differs for ignored ID: " + id + " Name: " + device.Name)
				}
				u.cachedDevices[id] = device
			}
		}
	}

	// Check for missing devices
	for id, device := range u.cachedDevices {
		if deviceIDMissing(currentDevices, id) {
			if !u.deviceIDIgnored(id) {
				u.Config.log("Old missing ID: " + id + " Name: " + device.Name)
				u.Config.exec(CalleeUSB, noExec)
			} else {
				if debug {
					u.Config.log("Ignored missing device ID: " + id + " Name: " + device.Name)
				}
			}
			// Delete as state is updated
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
			bus := parts[1]
			name := combineFields(parts, 6)
			if u, ok := devices[id]; ok {
				if _, ok := u.BusCount[bus]; ok {
					devices[id].BusCount[bus] += 1
				} else {
					devices[id].BusCount[bus] = 1
				}
			} else {
				m := make(map[string]uint8)
				m[bus] = 1
				devices[id] = USBDevice{ID: id, Name: name, BusCount: m}
			}
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

func (u *USBDevice) isBusCountEqual(u2 USBDevice) bool {
	if len(u.BusCount) != len(u2.BusCount) {
		return false
	}
	for bus, count := range u.BusCount {
		if u2.BusCount[bus] != count {
			return false
		}
	}
	return true
}

func (u *USBDevice) getBusSum() uint8 {
	busSum := uint8(0)
	for _, c := range u.BusCount {
		busSum += c
	}
	return busSum
}

func deviceIDMissing(current map[string]USBDevice, id string) bool {
	_, exists := current[id]
	return !exists
}

func combineFields(array []string, startIndex int) string {
	// Initialize an empty string to store the combined fields
	var combined string

	// Iterate over the array starting from the second element
	for i := startIndex; i < len(array); i++ {
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
