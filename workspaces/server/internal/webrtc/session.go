package rtc

import (
	"sync"
	"time"

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

	TryStream(trackID string, clientID string)

	SetSubscribedTrack(trackId string, subscribed bool)
	SetOwnerSessionIdForTrack(trackId string, sessionId string)

	RemoveRemoteTrackFromOwner(clientId string)

	IsInitialized() bool

	SetEmitFunc(fn func(event string, data ...any))
	Renegotiate(attempt *int) error

	GetOfferWaitChan() chan bool

	GetQueueTrackForwarding() chan func()

	RunWorker()
}

type WaitTrackResult struct {
	Track *webrtc.TrackRemote
	Err   error
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
	queueTrackForwarding chan func()
	failedRenegotiations []func(attempt int)
}

func NewSession(clientID string, pc *webrtc.PeerConnection) Session {

	return &session{
		pc:                   pc,
		clientId:             clientID,
		mu:                   sync.Mutex{},
		remoteSet:            false,
		remoteTracks:         make(map[string]SessionTrack),
		initialized:          true,
		closed:               false,
		offerWaitChan:        make(chan bool, 1),
		queueTrackForwarding: make(chan func(), 100),
		failedRenegotiations: make([]func(attempt int), 0),
	}
}

func (s *session) GetQueueTrackForwarding() chan func() {
	return s.queueTrackForwarding
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
	s.offerWaitChan <- true

	pc := s.pc
	emitFn := s.emitFn

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

	defer func() {
		time.Sleep(5 * time.Second)

		<-s.offerWaitChan
	}()

	return nil
}

func (s *session) removeTrackFromPeerConnection(trackID string) {
	if s.pc == nil {
		return
	}

	for _, transceiver := range s.pc.GetTransceivers() {
		if transceiver.Sender() == nil || transceiver.Sender().Track() == nil {
			continue
		}

		if transceiver.Sender().Track().ID() == trackID {
			s.pc.RemoveTrack(transceiver.Sender())

			s.emitFn("track_removed", transceiver.Mid())
		}
	}

	if err := s.Renegotiate(nil); err != nil {
		logger.Sugar.Errorf("Error renegotiating after removing track %s for session %s: %v", trackID, s.clientId, err)
	}
}

func (s *session) RemoveRemoteTrackFromOwner(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for trackId, track := range s.remoteTracks {
		if track.OwnerSessionId == clientId {
			delete(s.remoteTracks, trackId)
			if track.Track != nil {
				s.removeTrackFromPeerConnection(track.Track.ID())
			}
		}
	}
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

func (s *session) TryStream(trackID string, clientID string) {
	if s.clientId == clientID {
		return
	}

	s.mu.Lock()
	track, exists := s.remoteTracks[trackID]
	s.mu.Unlock()

	if !exists || track.Track == nil {
		return
	}

	s.SetOwnerSessionIdForTrack(track.Track.ID(), clientID)

	if track.Track == nil || track.Metadata == nil {
		logger.Sugar.Warnf("Skipping forwarding for track %s in session %s because track or metadata is nil", trackID, s.clientId)
		return
	}

	if track.IsSubscribed {
		return
	}

	transceiver, localTrack, err := createLocalTrancieverAndTrack(track.Track, s)

	if err != nil {
		logger.Sugar.Errorf("Error creating local track for forwarding track %s (kind=%s) for session %s: %v", trackID, track.Metadata.kind, s.clientId, err)
		return
	}

	err = forwardTrack(s.pc, transceiver, track.Track, localTrack, s.clientId)

	if err != nil {
		logger.Sugar.Errorf("Error forwarding track %s (kind=%s) for session %s: %v", trackID, track.Metadata.kind, s.clientId, err)
		return
	}

	if err := s.Renegotiate(nil); err != nil {
		logger.Sugar.Errorf("Error renegotiating after forwarding track %s (kind=%s) for session %s: %v", trackID, track.Metadata.kind, s.clientId, err)
		return
	}

	s.SetSubscribedTrack(trackID, true)
}

func (s *session) RunWorker() {
	go func() {
		// Use a recover block to prevent the worker from dying if one track fails
		defer func() {
			if r := recover(); r != nil {
				logger.Sugar.Errorf("Recovered from worker panic: %v", r)
			}
		}()

		for pending := range s.queueTrackForwarding {
			pending() // Execute the task
		}
	}()
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

	track, exists := s.remoteTracks[trackId]

	if exists && track.Track != nil {
		s.removeTrackFromPeerConnection(track.Track.ID())
	}

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
	if s.closed || s == nil {
		return
	}

	if s.pc != nil {
		s.pc.Close()
		s.pc = nil
	}

	s.remoteTracks = make(map[string]SessionTrack)
	s.queueTrackForwarding = make(chan func(), 100)
	s.closed = true
}

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
