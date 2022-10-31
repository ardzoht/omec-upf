// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"fmt"

	"github.com/omec-project/upf-epc/pfcpiface/metrics"
)

type PacketForwardingRules struct {
	Pdrs []Pdr
	Fars []Far
	Qers []Qer
	Urrs []Urr
}

// PFCPSession implements one PFCP session.
type PFCPSession struct {
	localSEID  uint64
	remoteSEID uint64
	metrics    *metrics.Session
	PacketForwardingRules
	// used to store session <-> UE Address mapping
	// which is needed to efficiently find UE address for UL PDRs in the PFCP messages.
	UeAddress uint32
}

func (p PacketForwardingRules) String() string {
	return fmt.Sprintf("PDRs=%v, FARs=%v, QERs=%v, URRs=%v", p.Pdrs, p.Fars, p.Qers, p.Urrs)
}

// NewPFCPSession allocates an session with ID.
func (pConn *PFCPConn) NewPFCPSession(rseid uint64) (PFCPSession, bool) {
	for i := 0; i < pConn.maxRetries; i++ {
		lseid := pConn.rng.Uint64()
		// Check if it already exists
		if _, ok := pConn.store.GetSession(lseid); ok {
			continue
		}

		s := PFCPSession{
			localSEID:  lseid,
			remoteSEID: rseid,
			PacketForwardingRules: PacketForwardingRules{
				Pdrs: make([]Pdr, 0, MaxItems),
				Fars: make([]Far, 0, MaxItems),
				Qers: make([]Qer, 0, MaxItems),
				Urrs: make([]Urr, 0, MaxItems),
			},
		}
		s.metrics = metrics.NewSession(pConn.nodeID.remote)

		// Metrics update
		pConn.SaveSessions(s.metrics)

		return s, true
	}

	return PFCPSession{}, false
}

// RemoveSession removes session using lseid.
func (pConn *PFCPConn) RemoveSession(session PFCPSession) {
	// Metrics update
	session.metrics.Delete()
	pConn.SaveSessions(session.metrics)

	if err := pConn.store.DeleteSession(session.localSEID); err != nil {
		log.Errorf("Failed to delete PFCP session from store: %v", err)
	}
}
