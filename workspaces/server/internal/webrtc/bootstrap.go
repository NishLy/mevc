package rtc

import (
	"fmt"
	"sync"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

type Session struct {
	pc                *webrtc.PeerConnection
	subscribedTracks  map[string]*webrtc.TrackRemote
	pendingCandidates []webrtc.ICECandidateInit
	remoteSet         bool
	mu                sync.Mutex
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

var MAXIMUM_TRANCEIVERS = 10

var (
	GlobalSessionManager = make(map[string]*SessionManager)
)

func WebRTCBootstrap(hub ws.WsHub) {
	// mustDialUDP()
	RegisterHandlers(hub)
}

func MustCreatePeerConnection() *webrtc.PeerConnection {
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

	for i := 0; i < MAXIMUM_TRANCEIVERS; i++ {
		pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
		pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
	}

	Must(err)
	return pc
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

func MustAddTransceivers(pc *webrtc.PeerConnection) {
	for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeAudio, webrtc.RTPCodecTypeVideo} {
		_, err := pc.AddTransceiverFromKind(kind)
		Must(err)
	}
}

// func mustDialUDP() {
// 	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")
// 	Must(err)

// 	for _, c := range udpConns {
// 		raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", c.port))
// 		Must(err)
// 		c.conn, err = net.DialUDP("udp", laddr, raddr)
// 		Must(err)
// 	}
// }

func RegisterPeerCallbacks(pc *webrtc.PeerConnection, conn ws.WebSocketConnection) {
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		logger.Sugar.Infof("Received new track from client %s: %s", conn.ID(), track.ID())

		groupID := conn.GetGroupId()
		if groupID == nil {
			return
		}

		sessionManager, exists := GlobalSessionManager[*groupID]

		if exists {
			for _, s := range sessionManager.sessions {
				if s.pc != pc {
					forwardTrack(track, receiver, s.pc)
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

			Session, exists := GlobalSessionManager[*groupID].GetSession(conn.ID())
			if exists {
				Session.pc.Close()
				GlobalSessionManager[*groupID].RemoveSession(conn.ID())
			}

			if len(GlobalSessionManager[*groupID].sessions) == 0 {
				delete(GlobalSessionManager, *groupID)
			}

			conn.Emit("peer_connection_closed", "Peer connection closed due to failure or closure")
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

		conn.Emit("ice_candidate", c)
	})
}

func forwardTrack(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver, peer *webrtc.PeerConnection) {
	for _, t := range peer.GetTransceivers() {
		sender := t.Sender()

		// find empty slot
		if sender.Track() == nil {
			localTrack, _ := webrtc.NewTrackLocalStaticRTP(
				track.Codec().RTPCodecCapability,
				track.ID(),
				track.StreamID(),
			)

			sender.ReplaceTrack(localTrack)

			// start forwarding RTP
			go func() {
				buf := make([]byte, 1500)
				for {
					n, _, err := track.Read(buf)
					if err != nil {
						return
					}
					localTrack.Write(buf[:n])
				}
			}()

			fmt.Println("Track assigned to slot")
			return
		}
	}

	fmt.Println("No available slot")
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
	hub.On("send_offer", handleOffer)
	hub.On("ice_candidate", handleIceCandidate)
}

func handleOffer(conn ws.WebSocketConnection, data ...any) {
	logger.Sugar.Infof("Received offer from client %s", conn.ID())

	groupID := conn.GetGroupId()

	if groupID == nil {
		return
	}

	// 1. Setup PeerConnection
	pc := MustCreatePeerConnection()
	MustAddTransceivers(pc)
	RegisterPeerCallbacks(pc, conn)

	session := &Session{pc: pc}

	// Track session
	logger.Sugar.Infof("Creating new session for connection %s in group %s", conn.ID(), *groupID)
	if _, exists := GlobalSessionManager[*groupID]; !exists {
		GlobalSessionManager[*groupID] = NewSessionManager(*groupID)
	}

	GlobalSessionManager[*groupID].AddSession(conn.ID(), session)

	// 2. Parse SDP
	offerMap, ok := data[0].(map[string]interface{})
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
	conn.Emit("receive_answer", answer)
}

func handleIceCandidate(conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()
	if groupID == nil {
		return
	}

	sm, exists := GlobalSessionManager[*groupID]
	session, exists := sm.GetSession(conn.ID())

	if !exists {
		return
	}

	candidateMap := data[0].(map[string]interface{})

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
