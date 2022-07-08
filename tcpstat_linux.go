package main

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

const lostRateThreshold = 1e-6

type TcpInfo struct {
	State         uint8  `json:"state"`
	CAState       uint8  `json:"ca_state"`
	Retransmits   uint8  `json:"retransmits"`
	Probes        uint8  `json:"probes"`
	Backoff       uint8  `json:"backoff"`
	Options       uint8  `json:"options"`
	WScale        uint8  `json:"w_scale"`
	AppLimited    uint8  `json:"app_limited"`
	RTO           uint32 `json:"rto"`
	ATO           uint32 `json:"ato"`
	SndMSS        uint32 `json:"snd_mss"`
	RcvMSS        uint32 `json:"rcv_mss"`
	Unacked       uint32 `json:"unacked"`
	Sacked        uint32 `json:"sacked"`
	Lost          uint32 `json:"lost"`
	Retrans       uint32 `json:"retrans"`
	Fackets       uint32 `json:"f_ackets"`
	LastDataSent  uint32 `json:"last_data_sent"`
	LastAckSent   uint32 `json:"last_ack_sent"`
	LastDataRecv  uint32 `json:"last_data_recv"`
	LastAckRecv   uint32 `json:"last_ack_recv"`
	PathMTU       uint32 `json:"p_mtu"`
	RcvSsThresh   uint32 `json:"rcv_ss_thresh"`
	RTT           uint32 `json:"rtt"`
	RTTVar        uint32 `json:"rtt_var"`
	SndSsThresh   uint32 `json:"snd_ss_thresh"`
	SndCwnd       uint32 `json:"snd_cwnd"`
	AdvMSS        uint32 `json:"adv_mss"`
	Reordering    uint32 `json:"reordering"`
	RcvRTT        uint32 `json:"rcv_rtt"`
	RcvSpace      uint32 `json:"rcv_space"`
	TotalRetrans  uint32 `json:"total_retrans"`
	PacingRate    int64  `json:"pacing_rate"`
	MaxPacingRate int64  `json:"max_pacing_rate"`
	BytesAcked    int64  `json:"bytes_acked"`
	BytesReceived int64  `json:"bytes_received"`
	SegsOut       int32  `json:"segs_out"`
	SegsIn        int32  `json:"segs_in"` // RFC4898 tcpEStatsPerfSegsIn
	NotSentBytes  uint32 `json:"notsent_bytes"`
	MinRTT        uint32 `json:"min_rtt"`
	DataSegsIn    uint32 `json:"data_segs_in"`  // RFC4898 tcpEStatsDataSegsIn
	DataSegsOut   uint32 `json:"data_segs_out"` // RFC4898 tcpEStatsDataSegsOut
	DeliveryRate  int64  `json:"delivery_rate"`
	BusyTime      int64  `json:"busy_time"`       // Time (usec) busy sending data
	RWndLimited   int64  `json:"r_wnd_limited"`   // Time (usec) limited by receive window
	SndBufLimited int64  `json:"snd_buf_limited"` // Time (usec) limited by send buffer
	Delivered     uint32 `json:"delivered"`
	DeliveredCE   uint32 `json:"delivered_ce"`
	BytesSent     int64  `json:"bytes_sent"`    // RFC4898 tcpEStatsPerfHCDataOctetsOut
	BytesRetrans  int64  `json:"bytes_retrans"` // RFC4898 tcpEStatsPerfOctetsRetrans
	DSackDups     uint32 `json:"d_sack_dups"`   // RFC4898 tcpEStatsStackDSACKDups
	ReordSeen     uint32 `json:"reord_seen"`    // reordering events seen
}

func GetTcpInfo(tcpConn *net.TCPConn) (*TcpInfo, error) {
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("error getting raw connection. error: %v", err)
	}

	tcpInfo := TcpInfo{}
	size := unsafe.Sizeof(tcpInfo)

	var errno syscall.Errno
	err = rawConn.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall6(syscall.SYS_GETSOCKOPT, fd, syscall.SOL_TCP, syscall.TCP_INFO,
			uintptr(unsafe.Pointer(&tcpInfo)), uintptr(unsafe.Pointer(&size)), 0)
	})
	if err != nil {
		return nil, fmt.Errorf("conn control failed, error: %v", err)
	}
	if errno != 0 {
		return nil, fmt.Errorf("syscall failed, errno: %d", errno)
	}

	return &tcpInfo, nil
}

// GetRetransRate returns the percent of lost packets.
func GetRetransRate(preTi, ti *TcpInfo) float64 {
	if preTi == nil {
		return 0
	}

	bytesDelta := ti.BytesSent - preTi.BytesSent
	var lostRate float64
	if bytesDelta != 0 {
		lostRate = 100 * float64(ti.BytesRetrans-preTi.BytesRetrans) / float64(bytesDelta)
		if lostRate < lostRateThreshold {
			lostRate = 0
		}
	}
	if lostRate < 0 {
		return 0
	} else if lostRate > 1 {
		return 1
	}

	return lostRate
}

// GetRTT returns Round Trip Time in milliseconds.
func (ti *TcpInfo) GetRTT() (uint32, uint32) {
	return ti.RTT / 1000, ti.RTTVar / 1000
}
