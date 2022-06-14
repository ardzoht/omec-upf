// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type ebpf struct {
	conn             *grpc.ClientConn
	endMarkerSocket  net.Conn
	notifyBessSocket net.Conn
	endMarkerChan    chan []byte
}

func (d *ebpf) IsConnected(accessIP *net.IP) bool {
	if (d.conn == nil) || (d.conn.GetState() != connectivity.Ready) {
		return false
	}

	return true
}

func (d *ebpf) Exit() {
	log.Println("Shutting down datapath...")
}

// SetUpfInfo is only called at pfcp-agent's startup
func (d *ebpf) SetUpfInfo(u *upf, conf *Conf) {
	log.Println("Setting UPF config...")
}

func (d *ebpf) SendMsgToUPF(
	method upfMsgType, rules PacketForwardingRules, updated PacketForwardingRules) uint8 {
	panic("Not implemented")
}

func (d *ebpf) AddSliceInfo(sliceInfo *SliceInfo) error {
	panic("Not implemented")
}

func (d *ebpf) PortStats(uc *upfCollector, ch chan<- prometheus.Metric) {
	panic("Not implemented")
}

func (d *ebpf) SendEndMarkers(endMarkerList *[][]byte) error {
	for _, eMarker := range *endMarkerList {
		d.endMarkerChan <- eMarker
	}

	return nil
}

func (d *ebpf) SessionStats(pc *PfcpNodeCollector, ch chan<- prometheus.Metric) (err error) {
	panic("Not implemented")
}

func (d *ebpf) SummaryLatencyJitter(uc *upfCollector, ch chan<- prometheus.Metric) {
	panic("Not implemented")
}
