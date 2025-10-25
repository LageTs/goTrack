# goTrack
goTrack aims to replace usb killer projects like [usbkill by hephaest0s](https://github.com/hephaest0s/usbkill) in Go and with the idea to be extended to more than usb devices to track.

## Disclaimer
This project is meant to increase security as it allows automatic reactions to detected events like forensic interference. Nevertheless, we give no guaranty that this works flawless. Neither any guaranty that this won't harm your data. Use it at your own risk.

## Build
Just run
```shell
go build
```

## Run
Run it with `-h` for help or `-n` to test without executing any destructive commands
```shell
goTrack -h
```
goTrack is meant to be run as root.

## Requirements
`lsusb` if usb tracking shall be enabled.

## Installation
Place the executable at `/usr/local/bin/goTrack` and the config file at `/etc/goTrack.yaml`.

## Config
The config file should be easy to understand, here is how you set the commands to be executed in worst case:
```yaml
commands:
  - command: "shutdown" # Command
    args: # Arguments as strings
      - "0"
    late: true # Set true to execute commands after others
    usb: true # Set true to execute command on usb changes
    ping: false # Set true to execute command on ping tracking
    web: false # Set true to execute command on web tracking
```
`command` is the terminal command to be executed if `goTrack` detects a not ignored change.
`args` are the arguments for that command
`late` Commands are executed in two queues: `late=false` commands will guaranteed be executed before `late=true` commands. Commands with the same `late` state are meant to be executed in order but that is not guaranteed.
`usb` commands with this parameter set to true will be executed with an usb-change detected by goTrack. `ping` and `web` work in the same way as `usb`.
```
usb_ignored_ids:
- "Test"
```
Includes ignored IDs for usb tracking. Those IDs can be seen if `goTrack -n` is run. Example:
```
usb_ignored_ids:
- "1234:5678"
- "ABCD:9876"
```

## Version
1.6

### Change log
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
- (optional) Timer: React if some timings are reached (period, time stamp)