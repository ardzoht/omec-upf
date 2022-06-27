// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"google.golang.org/grpc"
)

type ebpf struct {
	conn             *grpc.ClientConn
	endMarkerSocket  net.Conn
	notifyBessSocket net.Conn
	endMarkerChan    chan []byte
}

func (d *ebpf) IsConnected(accessIP *net.IP) bool {
	// TODO(ardzoht): Add connection check for server to DP service
    // Defaulting to true for now to test with PFCPsim
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
	panic("Not implemented")
}

func (d *ebpf) SessionStats(pc *PfcpNodeCollector, ch chan<- prometheus.Metric) (err error) {
	panic("Not implemented")
}

func (d *ebpf) SummaryLatencyJitter(uc *upfCollector, ch chan<- prometheus.Metric) {
	panic("Not implemented")
}
