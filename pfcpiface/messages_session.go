// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Intel Corporation

package pfcpiface

import (
	"errors"
	"net"
	"strings"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"go.uber.org/zap"
)

// errors
var (
	ErrWriteToDatapath = errors.New("write to datapath failed")
	ErrAssocNotFound   = errors.New("no association found for NodeID")
	ErrAllocateSession = errors.New("unable to allocate new PFCP session")
)

func (pConn *PFCPConn) handleSessionEstablishmentRequest(msg message.Message) (message.Message, error) {
	upf := pConn.upf

	sereq, ok := msg.(*message.SessionEstablishmentRequest)
	if !ok {
		return nil, errUnmarshal(errMsgUnexpectedType)
	}

	errUnmarshalReply := func(err error, offendingIE *ie.IE) (message.Message, error) {
		// Build response message
		pfdres := message.NewSessionEstablishmentResponse(0,
			0,
			0,
			sereq.SequenceNumber,
			0,
			ie.NewCause(ie.CauseRequestRejected),
			offendingIE,
		)

		return pfdres, errUnmarshal(err)
	}

	nodeID, err := sereq.NodeID.NodeID()
	if err != nil {
		return errUnmarshalReply(err, sereq.NodeID)
	}

	/* Read fseid from the IE */
	fseid, err := sereq.CPFSEID.FSEID()
	if err != nil {
		return errUnmarshalReply(err, sereq.CPFSEID)
	}

	remoteSEID := fseid.SEID
	fseidIP := ip2int(fseid.IPv4Address)

	errProcessReply := func(err error, cause uint8) (message.Message, error) {
		// Build response message
		seres := message.NewSessionEstablishmentResponse(0, /* MO?? <-- what's this */
			0,                    /* FO <-- what's this? */
			remoteSEID,           /* seid */
			sereq.SequenceNumber, /* seq # */
			0,                    /* priority */
			pConn.nodeID.localIE,
			ie.NewCause(cause),
		)

		return seres, errProcess(err)
	}

	if strings.Compare(nodeID, pConn.nodeID.remote) != 0 {
		log.Warnf("Association not found for Establishment request with nodeID: %v, Association NodeID: %v",
			nodeID, pConn.nodeID.remote)
		return errProcessReply(ErrAssocNotFound, ie.CauseNoEstablishedPFCPAssociation)
	}

	session, ok := pConn.NewPFCPSession(remoteSEID)
	if !ok {
		return errProcessReply(ErrAllocateSession,
			ie.CauseNoResourcesAvailable)
	}

	addPDRs := make([]Pdr, 0, MaxItems)
	addFARs := make([]Far, 0, MaxItems)
	addQERs := make([]Qer, 0, MaxItems)

	for _, cPDR := range sereq.CreatePDR {
		var p Pdr
		if err := p.parsePDR(cPDR, pConn.appPFDs, upf.ippool, upf, &session); err != nil {
			return errProcessReply(err, ie.CauseRequestRejected)
		}

		p.FseidIP = fseidIP
		session.CreatePDR(p)
		addPDRs = append(addPDRs, p)
	}

	for _, cFAR := range sereq.CreateFAR {
		var f Far
		if err := f.parseFAR(cFAR, session.localSEID, upf, create); err != nil {
			return errProcessReply(err, ie.CauseRequestRejected)
		}

		f.FseidIP = fseidIP
		session.CreateFAR(f)
		addFARs = append(addFARs, f)
	}

	for _, cQER := range sereq.CreateQER {
		var q Qer
		if err := q.parseQER(cQER, session.localSEID); err != nil {
			return errProcessReply(err, ie.CauseRequestRejected)
		}

		q.FseidIP = fseidIP
		session.CreateQER(q)
		addQERs = append(addQERs, q)
	}

	session.MarkSessionQer(session.Qers)
	// FIXME: since PacketForwardingRules doesn't store pointers,
	//  we must also mark session QERs in addQERs.
	//  We need a kind of refactoring to clean it up.
	session.MarkSessionQer(addQERs)

	// session.PacketForwardingRules stores all PFCP rules that has been installed so far,
	// while 'updated' stores only the PFCP rules that have been provided in this particular message.
	updated := PacketForwardingRules{
		Pdrs: addPDRs,
		Fars: addFARs,
		Qers: addQERs,
	}

	cause := upf.SendMsgToUPF(UpfMsgTypeAdd, session, updated)
	if cause == ie.CauseRequestRejected {
		pConn.RemoveSession(session)
		return errProcessReply(ErrWriteToDatapath,
			ie.CauseRequestRejected)
	}

	err = pConn.store.PutSession(session)
	if err != nil {
		log.Errorf("Failed to put PFCP session to store: %v", err)
	}

	var localFSEID *ie.IE

	localIP := pConn.LocalAddr().(*net.UDPAddr).IP
	if localIP.To4() != nil {
		localFSEID = ie.NewFSEID(session.localSEID, localIP, nil)
	} else {
		localFSEID = ie.NewFSEID(session.localSEID, nil, localIP)
	}

	log.Debugw("Sending session establishment response:",
		zap.Uint64("Local SEID:", session.localSEID),
		zap.Uint64("Remote SEID:", remoteSEID))

	// Build response message
	seres := message.NewSessionEstablishmentResponse(0, /* MO?? <-- what's this */
		0,                                    /* FO <-- what's this? */
		session.remoteSEID,                   /* seid */
		sereq.SequenceNumber,                 /* seq # */
		0,                                    /* priority */
		pConn.nodeID.localIE,                 /* node id */
		ie.NewCause(ie.CauseRequestAccepted), /* accept it blindly for the time being */
		localFSEID,
	)

	addPdrInfo(seres, &session)

	return seres, nil
}

func (pConn *PFCPConn) handleSessionModificationRequest(msg message.Message) (message.Message, error) {
	upf := pConn.upf

	smreq, ok := msg.(*message.SessionModificationRequest)
	if !ok {
		return nil, errUnmarshal(errMsgUnexpectedType)
	}

	var remoteSEID uint64

	sendError := func(err error) (message.Message, error) {
		log.Error(err)

		smres := message.NewSessionModificationResponse(0, /* MO?? <-- what's this */
			0,                                    /* FO <-- what's this? */
			remoteSEID,                           /* seid */
			smreq.SequenceNumber,                 /* seq # */
			0,                                    /* priority */
			ie.NewCause(ie.CauseRequestRejected), /* accept it blindly for the time being */
		)

		return smres, err
	}

	localSEID := smreq.SEID()

	session, ok := pConn.store.GetSession(localSEID)
	if !ok {
		return sendError(ErrNotFoundWithParam("PFCP session", "localSEID", localSEID))
	}

	var fseidIP uint32

	if smreq.CPFSEID != nil {
		fseid, err := smreq.CPFSEID.FSEID()
		if err == nil {
			session.remoteSEID = fseid.SEID
			fseidIP = ip2int(fseid.IPv4Address)

			log.Debug("Updated FSEID from session modification request")
		}
	}

	remoteSEID = session.remoteSEID

	addPDRs := make([]Pdr, 0, MaxItems)
	addFARs := make([]Far, 0, MaxItems)
	addQERs := make([]Qer, 0, MaxItems)

	for _, cPDR := range smreq.CreatePDR {
		var p Pdr
		if err := p.parsePDR(cPDR, pConn.appPFDs, upf.ippool, upf, &session); err != nil {
			return sendError(err)
		}

		p.FseidIP = fseidIP

		session.CreatePDR(p)
		addPDRs = append(addPDRs, p)
	}

	for _, cFAR := range smreq.CreateFAR {
		var f Far
		if err := f.parseFAR(cFAR, localSEID, upf, create); err != nil {
			return sendError(err)
		}

		f.FseidIP = fseidIP

		session.CreateFAR(f)
		addFARs = append(addFARs, f)
	}

	for _, cQER := range smreq.CreateQER {
		var q Qer
		if err := q.parseQER(cQER, localSEID); err != nil {
			return sendError(err)
		}

		q.FseidIP = fseidIP

		session.CreateQER(q)
		addQERs = append(addQERs, q)
	}

	updated := PacketForwardingRules{
		Pdrs: addPDRs,
		Fars: addFARs,
		Qers: addQERs,
	}

	cause := upf.SendMsgToUPF(UpfMsgTypeAdd, session, updated)
	if cause == ie.CauseRequestRejected {
		return sendError(ErrWriteToDatapath)
	}

	updatePDRs := make([]Pdr, 0, MaxItems)
	updateFARs := make([]Far, 0, MaxItems)
	updateQERs := make([]Qer, 0, MaxItems)
	endMarkerList := make([]EndMarker, 0, MaxItems)

	// maintain a copy of session before updating for session modification
	// datapath could leverage the previous session info to update forwarding rules
	oldSession := PFCPSession{
		UeAddress:  session.UeAddress,
		localSEID:  session.localSEID,
		remoteSEID: session.remoteSEID,
	}

	oldSession.Pdrs = make([]Pdr, len(session.Pdrs))
	oldSession.Fars = make([]Far, len(session.Fars))
	oldSession.Qers = make([]Qer, len(session.Qers))
	copy(oldSession.Pdrs, session.Pdrs)
	copy(oldSession.Fars, session.Fars)
	copy(oldSession.Qers, session.Qers)

	for _, uPDR := range smreq.UpdatePDR {
		var (
			p   Pdr
			err error
		)

		if err = p.parsePDR(uPDR, pConn.appPFDs, upf.ippool, upf, &session); err != nil {
			return sendError(err)
		}

		p.FseidIP = fseidIP

		err = session.UpdatePDR(p)
		if err != nil {
			log.Error("session PDR update failed ", err)
			continue
		}

		updatePDRs = append(updatePDRs, p)
	}

	for _, uFAR := range smreq.UpdateFAR {
		var (
			f   Far
			err error
		)

		if err = f.parseFAR(uFAR, localSEID, upf, update); err != nil {
			return sendError(err)
		}

		f.FseidIP = fseidIP

		err = session.UpdateFAR(&f, &endMarkerList)
		if err != nil {
			log.Error("session PDR update failed ", err)
			continue
		}

		updateFARs = append(updateFARs, f)
	}

	for _, uQER := range smreq.UpdateQER {
		var (
			q   Qer
			err error
		)

		if err = q.parseQER(uQER, localSEID); err != nil {
			return sendError(err)
		}

		q.FseidIP = fseidIP

		err = session.UpdateQER(q)
		if err != nil {
			log.Error("session QER update failed ", err)
			continue
		}

		updateQERs = append(updateQERs, q)
	}

	session.MarkSessionQer(session.Qers)
	// FIXME: since PacketForwardingRules doesn't store pointers,
	//  we must also mark session QERs in addQERs.
	//  We need a kind of refactoring to clean it up.
	session.MarkSessionQer(addQERs)

	updated = PacketForwardingRules{
		Pdrs: updatePDRs,
		Fars: updateFARs,
		Qers: updateQERs,
	}

	// Send the session before updating to datapath
	cause = upf.SendMsgToUPF(UpfMsgTypeMod, oldSession, updated)
	if cause == ie.CauseRequestRejected {
		return sendError(ErrWriteToDatapath)
	}

	if upf.enableEndMarker {
		err := upf.SendEndMarkers(&endMarkerList)
		if err != nil {
			log.Error("Sending End Markers Failed : ", err)
		}
	}

	delPDRs := make([]Pdr, 0, MaxItems)
	delFARs := make([]Far, 0, MaxItems)
	delQERs := make([]Qer, 0, MaxItems)

	for _, rPDR := range smreq.RemovePDR {
		pdrID, err := rPDR.PDRID()
		if err != nil {
			return sendError(err)
		}

		p, err := session.RemovePDR(uint32(pdrID))
		if err != nil {
			return sendError(err)
		}
		upf.teidAllocator.Free(p.TunnelTEID)

		delPDRs = append(delPDRs, *p)
	}

	for _, dFAR := range smreq.RemoveFAR {
		farID, err := dFAR.FARID()
		if err != nil {
			return sendError(err)
		}

		f, err := session.RemoveFAR(farID)
		if err != nil {
			return sendError(err)
		}

		delFARs = append(delFARs, *f)
	}

	for _, dQER := range smreq.RemoveQER {
		qerID, err := dQER.QERID()
		if err != nil {
			return sendError(err)
		}

		q, err := session.RemoveQER(qerID)
		if err != nil {
			return sendError(err)
		}

		delQERs = append(delQERs, *q)
	}

	deleted := PacketForwardingRules{
		Pdrs: delPDRs,
		Fars: delFARs,
		Qers: delQERs,
	}

	cause = upf.SendMsgToUPF(UpfMsgTypeDel, session, deleted)
	if cause == ie.CauseRequestRejected {
		return sendError(ErrWriteToDatapath)
	}

	err := pConn.store.PutSession(session)
	if err != nil {
		log.Errorf("Failed to put PFCP session to store: %v", err)
	}

	log.Debugw("Sending session modification response:",
		zap.Uint64("Local SEID:", localSEID),
		zap.Uint64("Remote SEID:", remoteSEID))
	// Build response message
	smres := message.NewSessionModificationResponse(0, /* MO?? <-- what's this */
		0,                                    /* FO <-- what's this? */
		remoteSEID,                           /* seid */
		smreq.SequenceNumber,                 /* seq # */
		0,                                    /* priority */
		ie.NewCause(ie.CauseRequestAccepted), /* accept it blindly for the time being */
	)

	return smres, nil
}

func (pConn *PFCPConn) handleSessionDeletionRequest(msg message.Message) (message.Message, error) {
	upf := pConn.upf

	sdreq, ok := msg.(*message.SessionDeletionRequest)
	if !ok {
		return nil, errUnmarshal(errMsgUnexpectedType)
	}

	sendError := func(err error) (message.Message, error) {
		smres := message.NewSessionDeletionResponse(0, /* MO?? <-- what's this */
			0,                                    /* FO <-- what's this? */
			0,                                    /* seid */
			sdreq.SequenceNumber,                 /* seq # */
			0,                                    /* priority */
			ie.NewCause(ie.CauseRequestRejected), /* accept it blindly for the time being */
		)

		return smres, err
	}

	/* retrieve sessionRecord */
	localSEID := sdreq.SEID()

	session, ok := pConn.store.GetSession(localSEID)
	if !ok {
		return sendError(ErrNotFoundWithParam("PFCP session", "localSEID", localSEID))
	}

	cause := upf.SendMsgToUPF(UpfMsgTypeDel, session, session.PacketForwardingRules)
	if cause == ie.CauseRequestRejected {
		return sendError(ErrWriteToDatapath)
	}

	if err := releaseAllocatedIPs(upf.ippool, &session); err != nil {
		return sendError(ErrOperationFailedWithReason("session IP dealloc", err.Error()))
	}

	// Release all TEIDs on the session
	releaseAllocatedTEIDs(upf.teidAllocator, &session)

	/* delete sessionRecord */
	pConn.RemoveSession(session)

	log.Debugw("Sending session deletion response:",
		zap.Uint64("Local SEID:", localSEID),
		zap.Uint64("Remote SEID:", session.remoteSEID))

	// Build response message
	smres := message.NewSessionDeletionResponse(0, /* MO?? <-- what's this */
		0,                                    /* FO <-- what's this? */
		session.remoteSEID,                   /* seid */
		sdreq.SequenceNumber,                 /* seq # */
		0,                                    /* priority */
		ie.NewCause(ie.CauseRequestAccepted), /* accept it blindly for the time being */
	)

	return smres, nil
}

func (pConn *PFCPConn) handleDigestReport(fseid uint64) {
	session, ok := pConn.store.GetSession(fseid)
	if !ok {
		log.Warn("No session found for fseid : ", fseid)
		return
	}

	seq := pConn.getSeqNum()
	srreq := message.NewSessionReportRequest(0, /* MO?? <-- what's this */
		0,                            /* FO <-- what's this? */
		0,                            /* seid */
		seq,                          /* seq # */
		0,                            /* priority */
		ie.NewReportType(0, 0, 0, 1), /*upir, erir, usar, dldr int*/
	)
	srreq.Header.SEID = session.remoteSEID

	var pdrID uint32

	var farID uint32

	for _, pdr := range session.Pdrs {
		if pdr.SrcIface == core {
			pdrID = pdr.PdrID

			farID = pdr.FarID

			break
		}
	}

	for _, far := range session.Fars {
		if far.FarID == farID {
			if far.ApplyAction&ActionNotify == 0 {
				log.Error("packet received for forwarding far. discard")
				return
			}
		}
	}

	if pdrID == 0 {
		log.Error("No Pdr found for downlink")

		return
	}

	srreq.DownlinkDataReport = ie.NewDownlinkDataReport(
		ie.NewPDRID(uint16(pdrID)))

	log.Debugw(
		"Sending Downlink Data Report",
		zap.Uint64("F-SEID", fseid),
		zap.Uint32("PDR ID", pdrID),
	)

	pConn.SendPFCPMsg(srreq)
}

func (pConn *PFCPConn) handleSessionReportResponse(msg message.Message) error {
	upf := pConn.upf

	srres, ok := msg.(*message.SessionReportResponse)
	if !ok {
		return errUnmarshal(errMsgUnexpectedType)
	}

	cause := srres.Cause.Payload[0]
	if cause == ie.CauseRequestAccepted {
		return nil
	}

	log.Warn("session req not accepted seq : ", srres.SequenceNumber)

	seid := srres.SEID()

	if cause == ie.CauseSessionContextNotFound {
		sessItem, ok := pConn.store.GetSession(seid)
		if !ok {
			return errProcess(ErrNotFoundWithParam("PFCP session context", "SEID", seid))
		}

		log.Warnf("Session context not found, deleting session locally with ID: %s", seid)

		pConn.RemoveSession(sessItem)

		cause := upf.SendMsgToUPF(
			UpfMsgTypeDel, sessItem, PacketForwardingRules{})
		if cause == ie.CauseRequestRejected {
			return errProcess(
				ErrOperationFailedWithParam("delete session from datapath", "seid", seid))
		}

		return nil
	}

	return nil
}
