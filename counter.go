package main

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

type connCounter struct {
	total       int64
	concurrent  int64
	max         int64
	conns       map[string]time.Time
	maxLifetime time.Duration
	lock        sync.Mutex
}

func NewConnCounter() Stater {
	return &connCounter{
		conns: make(map[string]time.Time),
	}
}

func (c *connCounter) AddConn(key string, conn *net.TCPConn) {
	atomic.AddInt64(&c.total, 1)
	val := atomic.AddInt64(&c.concurrent, 1)
	max := atomic.LoadInt64(&c.max)
	if val > max {
		atomic.CompareAndSwapInt64(&c.max, max, val)
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.conns[key] = time.Now()
}

func (c *connCounter) DelConn(key string) {
	atomic.AddInt64(&c.concurrent, -1)

	c.lock.Lock()
	defer c.lock.Unlock()
	start, ok := c.conns[key]
	delete(c.conns, key)
	if ok {
		lifetime := time.Since(start)
		if lifetime > c.maxLifetime {
			c.maxLifetime = lifetime
		}
	}
}

func (c *connCounter) Start() {
}

func (c *connCounter) Stop() {
	fmt.Println()
	color.HiWhite("Total connections: %d", atomic.LoadInt64(&c.total))
	color.HiWhite("Max concurrent connections: %d", atomic.LoadInt64(&c.max))
	color.HiWhite("Max connection lifetime: %s", c.maxLifetime)
}
