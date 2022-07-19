// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"net"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wmnsk/go-pfcp/ie"

	"google.golang.org/grpc"
)

type Ebpf struct {
	conn *grpc.ClientConn
}

func (d *Ebpf) IsConnected(accessIP *net.IP) bool {
	// TODO(ardzoht): Add connection check for server to DP service
	// Defaulting to true for now to test with PFCPsim
	return true
}

func (d *Ebpf) Exit() {
	log.Info("Shutting down datapath...")
}

// SetUpfInfo is only called at pfcp-agent's startup
func (d *Ebpf) SetUpfInfo(u *Upf, conf *Conf) {
	log.Info("Setting UPF config...")
}

func (d *Ebpf) SendMsgToUPF(
	method UpfMsgType, rules PacketForwardingRules, updated PacketForwardingRules) uint8 {
	var cause uint8 = ie.CauseRequestAccepted

	pdrs := rules.Pdrs
	fars := rules.Fars
	qers := rules.Qers

	if method == UpfMsgTypeMod {
		pdrs = updated.Pdrs
		fars = updated.Fars
		qers = updated.Qers
	}

	for _, pdr := range pdrs {
		log.Debugf("%s %s", method, pdr)
	}

	for _, far := range fars {
		log.Debugf("%s %s", method, far)
	}

	for _, qer := range qers {
		log.Debugf("%s %s", method, qer)
	}

	return cause
}

func (d *Ebpf) AddSliceInfo(sliceInfo *SliceInfo) error {
	panic("Not implemented")
}

func (d *Ebpf) PortStats(uc *UpfCollector, ch chan<- prometheus.Metric) {
	panic("Not implemented")
}

func (d *Ebpf) SendEndMarkers(endMarkerList *[]EndMarker) error {
	panic("Not implemented")
}

func (d *Ebpf) SessionStats(pc *PfcpNodeCollector, ch chan<- prometheus.Metric) (err error) {
	panic("Not implemented")
}

func (d *Ebpf) SummaryLatencyJitter(uc *UpfCollector, ch chan<- prometheus.Metric) {
	panic("Not implemented")
}
