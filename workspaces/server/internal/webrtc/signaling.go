package rtc

import (
	"fmt"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	pkg "github.com/NishLy/go-fiber-boilerplate/pkg/strings"
	"github.com/pion/webrtc/v4"
)

func HandleDisconnect(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	sessionManager.KickFromLoby(session)
	sessionManager.RemoveFromSessionManager(session.GetClientId())

	hub.EmitTo(sessionManager.GetGroupId(), "peer_left", nil, session.GetClientId())
}

func HandleJoinRoom(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	clientID, ok := data[0].(string)
	roomId, ok := data[1].(string)
	userName, ok2 := data[2].(string)

	if !ok || !ok2 || clientID == "" || roomId == "" || userName == "" {
		logger.Sugar.Warnf("Invalid join room data: %v", data)
		return
	}

	hub.Join(roomId, conn)

	if GetGroupManagerFromConn(conn) == nil {
		GlobalSessionManager[roomId] = NewSessionManager(roomId, true)
		setEmitFunc := func(event string, args ...interface{}) {
			hub.EmitTo(roomId, event, nil, args...)
		}
		GlobalSessionManager[roomId].SetEmitFCN(setEmitFunc)
	}
	sessionManager := GlobalSessionManager[roomId]
	pc := MustCreatePeerConnection()

	session := NewSession(pc, clientID, userName)

	session.SetEmitFunc(func(event string, args ...any) {
		conn.Emit(event, args...)
	})

	// AttachExistingStreams(sessionManager, session)
	RegisterSessionPCListeners(hub, sessionManager, session, conn)

	joined := sessionManager.AddSession(session, conn.ID())

	if !joined {
		participants := sessionManager.GetLobbySessions()
		participantsData := make([]map[string]string, len(participants))

		for i, p := range participants {
			participantsData[i] = map[string]string{
				"clientId": p.session.GetClientId(),
				"username": p.session.GetUsername(),
			}
		}

		conn.Emit("joined_lobby", roomId, participantsData)
		return
	} else {
		conn.Emit("joined_room", roomId)
	}
}

func HandleLeaveRoom(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	if hub.GetRoom(conn) == nil {
		return
	}

	hub.Leave(*hub.GetRoom(conn), conn)

	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	sessionManager.KickFromLoby(session)
	sessionManager.RemoveFromSessionManager(session.GetClientId())

	hub.EmitTo(sessionManager.GetGroupId(), "peer_left", nil, session.GetClientId())
}

func HandleOffer(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	pc := session.GetPeerConnection()

	// replace local offer with new one if exists (renegotiation)
	if session.IsRemoteSet() {
		if err := pc.SetLocalDescription(webrtc.SessionDescription{}); err != nil {
			logger.Sugar.Errorf("Failed to reset local description for session %s: %v", session.GetClientId(), err)
		}
	}

	// Block
	session.GetMutex().Lock()
	defer func() {
		session.GetMutex().Unlock()
		session.SetRemoteSet(true)
	}()

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
	if err := pc.SetRemoteDescription(offer); err != nil {
		logger.Sugar.Errorf("Failed to set remote description for session %s: %v", session.GetClientId(), err)
		return
	}

	// 4. Create Answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		logger.Sugar.Errorf("Failed to create answer for session %s: %v", session.GetClientId(), err)
		return
	}
	if err := pc.SetLocalDescription(answer); err != nil {
		logger.Sugar.Errorf("Failed to set local description for session %s: %v", session.GetClientId(), err)
		return
	}

	conn.Emit("receive_answer", session.GetClientId(), answer)
}

func HandleIceCandidate(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	candidateMap := data[1].(map[string]interface{})
	sdpMLineIndex := uint16(candidateMap["sdpMLineIndex"].(float64))

	var sdpMMid *string

	if mid, ok := candidateMap["sdpMid"].(string); ok {
		sdpMMid = &mid
	}

	candidate := webrtc.ICECandidateInit{
		Candidate:     candidateMap["candidate"].(string),
		SDPMid:        sdpMMid,
		SDPMLineIndex: &sdpMLineIndex,
	}

	if err := session.GetPeerConnection().AddICECandidate(candidate); err != nil {
		logger.Sugar.Errorf("Failed to add ICE candidate for session %s: %v", session.GetClientId(), err)
	}
}

func HandleTrackChanged(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	clientID := data[0].(string)
	trackMetadatMap := data[1].(map[string]interface{})

	if clientID == "" || trackMetadatMap["trackId"] == nil || trackMetadatMap["kind"] == nil || trackMetadatMap["streamGroupId"] == nil {
		logger.Sugar.Warnf("Invalid track change data from client %s: %v", clientID, trackMetadatMap)
		return
	}

	trackId := trackMetadatMap["trackId"].(string)
	streamId := trackMetadatMap["streamId"].(string)
	kind := trackMetadatMap["kind"].(string)
	streamGroupId := trackMetadatMap["streamGroupId"].(string)
	label := trackMetadatMap["label"].(string)
	enabled := trackMetadatMap["enabled"].(bool)
	username := trackMetadatMap["username"].(string)

	metadata := SessionTrackMetadata{
		TrackId:       trackId,
		Kind:          kind,
		StreamGroupId: streamGroupId,
		ClientId:      clientID,
		StreamId:      streamId,
		Label:         label,
		Enabled:       enabled,
		Username:      username,
	}

	session, exist := sessionManager.GetSession(clientID)

	if !exist || session == nil {
		return
	}

	_, exist = sessionManager.GetRouter(streamId)

	if !exist {
		newRouter := NewTrackRouter(session.GetPeerConnection(), clientID)
		newRouter.SetMetadata(&metadata)
		newRouter.SetStreamID(streamId)
		newRouter.SetStreamGroupId(streamGroupId)

		newRouter.Start()

		sessionManager.AddRouter(streamId, newRouter)

		hub.EmitTo(sessionManager.GetGroupId(), "new_track", &conn, session.GetClientId(), &metadata)
	} else {
		hub.EmitTo(sessionManager.GetGroupId(), "track_changed", &conn, session.GetClientId(), &metadata)
	}

	// logger.Sugar.Infof("Client %s changed track %s (kind=%s) in stream group %s", clientID, streamId, kind, streamGroupId)
}

func HandleRenegotiateAnswer(conn ws.WebSocketConnection, data ...any) (err error) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	defer func() {
		if err != nil {
			if err := session.GetPeerConnection().SetLocalDescription(webrtc.SessionDescription{}); err != nil {
				logger.Sugar.Errorf("Failed to reset local description for session %s: %v", session.GetClientId(), err)
			}
			<-session.GetOfferWaitChan()
		}
	}()

	answerMap, ok := data[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid answer map")
	}

	sdp, ok := answerMap["sdp"].(string)
	if !ok {
		return fmt.Errorf("invalid sdp")
	}

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}

	if err = session.GetPeerConnection().SetRemoteDescription(answer); err != nil {
		return
	}

	<-session.GetOfferWaitChan()

	return nil
}

func HandleRemoveTrack(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	sessionManager.RemoveRouter(data[1].(string))
}

func HandlePeerConnectionStateChange(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	state := data[1].(string)
	if state == "" {
		return
	}

	logger.Sugar.Infof("Status Peer %s", session.GetPeerConnection().ConnectionState().String())

	switch state {
	case "connected":
		sessionManager.SuscribeToPageinatedRouters(session, session.GetCurrentViewPage(), MAX_STREAMS_PER_PAGE)
		chatHistory := sessionManager.GetChatService().GetHistory(50, 0)

		conn.Emit("chat_history", chatHistory)
	default:
		logger.Sugar.Infof("Peer Status Changed But No Handler Implemented for State: %s", state)
		// Handle other states if needed`
	}

}

func HandlePageChangeRequest(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	page := data[1].(float64)
	if page == 0 {
		return
	}

	pageInt := int(page)

	if err := sessionManager.SuscribeToPageinatedRouters(session, pageInt, MAX_STREAMS_PER_PAGE); err != nil {
		logger.Sugar.Errorf("Failed to subscribe to paginated routers for session %s: %v", session.GetClientId(), err)
		return
	}

	session.SetCurrrentViewPage(pageInt)
}

func HandleParticipantDataRequest(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	participantsData := sessionManager.GetParticipantsData()

	conn.Emit("participants_data_response", participantsData)
}

func HandleChatMessage(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	message, ok := data[1].(string)
	if !ok || message == "" {
		return
	}

	message = pkg.TruncateString(message, MAX_CHAT_MESSAGE_LENGTH)

	chatMessage := ChatMessage{
		SenderID:   session.GetClientId(),
		SenderName: session.GetUsername(),
		Message:    message,
		Timestamp:  time.Now(),
	}

	sessionManager.GetChatService().AddMessage(chatMessage)

	hub.EmitTo(sessionManager.GetGroupId(), "chat_message_sent", &conn, chatMessage)
}

func HandleChatHistoryRequest(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	_, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	lastN := 50
	skip := int(data[1].(float64))

	chatHistory := sessionManager.GetChatService().GetHistory(lastN, skip)

	conn.Emit("chat_history_response", chatHistory)
}
