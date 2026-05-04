package rtc

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

type SessionManager interface {
	GetGroupId() string

	AddSession(session Session, wsID string)
	GetSession(clientId string) (Session, bool)
	GetSessions() []Session

	RemoveSession(clientId string)

	CloseAll() error

	AddRemoteTrackMeta(trackID string, metadata SessionTrackMetadata)
	AddRemoteTrackStream(trackID string, track *webrtc.TrackRemote)
	RemoveSubscribedTrack(trackId string)

	GetSubscribedTracks() []SessionTrack

	SetOwnerSessionIdForTrack(trackId string, sessionId string)
	RemoveTracksForSession(sessionId string)

	GetClientIDFromWsID(wsID string) (string, bool)
	GetSessionByWsID(wsID string) (Session, bool)
}

type sessionManager struct {
	id               string
	sessions         map[string]Session
	subscribedTracks map[string]SessionTrack
	wsToClientId     map[string]string
	mu               sync.RWMutex
}

func NewSessionManager(id string) SessionManager {
	return &sessionManager{
		id:               id,
		sessions:         make(map[string]Session),
		subscribedTracks: make(map[string]SessionTrack),
		wsToClientId:     make(map[string]string),
		mu:               sync.RWMutex{},
	}
}

func (sm *sessionManager) GetGroupId() string {
	return sm.id
}

func (sm *sessionManager) RemoveTracksForSession(sessionId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for trackId, track := range sm.subscribedTracks {
		if track.OwnerSessionId == sessionId {
			delete(sm.subscribedTracks, trackId)
		}
	}
}

func (sm *sessionManager) SetOwnerSessionIdForTrack(trackId string, sessionId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if track, exists := sm.subscribedTracks[trackId]; exists {
		track.OwnerSessionId = sessionId
		sm.subscribedTracks[trackId] = track
	}
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

func (sm *sessionManager) AddRemoteTrackMeta(trackID string, metadata SessionTrackMetadata) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.subscribedTracks[trackID] = SessionTrack{
		Metadata: &metadata,
		Track:    nil,
	}
}

func (sm *sessionManager) AddRemoteTrackStream(trackID string, track *webrtc.TrackRemote) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if existing, exists := sm.subscribedTracks[trackID]; exists {
		existing.Track = track
		sm.subscribedTracks[trackID] = existing
		return
	}
	sm.subscribedTracks[trackID] = SessionTrack{
		Metadata: nil,
		Track:    track,
	}
}

func (sm *sessionManager) RemoveSubscribedTrack(trackId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.subscribedTracks, trackId)
}

func (sm *sessionManager) GetSubscribedTracks() []SessionTrack {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	tracks := make([]SessionTrack, 0, len(sm.subscribedTracks))
	for _, track := range sm.subscribedTracks {
		tracks = append(tracks, track)
	}
	return tracks
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

func (sm *sessionManager) RemoveSession(clientId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, clientId)

	for wsID, cid := range sm.wsToClientId {
		if cid == clientId {

			delete(sm.wsToClientId, wsID)
			break
		}
	}
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
