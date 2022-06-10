// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/wmnsk/go-pfcp/ie"
)

const (
	// SockAddr : Unix Socket path to read bess notification from.
	SockAddr = "/tmp/notifycp"
	// PfcpAddr : Unix Socket path to send end marker packet.
	PfcpAddr = "/tmp/pfcpport"
)


type ebpf struct {
	endMarkerSocket  net.Conn
	notifyBessSocket net.Conn
	endMarkerChan    chan []byte
}

func (b *bess) IsConnected(accessIP *net.IP) bool {
	if (b.conn == nil) || (b.conn.GetState() != connectivity.Ready) {
		return false
	}

	return true
}

func (b *ebpf) Exit() {
	b.conn.Close()
}


// SetUpfInfo is only called at pfcp-agent's startup
func (b *bess) SetUpfInfo(u *upf, conf *Conf) {
	var err error

	log.Println("SetUpfInfo")

	b.endMarkerChan = make(chan []byte, 1024)

	if conf.EnableNotifyBess {
		notifySockAddr := conf.NotifySockAddr
		if notifySockAddr == "" {
			notifySockAddr = SockAddr
		}

		b.notifyBessSocket, err = net.Dial("unixpacket", notifySockAddr)
		if err != nil {
			log.Println("dial error:", err)
			return
		}

		go b.notifyListen(u.reportNotifyChan)
	}

	if conf.EnableEndMarker {
		pfcpCommAddr := conf.EndMarkerSockAddr
		if pfcpCommAddr == "" {
			pfcpCommAddr = PfcpAddr
		}

		b.endMarkerSocket, err = net.Dial("unixpacket", pfcpCommAddr)
		if err != nil {
			log.Println("dial error:", err)
			return
		}

		log.Println("Starting end marker loop")

		go b.endMarkerSendLoop(b.endMarkerChan)
	}
}

