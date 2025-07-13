package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/kevwan/tproxy/display"
	"github.com/olekukonko/tablewriter"
)

type StatPrinter struct {
	duration time.Duration
	conns    map[string]*net.TCPConn
	prev     map[string]*TcpInfo
	lock     sync.RWMutex
}

func NewStatPrinter(duration time.Duration) Stater {
	if !settings.Stat {
		return NilPrinter{}
	}

	return &StatPrinter{
		duration: duration,
		conns:    make(map[string]*net.TCPConn),
	}
}

func (p *StatPrinter) AddConn(key string, conn *net.TCPConn) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.conns[key] = conn
}

func (p *StatPrinter) DelConn(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.conns, key)
}

func (p *StatPrinter) Start() {
	ticker := time.NewTicker(p.duration)
	defer ticker.Stop()

	for range ticker.C {
		p.print()
	}
}

func (p *StatPrinter) Stop() {
	p.print()
}

func (p *StatPrinter) buildRows() [][]string {
	var keys []string
	infos := make(map[string]*TcpInfo)
	p.lock.RLock()
	prev := p.prev
	for k, v := range p.conns {
		info, err := GetTcpInfo(v)
		if err != nil {
			display.PrintfWithTime("GetTcpInfo: %v\n", err)
			continue
		}

		keys = append(keys, k)
		infos[k] = info
	}
	p.prev = infos
	p.lock.RUnlock()

	var rows [][]string
	now := time.Now().Format(display.TimeFormat)
	sort.Strings(keys)
	for _, k := range keys {
		v, ok := infos[k]
		if !ok {
			continue
		}

		var rate string
		pinfo, ok := prev[k]
		if ok {
			rate = fmt.Sprintf("%.2f", GetRetransRate(pinfo, v))
		} else {
			rate = "-"
		}
		rtt, rttv := v.GetRTT()
		rows = append(rows, []string{now, k, rate, fmt.Sprint(rtt), fmt.Sprint(rttv)})
	}

	return rows
}

func (p *StatPrinter) print() {
	rows := p.buildRows()
	if len(rows) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Timestamp", "Connection", "RetransRate(%)", "RTT(ms)", "RTT/Variance(ms)"})
	table.Bulk(rows)
	table.Render()
	fmt.Println()
}
