package rtc

import (
	"sync"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/webrtc/v4"
)

type ManagedTransceiver struct {
	t    *webrtc.RTPTransceiver
	kind webrtc.RTPCodecType
	mu   sync.Mutex
	busy bool
}

type SessionTrackMetadata struct {
	trackId       string
	kind          string
	streamGroupId string
	clientId      string
}

type SessionTrack struct {
	Metadata       *SessionTrackMetadata
	Track          *webrtc.TrackRemote
	IsSubscribed   bool
	OwnerSessionId string
}

type SessionTracks struct {
	tracks []SessionTrack
	mu     sync.RWMutex
}

type Session interface {
	GetPeerConnection() *webrtc.PeerConnection
	GetClientId() string
	Close()
	SetRemoteSet(bool)
	GetRemoteTracks() []SessionTrack
	GetRemoteTrack(trackId string) (SessionTrack, bool)

	AddRemoteTrackMeta(trackID string, metadata SessionTrackMetadata)
	AddRemoteTrackStream(trackID string, track *webrtc.TrackRemote)
	RemoveRemoteTrack(trackId string)

	HandleStreamForwarding(trackID string, clientID string, shouldRenegotiate bool)

	SetSubscribedTrack(trackId string, subscribed bool)
	SetOwnerSessionIdForTrack(trackId string, sessionId string)

	Init(pc *webrtc.PeerConnection)

	RemoveRemoteTrackFromOwner(clientId string)

	IsInitialized() bool

	SetEmitFunc(fn func(event string, data ...any))
	Renegotiate(attempt *int) error

	GetOfferWaitChan() chan bool
}

type session struct {
	pc *webrtc.PeerConnection
	// transceivers         []*ManagedTransceiver
	clientId             string
	mu                   sync.Mutex
	remoteSet            bool
	remoteTracks         map[string]SessionTrack
	initialized          bool
	closed               bool
	emitFn               func(event string, data ...any)
	offerWaitChan        chan bool
	failedRenegotiations []func(attempt int)
}

func NewSession(clientID string) Session {
	return &session{
		pc: nil,
		// transceivers:         nil,
		clientId:             clientID,
		mu:                   sync.Mutex{},
		remoteSet:            false,
		remoteTracks:         make(map[string]SessionTrack),
		initialized:          false,
		closed:               false,
		offerWaitChan:        make(chan bool, 1),
		failedRenegotiations: make([]func(attempt int), 0),
	}
}

func (s *session) GetOfferWaitChan() chan bool {
	return s.offerWaitChan
}

func (s *session) IsInitialized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.initialized
}

func (s *session) SetEmitFunc(fn func(event string, data ...any)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emitFn = fn
}

func (s *session) Renegotiate(attempt *int) error {
	// Ensure only one renegotiation happens at a time
	s.offerWaitChan <- true

	s.mu.Lock()
	pc := s.pc
	emitFn := s.emitFn
	s.mu.Unlock()

	if pc == nil || emitFn == nil {
		return nil
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err := pc.SetLocalDescription(offer); err != nil {
		return err
	}

	emitFn("new_offer", s.clientId, offer)
	return nil
}

func (s *session) RemoveRemoteTrackFromOwner(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for trackId, track := range s.remoteTracks {
		if track.OwnerSessionId == clientId {
			delete(s.remoteTracks, trackId)
		}
	}
}

func (s *session) Init(pc *webrtc.PeerConnection) {
	s.pc = pc
	s.initialized = true
}

func (s *session) SetOwnerSessionIdForTrack(trackId string, sessionId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if track, exists := s.remoteTracks[trackId]; exists {
		track.OwnerSessionId = sessionId
		s.remoteTracks[trackId] = track
	}
}

func (s *session) SetSubscribedTrack(trackId string, subscribed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if track, exists := s.remoteTracks[trackId]; exists {
		track.IsSubscribed = subscribed
		s.remoteTracks[trackId] = track
	}
}

func (s *session) HandleStreamForwarding(trackID string, clientID string, shouldRenegotiate bool) {
	s.mu.Lock()
	track, exists := s.remoteTracks[trackID]
	s.mu.Unlock()

	if !exists || track.Track == nil {
		return
	}

	if track.Track == nil || track.Metadata == nil {
		logger.Sugar.Warnf("Skipping forwarding for track %s in session %s because track or metadata is nil", trackID, s.clientId)
		return
	}

	s.SetOwnerSessionIdForTrack(track.Track.ID(), clientID)
	if track.IsSubscribed {
		logger.Sugar.Infof("Not forwarding track %s (kind=%s) for session %s because it's subscribed", trackID, track.Metadata.kind, s.clientId)
		return
	}

	transceiver, err := forwardTrack(track.Track, s)

	if err != nil {
		logger.Sugar.Errorf("Error forwarding track %s (kind=%s) for session %s: %v", trackID, track.Metadata.kind, s.clientId, err)
		return
	}

	s.SetSubscribedTrack(trackID, true)

	if shouldRenegotiate {
		if s.Renegotiate(nil) != nil {
			logger.Sugar.Errorf("Error renegotiating after forwarding track %s (kind=%s) for session %s: %v", trackID, track.Metadata.kind, s.clientId, err)
			return
		}
	}

	if transceiver.Mid() != "" {
		s.emitFn("new_track", clientID, map[string]interface{}{
			"clientId":       clientID,
			"trackId":        trackID,
			"kind":           track.Metadata.kind,
			"streamGroupId":  track.Metadata.streamGroupId,
			"transceiverMid": transceiver.Mid(),
		})
	}
}

func (s *session) GetRemoteTrack(trackId string) (SessionTrack, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	track, exists := s.remoteTracks[trackId]
	return track, exists
}

func (s *session) GetRemoteTracks() []SessionTrack {
	s.mu.Lock()
	defer s.mu.Unlock()
	tracks := make([]SessionTrack, 0, len(s.remoteTracks))
	for _, track := range s.remoteTracks {
		tracks = append(tracks, track)
	}
	return tracks
}

func (s *session) AddRemoteTrackMeta(trackID string, metadata SessionTrackMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if track, exists := s.remoteTracks[trackID]; exists {
		track.Metadata = &metadata
		s.remoteTracks[trackID] = track
		return
	}

	s.remoteTracks[trackID] = SessionTrack{
		Metadata: &metadata,
		Track:    nil,
	}
}

func (s *session) AddRemoteTrackStream(trackID string, track *webrtc.TrackRemote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, exists := s.remoteTracks[trackID]; exists {
		existing.Track = track
		s.remoteTracks[trackID] = existing
		return
	}

	s.remoteTracks[trackID] = SessionTrack{
		Metadata: nil,
		Track:    track,
	}
}

func (s *session) RemoveRemoteTrack(trackId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.remoteTracks, trackId)
}

func (s *session) GetPeerConnection() *webrtc.PeerConnection {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pc
}

func (s *session) SetRemoteSet(remoteSet bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remoteSet = remoteSet
}

func (s *session) GetClientId() string {
	return s.clientId
}

func (s *session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pc != nil {
		s.pc.Close()
		s.pc = nil
	}
	s.remoteTracks = make(map[string]SessionTrack)
	s.closed = true
}

type SessionManager interface {
	GetSession(clientId string) (Session, bool)
	AddSession(session Session)
	RemoveSession(clientId string)
	CloseAll() error

	AddRemoteTrackMeta(trackID string, metadata SessionTrackMetadata)
	AddRemoteTrackStream(trackID string, track *webrtc.TrackRemote)
	RemoveSubscribedTrack(trackId string)
	GetSubscribedTracks() []SessionTrack

	GetSessions(id string) []Session

	SetOwnerSessionIdForTrack(trackId string, sessionId string)
	RemoveTracksForSession(sessionId string)
}

type sessionManager struct {
	id               string
	sessions         map[string]Session
	subscribedTracks map[string]SessionTrack
	mu               sync.RWMutex
}

func NewSessionManager(id string) SessionManager {
	return &sessionManager{
		id:               id,
		sessions:         make(map[string]Session),
		subscribedTracks: make(map[string]SessionTrack),
		mu:               sync.RWMutex{},
	}
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

func (sm *sessionManager) GetSessions(id string) []Session {
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

func (sm *sessionManager) AddSession(session Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.GetClientId()] = session
}

func (sm *sessionManager) RemoveSession(clientId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, clientId)
}

func (sm *sessionManager) CloseAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, session := range sm.sessions {
		session.Close()
	}
	sm.sessions = make(map[string]Session)
	return nil
}
