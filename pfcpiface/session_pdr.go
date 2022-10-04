// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

import (
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

// Release allocated IPs.
func releaseAllocatedIPs(ippool *IPPool, session *PFCPSession) error {
	log.Info("release allocated IP")

	// Check if we allocated an UE IP for this session and delete it.
	for _, pdr := range session.Pdrs {
		if (pdr.AllocIPFlag) && (pdr.SrcIface == core) {
			var ueIP net.IP = int2ip(pdr.UeAddress)

			log.Debugf("Releasing IP %v of session %v", ueIP, session.localSEID)

			return ippool.DeallocIP(session.localSEID)
		}
	}

	return nil
}

// Release allocated teids.
func releaseAllocatedTEIDs(generator *IDAllocator, session *PFCPSession) {
	log.Infof("Release allocate TEIDs for session: %v", session.localSEID)

	for _, pdr := range session.Pdrs {
		log.Debugf("Releasing TEID %v of session %v", pdr.TunnelTEID, session.localSEID)
		generator.Free(pdr.TunnelTEID)
	}
}

func findChoosedPdr(session *PFCPSession, chooseID uint8) *Pdr {
	log.Debugf("Find choosed PDR with TEID allocation")

	for _, pdr := range session.Pdrs {
		if pdr.ChooseIDFlag && pdr.ChooseID == chooseID {
			return &pdr
		}
	}
	return nil
}

func addPdrInfo(msg *message.SessionEstablishmentResponse,
	session *PFCPSession) {
	log.Info("Add PDRs with UPF alloc teids/IPs to Establishment response")

	for _, pdr := range session.Pdrs {
		createdPDR := ie.NewCreatedPDR()
		createdPDR.Add(ie.NewPDRID(uint16(pdr.PdrID)))
		needAlloc := false

		if (pdr.AllocIPFlag) && (pdr.SrcIface == core) {
			needAlloc = true

			var (
				flags uint8  = 0x02
				ueIP  net.IP = int2ip(pdr.UeAddress)
			)

			log.Info("pdrID: %v Adding ueIP : %v", pdr.PdrID, ueIP.String())
			createdPDR.Add(ie.NewUEIPAddress(flags, ueIP.String(), "", 0, 0))
		}

		if pdr.AllocTEIDFlag {
			needAlloc = true

			var (
				flags  uint8  = 0x01 // IPv4 flag is present
				teidIP net.IP = int2ip(pdr.TunnelIP4Dst)
			)

			log.Info("pdrID: %v Adding TEID : %v", pdr.PdrID, pdr.TunnelTEID)
			createdPDR.Add(ie.NewFTEID(flags, pdr.TunnelTEID, teidIP, nil, pdr.ChooseID))
		}

		if needAlloc {
			msg.CreatedPDR = append(msg.CreatedPDR, createdPDR)
		}
	}
}

// CreatePDR appends pdr to existing list of PDRs in the session.
func (s *PFCPSession) CreatePDR(p Pdr) {
	if p.IsDownlink() && p.UeAddress != 0 {
		s.UeAddress = p.UeAddress
	}

	s.Pdrs = append(s.Pdrs, p)
}

// UpdatePDR updates existing pdr in the session.
func (s *PFCPSession) UpdatePDR(p Pdr) error {
	if p.IsDownlink() && p.UeAddress != 0 {
		s.UeAddress = p.UeAddress
	}

	for idx, v := range s.Pdrs {
		if v.PdrID == p.PdrID {
			s.Pdrs[idx] = p
			return nil
		}
	}

	return ErrNotFound("PDR")
}

// RemovePDR removes pdr from existing list of PDRs in the session.
func (s *PFCPSession) RemovePDR(id uint32) (*Pdr, error) {
	for idx, v := range s.Pdrs {
		if v.PdrID == id {
			s.Pdrs = append(s.Pdrs[:idx], s.Pdrs[idx+1:]...)
			return &v, nil
		}
	}

	return nil, ErrNotFound("PDR")
}
