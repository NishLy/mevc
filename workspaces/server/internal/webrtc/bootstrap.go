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

func RegisterSessionPCListeners(hub ws.WsHub, sessionManager SessionManager, session Session, conn ws.WebSocketConnection) {
	session.GetPeerConnection().OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {

		router, exist := sessionManager.GetRouter(track.StreamID())

		if !exist {
			router = NewTrackRouter(session.GetPeerConnection(), session.GetClientId())
		}

		router.SetIncomingTrack(track)
		router.Start() // Start the router to begin forwarding this track to viewers

		sessionManager.AddRouter(track.StreamID(), router)

		var sessionsToBeRenegotiated []Session

		for _, otherSession := range sessionManager.GetSessions() {
			if otherSession.GetClientId() == session.GetClientId() {
				continue
			}

			err := router.AddViewer(otherSession)

			if err != nil {
				logger.Sugar.Errorf("Failed to add viewer for session %s: %v", otherSession.GetClientId(), err)
				continue
			}

			sessionsToBeRenegotiated = append(sessionsToBeRenegotiated, otherSession)
		}

		for _, otherSession := range sessionsToBeRenegotiated {
			err := otherSession.Renegotiate(nil)
			if err != nil {
				logger.Sugar.Errorf("Failed to renegotiate for session %s: %v", otherSession.GetClientId(), err)
				continue
			}
		}

		if router.hasStarted && router.metadata != nil {
			hub.EmitTo(sessionManager.GetGroupId(), "new_track", &conn, session.GetClientId(), map[string]interface{}{
				"trackId":       track.ID(),
				"streamId":      track.StreamID(),
				"kind":          track.Kind().String(),
				"clientId":      session.GetClientId(),
				"streamGroupId": router.metadata.streamGroupId,
				"label":         router.metadata.label,
			})
		}

		logger.Sugar.Infof("Client %s added track %s (kind=%s) to stream group %s", session.GetClientId(), track.StreamID(), track.Kind().String(), router.metadata.streamGroupId)
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
			sessionManager.SubscribeToExistingTracks(session)
		}

		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			sessionManager.RemoveFromSessionManager(session.GetClientId())
			hub.EmitTo(sessionManager.GetGroupId(), "peer_connection_closed", nil, session.GetClientId(), "Peer connection closed due to failure or closure")
		}
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
		HandleTrackChanged(hub, conn, data...)
	})

	hub.On("leave_room", func(conn ws.WebSocketConnection, data ...any) {
		HandleLeaveRoom(hub, conn, data...)
	})
	hub.On("join_room", func(conn ws.WebSocketConnection, data ...any) {
		HandleJoinRoom(hub, conn, data...)
	})
	// hub.On("request_track_meta", HandleRequestMeta)
	hub.On("disconnect", func(conn ws.WebSocketConnection, data ...any) {
		HandleDisconnect(hub, conn, data...)
	})
	hub.On("track_removed", HandleRemoveTrack)
}
