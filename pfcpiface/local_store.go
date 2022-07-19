// SPDX-License-Identifier: Apache-2.0
// Copyright 2022-present Open Networking Foundation

package pfcpiface

import (
	"sync"

	"go.uber.org/zap"
)

type InMemoryStore struct {
	// sessions stores all PFCP sessions.
	// sync.Map is optimized for case when multiple goroutines
	// read, write, and overwrite entries for disjoint sets of keys.
	sessions sync.Map
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (i *InMemoryStore) GetAllSessions() []PFCPSession {
	sessions := make([]PFCPSession, 0)

	i.sessions.Range(func(key, value interface{}) bool {
		v := value.(PFCPSession)
		sessions = append(sessions, v)
		return true
	})

	log.Debugw(
		"Got all PFCP sessions from local store",
		zap.Any("sessions", sessions),
	)

	return sessions
}

func (i *InMemoryStore) PutSession(session PFCPSession) error {
	if session.localSEID == 0 {
		return ErrInvalidArgument("session.localSEID", session.localSEID)
	}

	i.sessions.Store(session.localSEID, session)

	log.Debugw(
		"Saved PFCP sessions to local store",
		zap.Any("session", session),
	)

	return nil
}

func (i *InMemoryStore) DeleteSession(fseid uint64) error {
	i.sessions.Delete(fseid)

	log.Debugw(
		"PFCP session removed from local store",
		zap.Uint64("F-SEID", fseid),
	)

	return nil
}

func (i *InMemoryStore) DeleteAllSessions() bool {
	i.sessions.Range(func(key, value interface{}) bool {
		i.sessions.Delete(key)
		return true
	})

	log.Debug("All PFCP sessions removed from local store")

	return true
}

func (i *InMemoryStore) GetSession(fseid uint64) (PFCPSession, bool) {
	sess, ok := i.sessions.Load(fseid)
	if !ok {
		return PFCPSession{}, false
	}

	session, ok := sess.(PFCPSession)
	if !ok {
		return PFCPSession{}, false
	}

	log.Debugw(
		"Got PFCP session from local store",
		zap.Any("session", session),
	)

	return session, ok
}
