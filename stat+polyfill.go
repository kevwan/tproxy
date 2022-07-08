//go:build !linux

package main

import (
	"net"
	"time"
)

type StatPrinter struct{}

func NewStatPrinter(_ time.Duration) Stater {
	return StatPrinter{}
}

func (p StatPrinter) AddConn(_ string, _ *net.TCPConn) {
}

func (p StatPrinter) DelConn(_ string) {
}

func (p StatPrinter) Start() {
}
