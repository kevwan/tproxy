package main

import (
	"net"
	"time"
)

type Stater interface {
	AddConn(key string, conn *net.TCPConn)
	DelConn(key string)
	Start()
}

type NilPrinter struct{}

func NewNilPrinter(_ time.Duration) Stater {
	return NilPrinter{}
}

func (p NilPrinter) AddConn(_ string, _ *net.TCPConn) {
}

func (p NilPrinter) DelConn(_ string) {
}

func (p NilPrinter) Start() {
}
