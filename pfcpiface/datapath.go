// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"net"

	"github.com/prometheus/client_golang/prometheus"
)

type UpfMsgType int

const (
	upfMsgTypeAdd UpfMsgType = iota
	upfMsgTypeMod
	upfMsgTypeDel
	upfMsgTypeClear
)

func (u UpfMsgType) String() string {
	if u == upfMsgTypeAdd {
		return "add"
	} else if u == upfMsgTypeMod {
		return "modify"
	} else if u == upfMsgTypeDel {
		return "delete" //nolint
	} else if u == upfMsgTypeClear {
		return "clear"
	} else {
		return "unknown"
	}
}

type Datapath interface {
	/* Close any pending sessions */
	Exit()
	/* setup internal parameters and channel with datapath */
	SetUpfInfo(u *Upf, conf *Conf)
	/* set up slice info */
	AddSliceInfo(sliceInfo *SliceInfo) error
	/* write endMarker to datapath */
	SendEndMarkers(endMarkerList *[]EndMarker) error
	/* write pdr/far/qer to datapath */
	// "master" function to send create/update/delete messages to UPF.
	// "new" PacketForwardingRules are only used for update messages to UPF.
	// TODO: we should have better CRUD API, with a single function per message type.
	SendMsgToUPF(method UpfMsgType, all PacketForwardingRules, new PacketForwardingRules) uint8
	/* check of communication channel to datapath is setup */
	IsConnected(accessIP *net.IP) bool
	SummaryLatencyJitter(uc *UpfCollector, ch chan<- prometheus.Metric)
	PortStats(uc *UpfCollector, ch chan<- prometheus.Metric)
	SessionStats(pc *PfcpNodeCollector, ch chan<- prometheus.Metric) error
}
