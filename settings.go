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
	UpLimit   int64
	DownLimit int64
}

func saveSettings(localHost string, localPort int, remote string, delay time.Duration,
	protocol string, stat, quiet bool, upLimit, downLimit int64) {
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
	settings.UpLimit = upLimit
	settings.DownLimit = downLimit
}
