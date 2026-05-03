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

func WebRTCBootstrap(hub ws.WsHub) {
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

var WSIDtoClientID = make(map[string]string)

func InitPeerConnectionForSession(hub ws.WsHub, conn ws.WebSocketConnection, session Session, sesssessionManager SessionManager) {
	pc := MustCreatePeerConnection()
	RegisterPeerCallbacks(hub, pc, conn)

	session.SetEmitFunc(func(event string, data ...any) {
		conn.Emit(event, data...)
	})

	session.Init(pc)

	// register tracks for existing streams in the session
	for _, track := range sesssessionManager.GetSubscribedTracks() {
		if track.Metadata == nil || track.Track == nil {
			continue
		}

		if track.Metadata.clientId == session.GetClientId() {
			continue
		}

		session.AddRemoteTrackStream(track.Track.ID(), track.Track)
		session.AddRemoteTrackMeta(track.Track.ID(), *track.Metadata)
		session.SetOwnerSessionIdForTrack(track.Track.ID(), track.Metadata.clientId)
		session.HandleStreamForwarding(track.Metadata.trackId, track.Metadata.clientId, false)
	}

}

func HandleDisconnectByWsClient(conn ws.WebSocketConnection) {

	groupID := conn.GetGroupId()
	sessionManager, exists := GlobalSessionManager[*groupID]

	if groupID == nil || !exists {
		return
	}

	clientID, exists := WSIDtoClientID[conn.ID()]
	if !exists {
		return
	}

	Session, exists := sessionManager.GetSession(clientID)
	if !exists || Session == nil {
		return
	}

	Session.Close()

	for _, track := range Session.GetRemoteTracks() {
		Session.RemoveRemoteTrack(track.Track.ID())
	}

	sessionManager.RemoveTracksForSession(Session.GetClientId())
	sessionManager.RemoveSession(Session.GetClientId())

	for _, s := range sessionManager.GetSessions(*groupID) {
		s.RemoveRemoteTrackFromOwner(clientID)
	}

	if len(sessionManager.GetSessions(*groupID)) == 0 {
		delete(GlobalSessionManager, *groupID)
	}

	delete(WSIDtoClientID, conn.ID())

}

func RegisterPeerCallbacks(hub ws.WsHub, pc *webrtc.PeerConnection, conn ws.WebSocketConnection) {
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		groupID := conn.GetGroupId()
		sessionManager, exists := GlobalSessionManager[*groupID]

		if groupID == nil || !exists {
			return
		}

		clientID, exists := WSIDtoClientID[conn.ID()]
		if !exists {
			return
		}

		Session, exists := sessionManager.GetSession(clientID)
		if !exists || Session == nil {
			return
		}

		if !Session.IsInitialized() {
			InitPeerConnectionForSession(hub, conn, Session, sessionManager)
		}

		sessionManager.AddRemoteTrackStream(track.ID(), track)
		sessionManager.SetOwnerSessionIdForTrack(track.ID(), clientID)

		for _, s := range sessionManager.GetSessions(*groupID) {
			if s.GetClientId() == clientID || !s.IsInitialized() {
				continue
			}

			s.AddRemoteTrackStream(track.ID(), track)
			s.HandleStreamForwarding(track.ID(), clientID, true)
		}

	})

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		logger.Sugar.Infof("ICE Connection state: %s", state)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		logger.Sugar.Infof("Peer Connection state: %s", state)
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			clientID, _ := WSIDtoClientID[conn.ID()]
			HandleDisconnectByWsClient(conn)

			if clientID != "" {
				conn.Emit("peer_connection_closed", clientID, "Peer connection closed due to failure or closure")
			}
		}
	})

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		groupID := conn.GetGroupId()
		sessionManager, exists := GlobalSessionManager[*groupID]

		if groupID == nil || !exists {
			return
		}

		clientID, exists := WSIDtoClientID[conn.ID()]
		if !exists {
			return
		}

		Session, exists := sessionManager.GetSession(clientID)
		if !exists || Session == nil {
			return
		}

		conn.Emit("ice_candidate", Session.GetClientId(), c)
	})
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
	// hub.On("ice_connected", HandleLateJoin)
	hub.On("disconnect", func(conn ws.WebSocketConnection, data ...any) {
		HandleDisconnectByWsClient(conn)
	})
	hub.On("leave_room", func(conn ws.WebSocketConnection, data ...any) {
		HandleLeaveRoom(hub, conn, data...)
	})
	hub.On("join_room", func(conn ws.WebSocketConnection, data ...any) {
		HandleJoinRoom(hub, conn, data...)
	})
	hub.On("request_track_meta", HandleRequestMeta)
}
