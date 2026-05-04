package rtc

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

var MAXIMUM_CONNECTION = 20

var (
	GlobalSessionManager = make(map[string]SessionManager)
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func GetGroupManagerFromConn(conn ws.WebSocketConnection) SessionManager {
	groupId := conn.GetGroupId()
	if groupId == nil {
		return nil
	}

	if _, exists := GlobalSessionManager[*groupId]; !exists {
		return nil
	}

	return GlobalSessionManager[*groupId]
}

func WebRTCBootstrap(hub ws.WsHub) {
	RegisterHandlers(hub)
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

func RegisterPCCallbacks(hub ws.WsHub, sessionManager SessionManager, session Session, conn ws.WebSocketConnection) {
	session.GetPeerConnection().OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		sessionManager.AddRemoteTrackStream(track.ID(), track)
		sessionManager.SetOwnerSessionIdForTrack(track.ID(), session.GetClientId())

		for _, s := range sessionManager.GetSessions() {
			if s.GetClientId() == session.GetClientId() || !s.IsInitialized() {
				continue
			}

			s.AddRemoteTrackStream(track.ID(), track)
			s.HandleStreamForwarding(track.ID(), s.GetClientId())
		}
	})

	session.GetPeerConnection().OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		conn.Emit("ice_candidate", session.GetClientId(), c)
	})

	session.GetPeerConnection().OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		logger.Sugar.Infof("ICE Connection state: %s", state)
	})

	session.GetPeerConnection().OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		logger.Sugar.Infof("Peer Connection state: %s", state)
		if state == webrtc.PeerConnectionStateConnected {
			session.RunWorker()
			BoostrapSession(sessionManager, session)
		}

		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			go RemoveFromSessionManager(hub, sessionManager, session.GetClientId())

			if session.GetClientId() != "" {
				conn.Emit("peer_connection_closed", session.GetClientId(), "Peer connection closed due to failure or closure")
			}
		}
	})
}

func RemoveFromSessionManager(hub ws.WsHub, sessionManager SessionManager, clientID string) {
	session, exists := sessionManager.GetSession(clientID)
	if !exists || session == nil {
		return
	}

	session.Close()
	sessionManager.RemoveTracksForSession(session.GetClientId())
	sessionManager.RemoveSession(session.GetClientId())

	for _, othersSession := range sessionManager.GetSessions() {
		othersSession.RemoveRemoteTrackFromOwner(session.GetClientId())
	}

	if len(sessionManager.GetSessions()) == 0 {
		delete(GlobalSessionManager, sessionManager.GetGroupId())
	}

	hub.EmitTo(sessionManager.GetGroupId(), "peer_left", nil, clientID)
}

func BoostrapSession(sessionManager SessionManager, session Session) {
	for _, track := range sessionManager.GetSubscribedTracks() {
		if track.Metadata == nil || track.Track == nil {
			continue
		}

		if track.Metadata.clientId == session.GetClientId() {
			continue
		}

		session.AddRemoteTrackStream(track.Track.ID(), track.Track)
		session.AddRemoteTrackMeta(track.Track.ID(), *track.Metadata)
		session.SetOwnerSessionIdForTrack(track.Track.ID(), track.Metadata.clientId)
		session.HandleStreamForwarding(track.Metadata.trackId, track.Metadata.clientId)
	}
}

func RegisterHandlers(hub ws.WsHub) {
	hub.On("send_offer", func(conn ws.WebSocketConnection, data ...any) {
		HandleOffer(hub, conn, data...)
	})
	hub.On("send_answer", func(conn ws.WebSocketConnection, data ...any) {
		HandleRenegotiateAnswer(conn, data...)
	})
	hub.On("ice_candidate", HandleIceCandidate)
	hub.On("track_changed", func(conn ws.WebSocketConnection, data ...any) {
		handleTrackChanged(conn, data...)
	})

	hub.On("leave_room", func(conn ws.WebSocketConnection, data ...any) {
		HandleLeaveRoom(hub, conn, data...)
	})
	hub.On("join_room", func(conn ws.WebSocketConnection, data ...any) {
		HandleJoinRoom(hub, conn, data...)
	})
	hub.On("request_track_meta", HandleRequestMeta)
	hub.On("disconnect", func(conn ws.WebSocketConnection, data ...any) {
		HandleDisconnect(hub, conn, data...)
	})
	hub.On("track_removed", HandleRemoveTrack)
}
