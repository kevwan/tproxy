package main

import "time"

type Settings struct {
	RemoteHost string
	LocalHost  string
	LocalPort  int
	Delay      time.Duration
	Protocol   string
	Silent     bool
}

func saveSettings(localHost string, localPort int, remoteHost string, delay time.Duration,
	protocol string, silent bool) {
	if localHost != "" {
		settings.LocalHost = localHost
	}
	if localPort != 0 {
		settings.LocalPort = localPort
	}
	if remoteHost != "" {
		settings.RemoteHost = remoteHost
	}
	settings.Delay = delay
	settings.Protocol = protocol
	settings.Silent = silent
}
