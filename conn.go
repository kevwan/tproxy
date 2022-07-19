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
	statInterval    = time.Second * 5
)

var (
	errClientCanceled = errors.New("client canceled")
	stat              Stater
)

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

func (c *PairedConnection) copyData(dst io.Writer, src io.Reader, tag string) {
	_, e := io.Copy(dst, src)
	if e != nil && e != io.EOF {
		netOpError, ok := e.(*net.OpError)
		if ok && netOpError.Err.Error() != useOfClosedConn {
			reason := netOpError.Unwrap().Error()
			display.PrintlnWithTime(color.HiRedString("[%d] %s error, %s", c.id, tag, reason))
		}
	}
}

func (c *PairedConnection) handleClientMessage() {
	// client closed also trigger server close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(c.svrConn, w)
	go protocol.NewDumper(r, clientSide, c.id, settings.Quiet, protocol.CreateInterop(settings.Protocol)).Dump()
	c.copyData(tee, c.cliConn, clientSide)
}

func (c *PairedConnection) handleServerMessage() {
	// server closed also trigger client close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(newDelayedWriter(c.cliConn, settings.Delay, c.stopChan), w)
	go protocol.NewDumper(r, serverSide, c.id, settings.Quiet, protocol.CreateInterop(settings.Protocol)).Dump()
	c.copyData(tee, c.svrConn, serverSide)
}

func (c *PairedConnection) process() {
	defer c.stop()

	conn, err := net.Dial("tcp", settings.Remote)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("[x][%d] Couldn't connect to server: %v", c.id, err))
		return
	}

	display.PrintlnWithTime(color.HiGreenString("[%d] Connected to server: %s", c.id, conn.RemoteAddr()))

	stat.AddConn(fmt.Sprintf("%d:server", c.id), conn.(*net.TCPConn))
	c.svrConn = conn
	go c.handleServerMessage()

	c.handleClientMessage()
}

func (c *PairedConnection) stop() {
	c.once.Do(func() {
		close(c.stopChan)
		stat.DelConn(fmt.Sprintf("%d:server", c.id))
		stat.DelConn(fmt.Sprintf("%d:client", c.id))

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
	stat = NewStatPrinter(statInterval)
	go stat.Start()

	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", settings.LocalHost, settings.LocalPort))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer conn.Close()

	display.PrintfWithTime("Listening on %s...\n", conn.Addr().String())

	var connIndex int
	for {
		cliConn, err := conn.Accept()
		if err != nil {
			return fmt.Errorf("server: accept: %w", err)
		}

		connIndex++
		display.PrintlnWithTime(color.HiGreenString("[%d] Accepted from: %s",
			connIndex, cliConn.RemoteAddr()))

		stat.AddConn(fmt.Sprintf("%d:client", connIndex), cliConn.(*net.TCPConn))
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
