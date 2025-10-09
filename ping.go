package main

import "time"
import "github.com/prometheus-community/pro-bing"

// PingTracker represents the Ping tracking service
type PingTracker struct {
	Config *Config
}

const PingSuc uint8 = 0
const PingNoSuc uint8 = 1
const PingExec uint8 = 2
const PingErr uint8 = 3
const PingTimeoutErr uint8 = 4

// NewPingTracker returns a new PingTracker with given Config
func NewPingTracker(config *Config) *PingTracker {
	return &PingTracker{Config: config}
}

// TrackPingTargets tracks ping targets. Meant to be executed periodically. Starts async pings for each config.
func (p *PingTracker) TrackPingTargets(noExec, debug bool) uint {
	var counter uint = 0
	for _, config := range p.Config.PingTrackingConfigs {
		go p.ping(noExec, debug, &config)
		counter++
	}
	return counter
}

// ping executes the ping and decides for executions calls. Meant to be executed async.
func (p *PingTracker) ping(noExec, debug bool, pingTarget *PingTarget) uint8 {
	pinger, err := probing.NewPinger(pingTarget.Target)
	if err != nil {
		p.Config.log(err.Error())
		if p.Config.ExecOnError {
			p.Config.exec(CalleePing, pingTarget.CommandId, noExec)
			return PingExec
		}
		return PingErr
	}
	if pingTarget.PingTimeout == 0 {
		p.Config.log("Timeout must be greater than zero")
		if p.Config.ExecOnError {
			p.Config.exec(CalleePing, pingTarget.CommandId, noExec)
			return PingExec
		}
		return PingTimeoutErr
	}
	pinger.Count = 1
	pinger.Timeout = pingTarget.PingTimeout

	for i := 0; i <= pingTarget.RetryCount; i++ {
		if debug && i > 0 {
			p.Config.log("Retrying: " + pingTarget.Target)
		}

		err = pinger.Run()
		if err != nil {
			p.Config.log(err.Error())
			if p.Config.ExecOnError {
				p.Config.exec(CalleePing, pingTarget.CommandId, noExec)
				return PingExec
			}
			return PingErr
		}
		success := pinger.Statistics().PacketsRecv > 0

		// Check for Config
		if pingTarget.OnSuccess {
			// If ping returns -> execute. Else return
			if success {
				p.Config.log("Ping successful. Execution started due to OnSuccess for " + pingTarget.Target)
				p.Config.exec(CalleePing, pingTarget.CommandId, noExec)
				return PingExec
			}
			return PingNoSuc
		} else {
			// If ping fails -> retry. Else return
			if success {
				return PingSuc
			}
			if debug {
				p.Config.log("Pinging failed for: " + pingTarget.Target)
			}
		}
		// Wait
		time.Sleep(pingTarget.RetryDelay)
	}
	p.Config.log("Ping unsuccessful after maximum retries. Execution started due failure for " + pingTarget.Target)
	p.Config.exec(CalleePing, pingTarget.CommandId, noExec)
	return PingExec
}
