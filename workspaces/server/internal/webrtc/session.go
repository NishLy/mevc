package rtc

import (
	"sync"

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
	streamId      string
	label         string
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
	GetMutex() *sync.Mutex

	SetEmitFunc(fn func(event string, data ...any))
	Renegotiate(attempt *int) error

	GetOfferWaitChan() chan bool

	Emit(event string, data ...any)

	IsRemoteSet() bool
}

type WaitTrackResult struct {
	Track *webrtc.TrackRemote
	Err   error
}

type session struct {
	pc *webrtc.PeerConnection
	// transceivers         []*ManagedTransceiver
	clientId           string
	mu                 sync.Mutex
	remoteSet          bool
	remoteTracks       map[string]SessionTrack
	closed             bool
	emitFn             func(event string, data ...any)
	offerWaitChan      chan bool
	selfTracksMetadata map[string]SessionTrackMetadata
}

func NewSession(clientID string, pc *webrtc.PeerConnection) Session {

	return &session{
		pc:                 pc,
		clientId:           clientID,
		mu:                 sync.Mutex{},
		remoteSet:          false,
		remoteTracks:       make(map[string]SessionTrack),
		closed:             false,
		offerWaitChan:      make(chan bool, 1),
		selfTracksMetadata: make(map[string]SessionTrackMetadata),
	}
}

func (s *session) IsRemoteSet() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.remoteSet
}

func (s *session) GetMutex() *sync.Mutex {
	return &s.mu
}

func (s *session) AddSelfTrackMetadata(trackID string, metadata SessionTrackMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selfTracksMetadata[trackID] = metadata
}

func (s *session) GetSelfTracksMetadata(streamGroupId string) (SessionTrackMetadata, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, metadata := range s.selfTracksMetadata {
		if metadata.streamGroupId == streamGroupId {
			return metadata, true
		}
	}
	return SessionTrackMetadata{}, false
}

func (s *session) RemoveSelfTrackMetadata(trackID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.selfTracksMetadata, trackID)
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

func (s *session) GetClientId() string {
	return s.clientId
}

func (s *session) GetOfferWaitChan() chan bool {
	return s.offerWaitChan
}

func (s *session) GetPeerConnection() *webrtc.PeerConnection {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pc
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

func (s *session) SetRemoteSet(remoteSet bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remoteSet = remoteSet
}

func (s *session) SetEmitFunc(fn func(event string, data ...any)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emitFn = fn
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

func (s *session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed || s == nil {
		return
	}

	if s.pc != nil {
		s.pc.Close()
		s.pc = nil
	}

	s.remoteTracks = make(map[string]SessionTrack)
	s.closed = true
}

func (s *session) Renegotiate(attempt *int) (err error) {
	s.offerWaitChan <- true

	// Always release
	defer func() {
		// <-s.offerWaitChan
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	pc := s.pc
	emitFn := s.emitFn

	if pc == nil || emitFn == nil {
		return nil
	}

	var offer webrtc.SessionDescription

	offer, err = pc.CreateOffer(nil)
	if err != nil {
		return
	}

	if err = pc.SetLocalDescription(offer); err != nil {
		return
	}

	emitFn("new_offer", s.clientId, offer)
	return nil
}

func (s *session) Emit(event string, data ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.emitFn != nil {
		s.emitFn(event, data...)
	}
}
