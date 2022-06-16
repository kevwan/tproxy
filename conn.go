package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/kevwan/tproxy/display"
	"github.com/kevwan/tproxy/protocol"
)

const (
	serverSide      = "SERVER"
	clientSide      = "CLIENT"
	useOfClosedConn = "use of closed network connection"
)

var errClientCanceled = errors.New("client canceled")

type PairedConnection struct {
	id       int
	cliConn  net.Conn
	svrConn  net.Conn
	once     sync.Once
	stopChan chan struct{}
}

func NewPairedConnection(id int, cliConn net.Conn) *PairedConnection {
	return &PairedConnection{
		id:       id,
		cliConn:  cliConn,
		stopChan: make(chan struct{}),
	}
}

func (c *PairedConnection) handleClientMessage() {
	// client closed also trigger server close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(c.svrConn, w)

	go protocol.NewDumper(r, clientSide, c.id, settings.Quiet, protocol.CreateInterop(settings.Protocol)).Dump()

	_, e := io.Copy(tee, c.cliConn)
	if e != nil && e != io.EOF {
		color.HiRed("handleClientMessage: io.Copy error: %v", e)
	}
}

func (c *PairedConnection) handleServerMessage() {
	// server closed also trigger client close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(newDelayedWriter(c.cliConn, settings.Delay, c.stopChan), w)
	go protocol.NewDumper(r, serverSide, c.id, settings.Quiet, protocol.CreateInterop(settings.Protocol)).Dump()
	_, e := io.Copy(tee, c.svrConn)
	if e != nil && e != io.EOF {
		netOpError, ok := e.(*net.OpError)
		if ok && netOpError.Err.Error() != useOfClosedConn {
			color.HiRed("handleServerMessage: io.Copy error: %v", e)
		}
	}
}

func (c *PairedConnection) process() {
	defer c.stop()

	conn, err := net.Dial("tcp", settings.RemoteHost)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("[x][%d] Couldn't connect to server: %v", c.id, err))
		return
	}

	display.PrintlnWithTime(color.HiGreenString("[%d] Connected to server: %s", c.id, conn.RemoteAddr()))

	c.svrConn = conn
	go c.handleServerMessage()

	c.handleClientMessage()
}

func (c *PairedConnection) stop() {
	c.once.Do(func() {
		close(c.stopChan)
		if c.cliConn != nil {
			display.PrintlnWithTime(color.HiBlueString("[%d] Client connection closed", c.id))
			c.cliConn.Close()
		}
		if c.svrConn != nil {
			display.PrintlnWithTime(color.HiBlueString("[%d] Server connection closed", c.id))
			c.svrConn.Close()
		}
	})
}

func startListener() error {
	conn, err := net.Listen("tcp", fmt.Sprint(settings.LocalHost, ":", settings.LocalPort))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	display.PrintlnWithTime("Listening...")
	defer conn.Close()

	var connIndex int
	for {
		cliConn, err := conn.Accept()
		if err != nil {
			return fmt.Errorf("server: accept: %w", err)
		}

		connIndex++
		display.PrintlnWithTime(color.HiGreenString("[%d] Accepted from: %s", connIndex, cliConn.RemoteAddr()))

		pconn := NewPairedConnection(connIndex, cliConn)
		go pconn.process()
	}
}

type delayedWriter struct {
	writer   io.Writer
	delay    time.Duration
	stopChan <-chan struct{}
}

func newDelayedWriter(writer io.Writer, delay time.Duration, stopChan <-chan struct{}) delayedWriter {
	return delayedWriter{
		writer:   writer,
		delay:    delay,
		stopChan: stopChan,
	}
}

func (w delayedWriter) Write(p []byte) (int, error) {
	if w.delay == 0 {
		return w.writer.Write(p)
	}

	timer := time.NewTimer(w.delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return w.writer.Write(p)
	case <-w.stopChan:
		return 0, errClientCanceled
	}
}
