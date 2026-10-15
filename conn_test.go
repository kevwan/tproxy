package main

import (
	"io"
	"net"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// pickFreePort grabs an ephemeral port that is very likely still free.
func pickFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// TestConnectionGoroutinesAreReleased verifies that per-connection goroutines
// are released once a proxied connection closes.
//
// handleClientMessage and handleServerMessage each create an io.Pipe, hand its
// writer to the copy path and start a protocol-dump goroutine reading from the
// pipe. The reader only observes EOF when the writer is closed, so without
// closing it the dump goroutines stay blocked on r.Read forever: two leaked
// goroutines per proxied connection.
func TestConnectionGoroutinesAreReleased(t *testing.T) {
	backend, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("backend listen: %v", err)
	}
	defer backend.Close()
	go func() {
		for {
			c, err := backend.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c) // echo
			}(c)
		}
	}()

	settings = Settings{
		Remote:    backend.Addr().String(),
		LocalHost: "127.0.0.1",
		LocalPort: pickFreePort(t),
	}
	go func() { _ = startListener() }()
	time.Sleep(300 * time.Millisecond)

	baseline := runtime.NumGoroutine()

	const connections = 5
	for i := 0; i < connections; i++ {
		conn, err := net.Dial("tcp",
			net.JoinHostPort(settings.LocalHost, strconv.Itoa(settings.LocalPort)))
		if err != nil {
			t.Fatalf("dial %d: %v", i, err)
		}
		if _, err := conn.Write([]byte("hello")); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		if _, err := io.ReadFull(conn, make([]byte, 5)); err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		_ = conn.Close()
	}

	// Give the proxy time to observe the closes and retire per-connection work.
	var current int
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		current = runtime.NumGoroutine()
		if current <= baseline+2 {
			break
		}
	}

	t.Logf("goroutines baseline=%d after=%d connections=%d", baseline, current, connections)
	if current > baseline+2 {
		t.Fatalf("per-connection goroutines leaked: baseline=%d after=%d", baseline, current)
	}
}
