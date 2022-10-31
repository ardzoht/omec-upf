package pfcpiface

// CreateURR appends urr to existing list of URRs in the session.
func (s *PFCPSession) CreateURR(u Urr) {
	s.Urrs = append(s.Urrs, u)
}

func (s *PFCPSession) UpdateURR(u Urr) error {
	for idx, v := range s.Urrs {
		if v.UrrID == u.UrrID {
			s.Urrs[idx] = u
			return nil
		}
	}

	return ErrNotFound("URR")
}

func (s *PFCPSession) RemoveURR(id uint32) (*Urr, error) {
	for idx, v := range s.Urrs {
		if v.UrrID == id {
			s.Urrs = append(s.Urrs[:idx], s.Urrs[idx+1:]...)
			return &v, nil
		}
	}

	return nil, ErrNotFound("URR")
}
