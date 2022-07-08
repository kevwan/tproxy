package main

import "time"

type Settings struct {
	Remote    string
	LocalHost string
	LocalPort int
	Delay     time.Duration
	Protocol  string
	Stat      bool
	Quiet     bool
}

func saveSettings(localHost string, localPort int, remote string, delay time.Duration,
	protocol string, stat, quiet bool) {
	if localHost != "" {
		settings.LocalHost = localHost
	}
	if localPort != 0 {
		settings.LocalPort = localPort
	}
	if remote != "" {
		settings.Remote = remote
	}
	settings.Delay = delay
	settings.Protocol = protocol
	settings.Stat = stat
	settings.Quiet = quiet
}
