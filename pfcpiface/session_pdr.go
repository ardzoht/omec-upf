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

func addPdrInfo(msg *message.SessionEstablishmentResponse,
	session *PFCPSession) {
	log.Info("Add PDRs with UPF alloc IPs to Establishment response")

	for _, pdr := range session.Pdrs {
		if (pdr.AllocIPFlag) && (pdr.SrcIface == core) {
			log.Info("pdrID : ", pdr.PdrID)

			var (
				flags uint8  = 0x02
				ueIP  net.IP = int2ip(pdr.UeAddress)
			)

			log.Info("ueIP : ", ueIP.String())
			msg.CreatedPDR = append(msg.CreatedPDR,
				ie.NewCreatedPDR(
					ie.NewPDRID(uint16(pdr.PdrID)),
					ie.NewUEIPAddress(flags, ueIP.String(), "", 0, 0),
				))
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
