package rtc

import (
	"sync"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

type ManagedTransceiver struct {
	t    *webrtc.RTPTransceiver
	kind webrtc.RTPCodecType
	mu   sync.Mutex
	busy bool
}

type Session struct {
	pc                *webrtc.PeerConnection
	transceivers      []*ManagedTransceiver // ← replaces raw GetTransceivers()
	subscribedTracks  map[string]*webrtc.TrackRemote
	pendingCandidates []webrtc.ICECandidateInit
	remoteSet         bool
	clientId          string
	mu                sync.Mutex
}

func isAlreadyAssigned(tracks map[string]*webrtc.TrackRemote, trackId string) bool {
	_, exists := tracks[trackId]
	return exists
}

type SessionManager struct {
	id       string
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager(id string) *SessionManager {
	return &SessionManager{
		id:       id,
		sessions: make(map[string]*Session),
		mu:       sync.RWMutex{},
	}
}

func (sm *SessionManager) GetSession(connID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[connID]
	return session, exists
}

func (sm *SessionManager) AddSession(connID string, session *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[connID] = session
}

func (sm *SessionManager) RemoveSession(connID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, connID)
}

type UserTrackMetadata struct {
	trackId       string
	kind          string
	streamGroupId string
	clientId      string
}

type UserTracksMetadata struct {
	tracks []UserTrackMetadata
}

var GlobalUserTracksMetadata = make(map[string]*UserTracksMetadata)

var MAXIMUM_TRANCEIVERS = 10

var (
	GlobalSessionManager = make(map[string]*SessionManager)
)

func WebRTCBootstrap(hub ws.WsHub) {
	// mustDialUDP()
	RegisterHandlers(hub)
}

func MustCreatePeerConnection() (*webrtc.PeerConnection, []*ManagedTransceiver) {
	me := &webrtc.MediaEngine{}
	MustRegisterCodecs(me)

	ir := &interceptor.Registry{}
	pli, err := intervalpli.NewReceiverInterceptor()
	Must(err)
	ir.Add(pli)
	Must(webrtc.RegisterDefaultInterceptors(me, ir))

	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithInterceptorRegistry(ir))
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	Must(err)

	var managed []*ManagedTransceiver
	for i := 0; i < MAXIMUM_TRANCEIVERS; i++ {
		for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
			t, err := pc.AddTransceiverFromKind(kind, webrtc.RTPTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendrecv,
			})
			Must(err)
			managed = append(managed, &ManagedTransceiver{t: t, kind: kind})
		}
	}

	return pc, managed
}

func MustRegisterCodecs(me *webrtc.MediaEngine) {
	codecs := []struct {
		mime      string
		clockRate uint32
		kind      webrtc.RTPCodecType
	}{
		{webrtc.MimeTypeVP8, 90000, webrtc.RTPCodecTypeVideo},
		{webrtc.MimeTypeOpus, 48000, webrtc.RTPCodecTypeAudio},
	}
	for _, c := range codecs {
		Must(me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: c.mime, ClockRate: c.clockRate},
		}, c.kind))
	}
}

func RegisterPeerCallbacks(hub ws.WsHub, pc *webrtc.PeerConnection, conn ws.WebSocketConnection) {
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {

		groupID := conn.GetGroupId()
		if groupID == nil {
			return
		}

		sessionManager, exists := GlobalSessionManager[*groupID]
		session, exists := sessionManager.GetSession(conn.ID())

		hub.EmitTo(*groupID, "new_track", map[string]interface{}{
			"clientId": session.clientId,
			"trackId":  track.ID(),
			"kind":     track.Kind().String(),
		})

		if exists {
			for _, s := range sessionManager.sessions {
				if s.pc != pc && !isAlreadyAssigned(s.subscribedTracks, track.ID()) {
					forwardTrack(track, receiver, s)
					s.subscribedTracks[track.ID()] = track
				}
			}
		}
	})

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		logger.Sugar.Infof("ICE Connection state: %s", state)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		logger.Sugar.Infof("Peer Connection state: %s", state)
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {

			groupID := conn.GetGroupId()
			if groupID == nil {
				return
			}

			SessionManager, exists := GlobalSessionManager[*groupID]
			if !exists {
				logger.Sugar.Warnf("Session manager for group %s does not exist", *groupID)
				return
			}

			Session, exists := SessionManager.GetSession(conn.ID())
			if !exists {
				logger.Sugar.Warnf("Session for connection %s does not exist in group %s", conn.ID(), *groupID)
				return
			}

			if exists {
				Session.pc.Close()
				SessionManager.RemoveSession(conn.ID())
			}

			if len(SessionManager.sessions) == 0 {
				delete(GlobalSessionManager, *groupID)
			}

			conn.Emit("peer_connection_closed", Session.clientId, "Peer connection closed due to failure or closure")
		}
	})

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		groupID := conn.GetGroupId()
		if groupID == nil {
			return
		}

		Session := GetSessionNested(*groupID, conn.ID())

		if Session == nil {
			return
		}

		conn.Emit("ice_candidate", Session.clientId, c)
	})
}

func GetSessionNested(groupID string, connID string) *Session {
	SessionManager, exists := GlobalSessionManager[groupID]
	if !exists {
		return nil
	}

	Session, exists := SessionManager.GetSession(connID)
	if !exists {
		return nil
	}

	return Session
}

func forwardTrack(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver, session *Session) {
	logger.Sugar.Debugf("Looking for slot: track=%s kind=%s", track.ID(), track.Kind())

	for _, mt := range session.transceivers {
		if mt.kind != track.Kind() {
			continue // skip if kind doesn't match
		}

		mt.mu.Lock()
		if mt.busy {
			mt.mu.Unlock()
			continue
		}
		mt.busy = true
		mt.mu.Unlock()

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			track.ID(),
			track.StreamID(),
		)
		if err != nil {
			mt.mu.Lock()
			mt.busy = false
			mt.mu.Unlock()
			return
		}

		if err := mt.t.Sender().ReplaceTrack(localTrack); err != nil {
			logger.Sugar.Warnf("ReplaceTrack failed for track %s: %v", track.ID(), err)
			mt.mu.Lock()
			mt.busy = false
			mt.mu.Unlock()
			continue
		}

		go func() {
			defer func() {
				mt.t.Sender().ReplaceTrack(nil)
				mt.mu.Lock()
				mt.busy = false
				mt.mu.Unlock()
				session.mu.Lock()
				delete(session.subscribedTracks, track.ID())
				session.mu.Unlock()
				logger.Sugar.Debugf("Slot freed for track %s", track.ID())
			}()

			buf := make([]byte, 1500)
			for {
				n, _, err := track.Read(buf)
				if err != nil {
					return
				}
				localTrack.Write(buf[:n])
			}
		}()

		logger.Sugar.Infof("Track assigned to slot: %s (kind=%s)", track.ID(), track.Kind())
		return
	}

	logger.Sugar.Warnf("No available transceiver slot for track %s (kind=%s)", track.ID(), track.Kind())
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func ptr[T any](v T) *T {
	return &v
}

func ptrUint16(v uint16) *uint16 {
	return &v
}

func RegisterHandlers(hub ws.WsHub) {
	hub.On("send_offer", func(conn ws.WebSocketConnection, data ...any) {
		handleOffer(hub, conn, data...)
	})
	hub.On("ice_candidate", handleIceCandidate)
	hub.On("track_changed", func(conn ws.WebSocketConnection, data ...any) {
		handleTrackChanged(hub, conn, data...)
	})
}

func handleOffer(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	logger.Sugar.Infof("Received offer from client %s", conn.ID())

	groupID := conn.GetGroupId()

	if groupID == nil {
		return
	}

	// 1. Setup PeerConnection
	pc, managed := MustCreatePeerConnection()
	RegisterPeerCallbacks(hub, pc, conn)

	// Track session
	if _, exists := GlobalSessionManager[*groupID]; !exists {
		logger.Sugar.Infof("Creating new session manager for group %s", *groupID)
		GlobalSessionManager[*groupID] = NewSessionManager(*groupID)
	}

	clientID, ok := data[0].(string)
	if !ok {
		return
	}

	session := &Session{pc: pc, subscribedTracks: make(map[string]*webrtc.TrackRemote), clientId: clientID, transceivers: managed}

	// 2. Parse SDP
	offerMap, ok := data[1].(map[string]interface{})
	if !ok {
		return
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerMap["sdp"].(string),
	}

	// 3. Set Remote and process buffer
	session.mu.Lock()
	Must(pc.SetRemoteDescription(offer))
	session.remoteSet = true

	// Apply buffered candidates
	for _, cand := range session.pendingCandidates {
		pc.AddICECandidate(cand)
	}

	session.pendingCandidates = nil
	session.mu.Unlock()

	// 4. Create Answer
	answer, err := pc.CreateAnswer(nil)
	Must(err)
	Must(pc.SetLocalDescription(answer))

	logger.Sugar.Infof("Sending answer to client %s", conn.ID())
	GlobalSessionManager[*groupID].AddSession(conn.ID(), session)
	conn.Emit("receive_answer", clientID, answer)
}

func handleIceCandidate(conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()

	if groupID == nil {
		return
	}

	if _, exists := GlobalSessionManager[*groupID]; !exists {
		logger.Sugar.Infof("Creating new session manager for group %s", *groupID)
		GlobalSessionManager[*groupID] = NewSessionManager(*groupID)
	}

	sm, exists := GlobalSessionManager[*groupID]
	session, exists := sm.GetSession(conn.ID())

	if !exists {
		return
	}

	candidateMap := data[1].(map[string]interface{})

	candidate := webrtc.ICECandidateInit{
		Candidate:     candidateMap["candidate"].(string),
		SDPMid:        ptr(candidateMap["sdpMid"].(string)),
		SDPMLineIndex: ptrUint16(uint16(candidateMap["sdpMLineIndex"].(float64))),
	}
	session.mu.Lock()
	defer session.mu.Unlock()

	if !session.remoteSet {
		session.pendingCandidates = append(session.pendingCandidates, candidate)
	} else {
		Must(session.pc.AddICECandidate(candidate))
	}
}

func handleTrackChanged(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	connGroupPtr := conn.GetGroupId()
	if connGroupPtr == nil {
		logger.Sugar.Warnf("Connection %s does not have a group ID, cannot emit track change", conn.ID())
		return
	}

	clientId := data[0].(string)
	trackMetadatMap := data[1].(map[string]interface{})

	trackId := trackMetadatMap["trackId"].(string)
	kind := trackMetadatMap["kind"].(string)
	streamGroupId := trackMetadatMap["streamGroupId"].(string)

	logger.Sugar.Infof("Received track change from client %s: trackId=%s, kind=%s, streamGroupId=%s", clientId, trackId, kind, streamGroupId)

	if _, exists := GlobalUserTracksMetadata[clientId]; !exists {
		GlobalUserTracksMetadata[clientId] = &UserTracksMetadata{
			tracks: []UserTrackMetadata{},
		}
	}

	GlobalUserTracksMetadata[clientId].tracks = append(GlobalUserTracksMetadata[clientId].tracks, UserTrackMetadata{
		trackId:       trackId,
		kind:          kind,
		streamGroupId: streamGroupId,
		clientId:      clientId,
	})

	hub.EmitTo(*connGroupPtr, "track_changed", map[string]interface{}{
		"clientId":      clientId,
		"trackId":       trackId,
		"kind":          kind,
		"streamGroupId": streamGroupId,
	})
}
