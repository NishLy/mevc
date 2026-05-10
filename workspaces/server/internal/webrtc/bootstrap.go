package rtc

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

var (
	GlobalSessionManager = make(map[string]SessionManager)
)

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
		if err := me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: c.mime, ClockRate: c.clockRate},
		}, c.kind); err != nil {
			logger.Sugar.Errorf("Failed to register codec %s: %v", c.mime, err)
		}
	}
}

func MustCreatePeerConnection() *webrtc.PeerConnection {
	me := &webrtc.MediaEngine{}
	MustRegisterCodecs(me)

	s := webrtc.SettingEngine{}

	// Set ICE timeouts to detect disconnections faster and clean up resources
	// s.SetICETimeouts(120, 120*time.Second, 10*time.Second)

	ir := &interceptor.Registry{}
	pli, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		logger.Sugar.Errorf("Failed to create PLI interceptor: %v", err)
	}
	ir.Add(pli)
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		logger.Sugar.Errorf("Failed to register default interceptors: %v", err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithInterceptorRegistry(ir), webrtc.WithSettingEngine(s))
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	if err != nil {
		logger.Sugar.Errorf("Failed to create peer connection: %v", err)
	}

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

		for _, otherSession := range sessionManager.GetSessions() {
			if otherSession.GetClientId() == session.GetClientId() {
				continue
			}

			sessionManager.SuscribeToPageinatedRouters(otherSession, otherSession.GetCurrentViewPage(), MAX_STREAMS_PER_PAGE) // Resubscribe the viewer to the paginated set of tracks which will now
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

		// if state == webrtc.PeerConnectionStateConnected {
		// 	sessionManager.SuscribeToPageinatedRouters(session, session.GetCurrentViewPage(), MAX_STREAMS_PER_PAGE)
		// }

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
	// hub.On("stream_metadata_changed", func(conn ws.WebSocketConnection, data ...any) {
	// 	HandleStreamMetadataChanged(hub, conn, data...)
	// })
	hub.On("peer_status_changed", HandlePeerConnectionStateChange)
	hub.On("page_change_request", HandlePageChangeRequest)

	// This handler is for when the client requests the participant data, it will trigger a response with the current participant data of the session
	hub.On("participant_data_request", HandleParticipantDataRequest)

	// Chat handlers
	hub.On("chat_message_sent", func(conn ws.WebSocketConnection, data ...any) {
		HandleChatMessage(hub, conn, data...)
	})
	// This handler is for when the client requests the chat history, it will trigger a response with the current chat history of the session
	hub.On("chat_history_request", HandleChatHistoryRequest)

	// This handler is for when the client sends a reaction, it will trigger a response to all clients in the session with the reaction data
	hub.On("reaction_sent", func(conn ws.WebSocketConnection, data ...any) {
		HandleReaction(hub, conn, data...)
	})

	hub.On("room_metadata_request", HandleRoomMetadataRequest)
}
