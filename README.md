# goTrack
goTrack aims to replace usb killer projects like [usbkill by hephaest0s](https://github.com/hephaest0s/usbkill) in Go and with the idea to be extended to more than usb devices to track.

## Disclaimer
This project is meant to increase security as it allows automatic reactions to detected events like forensic interference. Nevertheless, we give no guaranty that this works flawless. Neither any guaranty that this won't harm your data. Use it at your own risk.

## Build
Just run
```shell
go tidy
go build
```

## Run
Run it with `-h` for help or `-n` to test without executing any (destructive) commands
```shell
goTrack -h
```
goTrack is meant to be run as root.

## Requirements
`lsusb` if usb tracking shall be enabled.

## Installation
Place the executable at `/usr/local/bin/goTrack` and the config file at `/etc/goTrack.yaml`.

## Config
The config file should be easy to understand.

## Version
1.8.1

### Change log
#### V1.8.1
Config file correction for timestamps
#### V1.8
Added actions by timer
#### V1.7.1
Added more debugging output
Added version checking
#### V1.7
Added fileLock inverted mode
#### V1.6
Added execution on error configuration
#### V1.5
Added possibility to bind commands to targets
#### V1.4
Added basic testing for development.
#### V1.3
Independent intervals for all trackers.
#### V1.2
Refactoring of configuration.
#### V1.1
Changed `fileLock` functionality to be blocking not enabling. See default `goTrack.yaml` for new functionality.
#### V1.0
Initial release

## Future Work
In the future this project can be extended with the ability to check:
- Events: React to events like shortcuts