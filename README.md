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
```
`command` is the terminal command to be executed if `goTrack` detects a not ignored change.
`args` are the arguments for that command
`late` Commands are executed in two queues: `late=false` commands will guaranteed be executed before `late=true` commands. Commands with the same `late` state are meant to be executed in order but that is not guaranteed.
`usb` commands with this parameter set to true will be executed with an usb-change detected by goTrack. This is meant for later extensions of goTrack to track other changes.

```
ignoredIDs:
- "Test"
```
Includes ignored IDs for usb tracking. Those IDs can be seen if `goTrack -n` is run. Example:
```
ignoredIDs:
- "1234:5678"
- "ABCD:9876"
```

## Version
1.0

## Future Work
In the future this project can be extended with the ability to check:
- Events: React to events like shortcuts
- (optional) Timer: React if some timings are reached (period, time stamp)