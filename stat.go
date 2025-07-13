package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type (
	Stater interface {
		AddConn(key string, conn *net.TCPConn)
		DelConn(key string)
		Start()
		Stop()
	}

	compositeStater struct {
		staters []Stater
	}
)

func NewStater(staters ...Stater) Stater {
	stat := compositeStater{
		staters: append([]Stater(nil), staters...),
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		for sig := range c {
			signal.Stop(c)
			stat.Stop()

			p, err := os.FindProcess(syscall.Getpid())
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			if err := p.Signal(sig); err != nil {
				fmt.Println(err)
			}
		}
	}()

	return stat
}

func (c compositeStater) AddConn(key string, conn *net.TCPConn) {
	for _, s := range c.staters {
		s.AddConn(key, conn)
	}
}

func (c compositeStater) DelConn(key string) {
	for _, s := range c.staters {
		s.DelConn(key)
	}
}

func (c compositeStater) Start() {
	for _, s := range c.staters {
		s.Start()
	}
}

func (c compositeStater) Stop() {
	for _, s := range c.staters {
		s.Stop()
	}
}

type NilPrinter struct{}

func (p NilPrinter) AddConn(_ string, _ *net.TCPConn) {
}

func (p NilPrinter) DelConn(_ string) {
}

func (p NilPrinter) Start() {
}

func (p NilPrinter) Stop() {
}
