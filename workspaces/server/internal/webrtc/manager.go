package rtc

import (
	"sync"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
)

type SessionManager interface {
	GetGroupId() string

	AddSession(session Session, wsID string) bool
	GetSession(clientId string) (Session, bool)
	GetSessions() []Session

	CloseAll() error

	GetClientIDFromWsID(wsID string) (string, bool)
	GetSessionByWsID(wsID string) (Session, bool)

	AddRouter(streamID string, router *TrackRouter)
	GetRouter(streamID string) (*TrackRouter, bool)
	RemoveRouter(streamID string)

	MoveToRoom(session Session)
	KickFromLoby(session Session)
	GetLobbySessions() []LobySession

	// AddTemporaryMetadata(streamID string, metadata SessionTrackMetadata)
	// GetTemporaryMetadata(streamID string) (SessionTrackMetadata, bool)
	// RemoveTemporaryMetadata(streamID string)

	SubscribeToExistingTracks(newSession Session)

	RemoveFromSessionManager(clientID string)
}

type LobySession struct {
	wsID    string
	session Session
}

type sessionManager struct {
	id           string
	sessions     map[string]Session
	wsToClientId map[string]string
	routers      map[string]*TrackRouter
	autoClose    bool
	autoAccept   bool
	lobby        map[string]LobySession
	mu           sync.RWMutex
}

func NewSessionManager(id string, autoAccept bool) SessionManager {
	return &sessionManager{
		id:           id,
		sessions:     make(map[string]Session),
		wsToClientId: make(map[string]string),
		routers:      make(map[string]*TrackRouter),
		autoClose:    true,
		autoAccept:   autoAccept,
		lobby:        make(map[string]LobySession),
		mu:           sync.RWMutex{},
	}
}

func (sm *sessionManager) GetLobbySessions() []LobySession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	lobbySessions := make([]LobySession, 0, len(sm.lobby))

	for _, lobbySession := range sm.lobby {
		lobbySessions = append(lobbySessions, lobbySession)
	}
	return lobbySessions
}

func (sm *sessionManager) MoveToRoom(session Session) {
	sm.mu.Lock()
	lobbySession, exist := sm.lobby[session.GetClientId()]
	if !exist {
		logger.Sugar.Warnf("No lobby session found for client %s, rejecting session", session.GetClientId())
	}

	delete(sm.lobby, session.GetClientId())
	sm.mu.Unlock()
	sm.AddSession(lobbySession.session, lobbySession.wsID)
}

func (sm *sessionManager) KickFromLoby(session Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, exist := sm.lobby[session.GetClientId()]
	if !exist {
		return
	}
	delete(sm.lobby, session.GetClientId())
	session.Close()
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

func (sm *sessionManager) AddSession(session Session, wsID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.wsToClientId[wsID] = session.GetClientId()
	if !sm.autoAccept {
		sm.lobby[session.GetClientId()] = LobySession{
			wsID:    wsID,
			session: session,
		}
		logger.Sugar.Infof("Added session for client %s to lobby", session.GetClientId())
		return false
	}

	sm.sessions[session.GetClientId()] = session
	return true
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

	lobbySession, existInLobby := sm.lobby[clientID]
	_, existInSessions := sm.sessions[clientID]

	if existInLobby {
		return lobbySession.session, true
	}
	if !existInSessions {
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
				newSession.Emit("new_track", router.incomingTrack.StreamID(), &SessionTrackMetadata{
					TrackId:       router.incomingTrack.ID(),
					StreamId:      router.incomingTrack.StreamID(),
					Kind:          router.incomingTrack.Kind().String(),
					ClientId:      router.publisherID,
					StreamGroupId: router.metadata.StreamGroupId,
					Label:         router.metadata.Label,
					Enabled:       router.metadata.Enabled,
					Username:      router.metadata.Username,
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

	if router, exists := sm.routers[trackId]; exists {
		router.Close()
	}

	delete(sm.routers, trackId)
}

func (sm *sessionManager) RemoveFromSessionManager(clientID string) {
	session, exists := sm.GetSession(clientID)
	if !exists || session == nil {
		return
	}

	var routersToClose []string

	sm.mu.RLock()
	for id, router := range sm.routers {
		if router.publisherID == session.GetClientId() {
			routersToClose = append(routersToClose, id)
		}
		// Remove the viewer from all routers
		if err := router.RemoveViewer(clientID); err != nil {
			logger.Sugar.Warnf("Failed to remove viewer %s from router for stream %s: %v", clientID, router.incomingTrack.StreamID(), err)
		}
	}
	sm.mu.RUnlock()

	// Now close all routers where this client was the publisher
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
	session.Close()

	if len(sm.GetSessions()) == 0 {
		delete(GlobalSessionManager, sm.GetGroupId())
	}

	logger.Sugar.Infof("Removed session for client %s", clientID)
}
