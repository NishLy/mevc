package rtc

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type Session struct {
	pc                *webrtc.PeerConnection
	pendingCandidates []webrtc.ICECandidateInit
	remoteSet         bool
	mu                sync.Mutex
}

var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

type udpConn struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}

var udpConns = map[string]*udpConn{
	"audio": {port: 4000, payloadType: 111},
	"video": {port: 4002, payloadType: 96},
}

func WebRTCBootstrap(hub ws.WsHub) {
	mustDialUDP()
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

func mustDialUDP() {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")
	Must(err)

	for _, c := range udpConns {
		raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", c.port))
		Must(err)
		c.conn, err = net.DialUDP("udp", laddr, raddr)
		Must(err)
	}
}

func RegisterPeerCallbacks(pc *webrtc.PeerConnection, conn ws.WebSocketConnection) {
	pc.OnTrack(forwardTrack)

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		logger.Sugar.Infof("ICE Connection state: %s", state)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		logger.Sugar.Infof("Peer Connection state: %s", state)
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			mu.Lock()
			delete(sessions, conn.ID())
			mu.Unlock()
			conn.Emit("peer_connection_closed", "Peer connection closed due to failure or closure")
		}
	})

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		logger.Sugar.Infof("New ICE candidate for client %s: %s", conn.ID(), c.String())
		conn.Emit("ice_candidate", c)
	})
}

func forwardTrack(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	conn, ok := udpConns[track.Kind().String()]
	if !ok {
		return
	}

	buf := make([]byte, 1500)
	pkt := &rtp.Packet{}

	for {
		n, _, err := track.Read(buf)
		if err != nil {
			logger.Sugar.Infof("Track ended: %v", err)
			return
		}

		if err := pkt.Unmarshal(buf[:n]); err != nil {
			logger.Sugar.Errorf("Unmarshal error: %v", err)
			continue
		}

		// pkt.PayloadType = conn.payloadType

		n, err = pkt.MarshalTo(buf)
		if err != nil {
			logger.Sugar.Errorf("Marshal error: %v", err)
			continue
		}

		if _, err = conn.conn.Write(buf[:n]); err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Err.Error() == "write: connection refused" {
				continue
			}
			logger.Sugar.Errorf("UDP write error: %v", err)
			return
		}
	}
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

	// 1. Setup PeerConnection
	pc := MustCreatePeerConnection()
	MustAddTransceivers(pc)
	RegisterPeerCallbacks(pc, conn)

	session := &Session{pc: pc}

	// Track session
	mu.Lock()
	sessions[conn.ID()] = session
	mu.Unlock()

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
	logger.Sugar.Infof("Received ICE candidate from client %s", conn.ID())

	mu.RLock()
	session, exists := sessions[conn.ID()]
	mu.RUnlock()
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
