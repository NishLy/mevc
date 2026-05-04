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

	StartRTPStream(trackID string, clientID string) Session

	SetSubscribedTrack(trackId string, subscribed bool)
	SetOwnerSessionIdForTrack(trackId string, sessionId string)

	RemoveRemoteTrackFromOwner(clientId string)

	IsInitialized() bool

	SetEmitFunc(fn func(event string, data ...any))
	Renegotiate(attempt *int) error

	GetOfferWaitChan() chan bool
	AddSelfTrackMetadata(trackID string, metadata SessionTrackMetadata)
	RemoveSelfTrackMetadata(trackID string)

	GetSelfTracksMetadata(streamGroupId string) (SessionTrackMetadata, bool)
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
	initialized        bool
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
		initialized:        true,
		closed:             false,
		offerWaitChan:      make(chan bool, 1),
		selfTracksMetadata: make(map[string]SessionTrackMetadata),
	}
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

func (s *session) IsInitialized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.initialized
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

func (s *session) StartRTPStream(trackID string, clientID string) Session {
	if s.clientId == clientID {
		return nil
	}

	s.mu.Lock()
	track, exists := s.remoteTracks[trackID]
	s.mu.Unlock()

	if !exists || track.Track == nil || track.Metadata == nil || track.IsSubscribed {
		return nil
	}

	s.SetOwnerSessionIdForTrack(track.Track.ID(), clientID)
	err := forwardTrack(s, track.Track)

	if err != nil {
		logger.Sugar.Errorf("Error forwarding track %s (kind=%s) for session %s: %v", trackID, track.Metadata.kind, s.clientId, err)
		return nil
	}

	s.SetSubscribedTrack(trackID, true)

	return s
}

func (s *session) Renegotiate(attempt *int) error {
	// defer func() {
	// 	time.Sleep(5 * time.Second)

	// 	logger.Sugar.Infof("Releasing offer wait lock for session %s after renegotiation attempt", s.clientId)

	// 	<-s.offerWaitChan
	// }()

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
	return nil
}

func (s *session) RemoveRemoteTrack(trackId string) {
	s.mu.Lock()
	track, exists := s.remoteTracks[trackId]
	s.mu.Unlock()

	if exists && track.Track != nil {
		if s.pc != nil {
			mid := RemoveTrackFromPeerConnection(s.pc, track.Track.ID())
			if mid != nil {
				s.emitFn("track_removed", *mid)
			}
		}
	}

	s.mu.Lock()
	delete(s.remoteTracks, trackId)
	s.mu.Unlock()
}

func (s *session) RemoveRemoteTrackFromOwner(clientId string) {
	s.mu.Lock()
	remoteTracks := s.remoteTracks
	s.mu.Unlock()

	for trackId, track := range remoteTracks {
		if track.OwnerSessionId == clientId {
			s.mu.Lock()
			delete(s.remoteTracks, trackId)
			s.mu.Unlock()

			if track.Track != nil {
				if s.pc != nil {
					mid := RemoveTrackFromPeerConnection(s.pc, track.Track.ID())

					if mid != nil {
						s.emitFn("track_removed", *mid)
					}
				}
			}
		}
	}
}

func RemoveTrackFromPeerConnection(pc *webrtc.PeerConnection, trackID string) *string {
	for _, transceiver := range pc.GetTransceivers() {
		if transceiver.Sender() == nil || transceiver.Sender().Track() == nil {
			continue
		}

		if transceiver.Sender().Track().ID() == trackID {
			pc.RemoveTrack(transceiver.Sender())
			mid := transceiver.Mid()
			return &mid
		}
	}
	return nil
}
