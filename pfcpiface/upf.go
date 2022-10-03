// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"math"
	"net"
	"time"

	"github.com/Showmax/go-fqdn"
)

// QosConfigVal : Qos configured value.
type QosConfigVal struct {
	cbs              uint32
	pbs              uint32
	ebs              uint32
	burstDurationMs  uint32
	schedulePriority uint32
}

type SliceInfo struct {
	name         string
	uplinkMbr    uint64
	downlinkMbr  uint64
	ulBurstBytes uint64
	dlBurstBytes uint64
	ueResList    []UeResource
}

type UeResource struct {
	name string
	dnn  string
}

type Upf struct {
	enableUeIPAlloc   bool
	enableEndMarker   bool
	enableFlowMeasure bool
	accessIface       string
	coreIface         string
	ippoolCidr        string
	AccessIP          net.IP
	CoreIP            net.IP
	NodeID            string

	ippool        *IPPool
	teidAllocator *IDAllocator

	peers            []string
	dnn              string
	ReportNotifyChan chan uint64
	sliceInfo        *SliceInfo
	readTimeout      time.Duration

	Datapath
	maxReqRetries uint8
	respTimeout   time.Duration
	enableHBTimer bool
	hbInterval    time.Duration
}

// to be replaced with go-pfcp structs
// Don't change these values.
const (
	tunnelGTPUPort = 2152

	// src-iface consts.
	core   = 0x2
	access = 0x1

	// far-id specific directions.
	n3 = 0x0
	n6 = 0x1
	n9 = 0x2
)

func (u *Upf) isConnected() bool {
	return u.Datapath.IsConnected(&u.AccessIP)
}

func (u *Upf) addSliceInfo(sliceInfo *SliceInfo) error {
	if sliceInfo == nil {
		return ErrInvalidArgument("sliceInfo", sliceInfo)
	}

	u.sliceInfo = sliceInfo

	return u.Datapath.AddSliceInfo(sliceInfo)
}

func NewUPF(conf *Conf, fp Datapath) *Upf {
	var (
		err    error
		nodeID string
	)

	nodeID = conf.CPIface.NodeID
	if conf.CPIface.UseFQDN && nodeID == "" {
		nodeID, err = fqdn.FqdnHostname()
		if err != nil {
			log.Fatal("Unable to get hostname", err)
		}
	}

	// TODO: Delete this once CI config is fixed
	if nodeID != "" {
		hosts, err := net.LookupHost(nodeID)
		if err != nil {
			log.Fatalf("Unable to resolve hostname %v %v", nodeID, err)
		}

		nodeID = hosts[0]
	}

	u := &Upf{
		enableUeIPAlloc:   conf.CPIface.EnableUeIPAlloc,
		enableEndMarker:   conf.EnableEndMarker,
		enableFlowMeasure: conf.EnableFlowMeasure,
		accessIface:       conf.AccessIface.IfName,
		coreIface:         conf.CoreIface.IfName,
		ippoolCidr:        conf.CPIface.UEIPPool,
		NodeID:            nodeID,
		Datapath:          fp,
		dnn:               conf.CPIface.Dnn,
		peers:             conf.CPIface.Peers,
		ReportNotifyChan:  make(chan uint64, 1024),
		maxReqRetries:     conf.MaxReqRetries,
		enableHBTimer:     conf.EnableHBTimer,
		readTimeout:       time.Second * time.Duration(conf.ReadTimeout),
	}

	if len(conf.CPIface.Peers) > 0 {
		u.peers = make([]string, len(conf.CPIface.Peers))
		nc := copy(u.peers, conf.CPIface.Peers)

		if nc == 0 {
			log.Warn("Failed to parse cpiface peers, PFCP Agent will not initiate connection to N4 peers.")
		}
	}

	if !conf.EnableP4rt {
		u.AccessIP, err = GetUnicastAddressFromInterface(conf.AccessIface.IfName)
		if err != nil {
			log.Error(err)
			return nil
		}

		u.CoreIP, err = GetUnicastAddressFromInterface(conf.CoreIface.IfName)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	u.respTimeout, err = time.ParseDuration(conf.RespTimeout)
	if err != nil {
		log.Fatal("Unable to parse resp_timeout")
	}

	if u.enableHBTimer {
		if conf.HeartBeatInterval != "" {
			u.hbInterval, err = time.ParseDuration(conf.HeartBeatInterval)
			if err != nil {
				log.Fatal("Unable to parse heart_beat_interval")
			}
		}
	}

	if u.enableUeIPAlloc {
		u.ippool, err = NewIPPool(u.ippoolCidr)
		if err != nil {
			log.Fatal("ip pool init failed", err)
		}
	}

	u.teidAllocator = NewIDAllocator(1, math.MaxUint32)

	u.Datapath.SetUpfInfo(u, conf)

	return u
}

