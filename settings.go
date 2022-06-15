package main

import "time"

type Settings struct {
	RemoteHost string
	LocalHost  string
	LocalPort  int
	Delay      time.Duration
	Protocol   string
	Quiet      bool
}

func saveSettings(localHost string, localPort int, remoteHost string, delay time.Duration,
	protocol string, quiet bool) {
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
	settings.Quiet = quiet
}
