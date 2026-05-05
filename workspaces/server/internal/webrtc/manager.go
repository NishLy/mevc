package rtc

import (
	"sync"
)

type SessionManager interface {
	GetGroupId() string

	AddSession(session Session, wsID string)
	GetSession(clientId string) (Session, bool)
	GetSessions() []Session

	CloseAll() error

	GetClientIDFromWsID(wsID string) (string, bool)
	GetSessionByWsID(wsID string) (Session, bool)

	AddRouter(streamID string, router *TrackRouter)
	GetRouter(streamID string) (*TrackRouter, bool)
	RemoveRouter(streamID string)

	// AddTemporaryMetadata(streamID string, metadata SessionTrackMetadata)
	// GetTemporaryMetadata(streamID string) (SessionTrackMetadata, bool)
	// RemoveTemporaryMetadata(streamID string)

	SubscribeToExistingTracks(newSession Session)

	RemoveFromSessionManager(clientID string)
}

type sessionManager struct {
	id           string
	sessions     map[string]Session
	wsToClientId map[string]string
	routers      map[string]*TrackRouter
	// temporaryMetadata map[string]SessionTrackMetadata
	mu sync.RWMutex
}

func NewSessionManager(id string) SessionManager {
	return &sessionManager{
		id:           id,
		sessions:     make(map[string]Session),
		wsToClientId: make(map[string]string),
		routers:      make(map[string]*TrackRouter),
		// temporaryMetadata: make(map[string]SessionTrackMetadata),
		mu: sync.RWMutex{},
	}
}

func (sm *sessionManager) GetGroupId() string {
	return sm.id
}

func (sm *sessionManager) GetSessions() []Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *sessionManager) GetSession(clientId string) (Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[clientId]
	return session, exists
}

func (sm *sessionManager) AddSession(session Session, wsID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.GetClientId()] = session
	sm.wsToClientId[wsID] = session.GetClientId()
}

func (sm *sessionManager) CloseAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, session := range sm.sessions {
		session.Close()
	}
	sm.sessions = make(map[string]Session)
	sm.wsToClientId = make(map[string]string)
	return nil
}

func (sm *sessionManager) GetClientIDFromWsID(wsID string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	clientID, exists := sm.wsToClientId[wsID]
	return clientID, exists
}

func (sm *sessionManager) GetSessionByWsID(wsID string) (Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	clientID, exists := sm.wsToClientId[wsID]
	if !exists {
		return nil, false
	}
	session, exists := sm.sessions[clientID]
	return session, exists
}

func (sm *sessionManager) SubscribeToExistingTracks(newSession Session) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	needsRenegotiation := false

	for _, router := range sm.routers {
		// Don't subscribe the user to their own published tracks
		if router.publisherPC == newSession.GetPeerConnection() {
			continue
		}

		err := router.AddViewer(newSession)
		if err == nil {
			needsRenegotiation = true

			if router.hasStarted && router.metadata != nil {
				newSession.Emit("new_track", router.incomingTrack.StreamID(), map[string]interface{}{
					"trackId":       router.incomingTrack.ID(),
					"streamId":      router.incomingTrack.StreamID(),
					"kind":          router.incomingTrack.Kind().String(),
					"clientId":      newSession.GetClientId(),
					"streamGroupId": sm.GetGroupId(),
					"label":         router.metadata.label,
				})
			}
		}
	}

	if needsRenegotiation {
		newSession.Renegotiate(nil)
	}
}

func (sm *sessionManager) AddRouter(trackId string, router *TrackRouter) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.routers == nil {
		sm.routers = make(map[string]*TrackRouter)
	}
	sm.routers[trackId] = router
}

func (sm *sessionManager) GetRouter(trackId string) (*TrackRouter, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	router, exists := sm.routers[trackId]
	return router, exists
}

func (sm *sessionManager) RemoveRouter(trackId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.routers, trackId)
}

func (sm *sessionManager) RemoveFromSessionManager(clientID string) {
	session, exists := sm.GetSession(clientID)
	if !exists || session == nil {
		return
	}

	session.Close()

	var routersToClose []string
	for _, router := range sm.routers {
		if router.publisherID == session.GetClientId() {
			router.Close()
			routersToClose = append(routersToClose, router.incomingTrack.StreamID())
		}
	}

	for _, trackId := range routersToClose {
		sm.RemoveRouter(trackId)
	}

	for wsID, cid := range sm.wsToClientId {
		if cid == clientID {
			delete(sm.wsToClientId, wsID)
			break
		}
	}

	delete(sm.sessions, clientID)

	if len(sm.GetSessions()) == 0 {
		delete(GlobalSessionManager, sm.GetGroupId())
	}
}
