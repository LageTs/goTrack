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

### Config examples
In the following I will describe some commands with examples for their usage in combination with detection modes.
#### Shutdown 0
This command will shut down your system gratefully. This could be useful if you're on full disk encryption and you want to be sure the system is not left unlocked to an attacker. Late command execution is adviced, exspecially if there are other commands configured that shall be completed before shut down.
```
  - command: "shutdown" # Command
    args: # Arguments as strings
      - "0"
    late: true
    usb: true
    ping: false
    web: false
    time: false
    interval: false
    command_id: -1
```

#### loginctl lock-sessions
This command (depending on system, especially wether systemd is used) will lock all running sessions. This command is useful if you want to exclude an attacker from direct physical access to an unlocked system. As standalone this offers very low protection. Could be used on intervals to be sure your system won't stay unlocked for infinite time.
```
  - command: "loginctl"
    args:
      - "lock-sessions"
    late: false
    usb: false
    ping: false
    web: false
    time: false
    interval: true
    command_id: -1
```

#### Ping example local net
This configurations tries to ping a target on the local network every 10 seconds and triggers on lost connection if ping fails multiple times. This could be used for either reconnection attempts or security reactions like shut downs or network interface shut downs as a corrupted local network is expected.
```
ping_tracking: true
ping_interval: 10000ms
ping_targets:
  - target: "192.168.1.1"
    ping_timeout: 1s
    on_success: false
    retry_count: 3
    retry_delay: 100ms
    command_id: -1
```

#### Ping example internet
This configurations tries to ping a target on the internet every 10 seconds and triggers on successful connection. This could be used if you want to react to unwanted internet access or to react to the status of a web service.
```
ping_tracking: true
ping_interval: 10000ms
ping_targets:
  - target: "8.8.8.8"
    ping_timeout: 1s
    on_success: true
    retry_count: 3
    retry_delay: 100ms
    command_id: -1
```

#### Web Tracking example: Content
This configurations could track your personal status page on the web that is used as a kill switch. Could be used in combination with deletion of files, disks or encryptions headers.
```
web_tracking: true
web_interval: 60000ms
web_targets:
  - target: "https://yourpage.link"
    content: "SECURITY BREACH"
    content_is_exact: true
    status_code: 200
    on_code_identical: false
    on_https_fails: false
    retry_count: 3
    retry_delay: 500ms
    command_id: -1
```
    
#### Web Tracking example: Status
This configuration triggers if the Webpage is reached and returns HTTP status 200 (OK)
```
web_tracking: true
web_interval: 60000ms
web_targets:
  - target: "https://example.com"
    content: "Test"
    content_is_exact: false
    status_code: 200
    on_code_identical: true
    on_https_fails: false
    retry_count: 3
    retry_delay: 500ms
    command_id: -1
```

#### Interval example
This configuration will trigger every 30 minutes unless the system time is not in the year 2026 (UTC)
```
interval_tracking: true
# Timestamps to react to
interval_targets:
  - interval: 30m
    start_at: "2026-01-01T00:00:00Z" # Timestamp as ISO 8601. Do not react before timestamp. Must be set correctly
    stop_at: "2027-01-01T00:00:00Z" # Timestamp as ISO 8601. Do not react after timestamp. Must be set correctly
    execute_on_start: false # If true, commands will be executed once at start and then every set duration.
    command_id: -1
```
    
#### Timestamp example exceeded
This configuration will trigger if the configured timestamp has exceeded somewhen during or before program execution.
```
time_tracking: true
time_targets:
  - timestamp: "2000-01-20T12:31:00Z" # Timestamp as ISO 8601, UTC or with timezone (Z01:00 for UTC+1)
    tolerance_window: -1
    command_id: -1
```

    
#### Timestamp example exact
This configuration will trigger if the configured timestamp has exceeded somewhen during program execution.
```
time_tracking: true
time_targets:
  - timestamp: "2000-01-20T12:31:00Z" # Timestamp as ISO 8601, UTC or with timezone (Z01:00 for UTC+1)
    tolerance_window: 0
    command_id: -1
```

## Version
1.8.2

### Change log
#### V1.8.2
Bug fix for execution of time and interval commands
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
- Web Tracking: Configuration option if content is or is not like defined
