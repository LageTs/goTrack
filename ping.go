package main

import "time"
import "github.com/prometheus-community/pro-bing"

// PingTracker represents the Ping tracking service
type PingTracker struct {
	Config *Config
}

// NewPingTracker returns a new PingTracker with given Config
func NewPingTracker(config *Config) *PingTracker {
	return &PingTracker{Config: config}
}

// TrackPingTargets tracks ping targets. Meant to be executed periodically. Starts async pings for each config.
func (p *PingTracker) TrackPingTargets(noExec, debug bool) {
	for _, config := range p.Config.PingTrackingConfigs {
		go p.ping(noExec, debug, &config)
	}
}

// ping executes the ping and decides for executions calls. Meant to be executed async.
func (p *PingTracker) ping(noExec, debug bool, pingConfig *PingConfig) {
	pinger, err := probing.NewPinger(pingConfig.Target)
	if err != nil {
		p.Config.log(err.Error())
		p.Config.exec(CalleePing, noExec)
		return
	}
	pinger.Count = 1
	pinger.Timeout = pingConfig.PingTimeout

	for i := 0; i < pingConfig.RetryCount; i++ {
		if debug && i > 0 {
			p.Config.log("Retrying: " + pingConfig.Target)
		}

		err = pinger.Run()
		if err != nil {
			p.Config.log(err.Error())
			p.Config.exec(CalleePing, noExec)
		}
		success := pinger.Statistics().PacketsRecv > 0

		// Check for Config
		if pingConfig.OnSuccess {
			// If ping returns -> execute. Else return
			if success {
				p.Config.log("Ping successful. Execution started due to OnSuccess for " + pingConfig.Target)
				p.Config.exec(CalleePing, noExec)
			}
			return
		} else {
			// If ping fails -> retry. Else return
			if success {
				return
			}
			if debug {
				p.Config.log("Pinging failed for: " + pingConfig.Target)
			}
		}
		// Wait
		time.Sleep(pingConfig.RetryDelay)
	}
	p.Config.exec(CalleePing, noExec)
}
