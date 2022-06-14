package main

type Settings struct {
	RemoteHost string
	LocalHost  string
	LocalPort  int
	Protocol   string
	Silent     bool
}

func saveSettings(localHost string, localPort int, remoteHost, protocol string, silent bool) {
	if localHost != "" {
		settings.LocalHost = localHost
	}
	if localPort != 0 {
		settings.LocalPort = localPort
	}
	if remoteHost != "" {
		settings.RemoteHost = remoteHost
	}
	settings.Protocol = protocol
	settings.Silent = silent
}
