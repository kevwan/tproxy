package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
)

var settings Settings

func main() {
	var (
		localPort = flag.Int("p", 0, "Local port to listen on")
		localHost = flag.String("l", "localhost", "Local address to listen on")
		remote    = flag.String("r", "", "Remote address (host:port) to connect")
		delay     = flag.Duration("d", 0, "the delay to relay packets")
		protocol  = flag.String("t", "", "The type of protocol, currently support grpc")
		silent    = flag.Bool("silent", false, "Only prints connection open/close and stats, default false")
	)

	flag.Parse()
	saveSettings(*localHost, *localPort, *remote, *delay, *protocol, *silent)

	if settings.RemoteHost == "" {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Remote host required"))
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := startListener(); err != nil {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Failed to start listener: %v", err))
		os.Exit(1)
	}
}
