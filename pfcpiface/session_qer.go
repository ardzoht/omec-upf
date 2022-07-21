// SPDX-License-Identifier: Apache-2.0
// Copyright 2020 Intel Corporation

package pfcpiface

type QosLevel uint8

const (
	ApplicationQos QosLevel = 0
	SessionQos     QosLevel = 1
)

// CreateQER appends qer to existing list of QERs in the session.
func (s *PFCPSession) CreateQER(q qer) {
	s.Qers = append(s.Qers, q)
}

// UpdateQER updates existing qer in the session.
func (s *PFCPSession) UpdateQER(q qer) error {
	for idx, v := range s.Qers {
		if v.qerID == q.qerID {
			s.Qers[idx] = q
			return nil
		}
	}

	return ErrNotFound("QER")
}

// Int version of code present at https://github.com/juliangruber/go-intersect
func Intersect(a []uint32, b []uint32) []uint32 {
	set := make([]uint32, 0)

	for i := 0; i < len(a); i++ {
		if contains(b, a[i]) {
			set = append(set, a[i])
		}
	}

	return set
}

func contains(a []uint32, val uint32) bool {
	for i := 0; i < len(a); i++ {
		if val == a[i] {
			return true
		}
	}

	return false
}

func findItemIndex(slice []uint32, val uint32) int {
	for i := 0; i < len(slice); i++ {
		if val == slice[i] {
			return i
		}
	}

	return len(slice)
}

// MarkSessionQer : identify and Mark session QER with flag.
func (s *PFCPSession) MarkSessionQer(qers []qer) {
	sessQerIDList := make([]uint32, 0)
	lastPdrIndex := len(s.Pdrs) - 1
	// create search list with first pdr's qerlist */
	sessQerIDList = append(sessQerIDList, s.Pdrs[lastPdrIndex].QerIDList...)

	// If PDRs have no QERs, then no marking for session qers is needed.
	// If PDRS have one QER and all PDRs point to same QER, then consider it as application qer.
	// If number of QERS is 2 or more, then search for session QER
	if (len(sessQerIDList) < 1) || (len(qers) < 2) {
		log.Info("need atleast 1 QER in PDR or 2 QERs in session to mark session QER.")
		return
	}

	// loop around all pdrs and find matching qers.
	for i := range s.Pdrs {
		// match every qer in searchlist in pdr's qer list
		sList := Intersect(sessQerIDList, s.Pdrs[i].QerIDList)
		if len(sList) == 0 {
			return
		}

		copy(sessQerIDList, sList)
	}

	// Loop through qer list and mark qer which matches
	//	  with entry in searchlist as sessionQos
	//    if len(sessQerIDList) = 1 : use as matching session QER
	//    if len(sessQerIDList) = 2 : loop and search for qer with
	//                                bigger MBR and choose as session QER
	//    if len(sessQerIDList) = 0 : no session QER
	//    if len(sessQerIDList) = 3 : TBD (UE level QER handling).
	//                                Currently handle same as len = 2
	var (
		sessionIdx int
		sessionMbr uint64
		sessQerID  uint32
	)

	if len(sessQerIDList) > 3 {
		log.Warn("Qer ID list size above 3. Not supported.")
	}

	for idx, qer := range qers {
		if contains(sessQerIDList, qer.qerID) {
			if qer.ulGbr > 0 || qer.dlGbr > 0 {
				log.Warn("Do not consider qer with non zero gbr value for session qer")
				continue
			}

			if qer.ulMbr >= sessionMbr {
				sessionIdx = idx
				sessQerID = qer.qerID
				sessionMbr = qer.ulMbr
			}
		}
	}

	log.Warn("session QER found. QER ID : ", sessQerID)

	qers[sessionIdx].qosLevel = SessionQos

	for i := range s.Pdrs {
		// remove common qerID from pdr's qer list
		idx := findItemIndex(s.Pdrs[i].QerIDList, sessQerID)
		if idx != len(s.Pdrs[i].QerIDList) {
			s.Pdrs[i].QerIDList = append(s.Pdrs[i].QerIDList[:idx], s.Pdrs[i].QerIDList[idx+1:]...)
			s.Pdrs[i].QerIDList = append(s.Pdrs[i].QerIDList, sessQerID)
		}
	}
}

// RemoveQER removes qer from existing list of QERs in the session.
func (s *PFCPSession) RemoveQER(id uint32) (*qer, error) {
	for idx, v := range s.Qers {
		if v.qerID == id {
			s.Qers = append(s.Qers[:idx], s.Qers[idx+1:]...)
			return &v, nil
		}
	}

	return nil, ErrNotFound("QER")
}
