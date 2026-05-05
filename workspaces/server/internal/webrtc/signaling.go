package rtc

import (
	"fmt"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
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

	sessionManager.RemoveFromSessionManager(session.GetClientId())

	hub.EmitTo(sessionManager.GetGroupId(), "peer_left", nil, session.GetClientId())
}

func HandleJoinRoom(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	clientID, ok := data[0].(string)

	if !ok || clientID == "" {
		return
	}

	roomId, ok := data[1].(string)
	if !ok || roomId == "" {
		return
	}

	hub.Join(roomId, conn)

	if GetGroupManagerFromConn(conn) == nil {
		GlobalSessionManager[roomId] = NewSessionManager(roomId)
	}
	sessionManager := GlobalSessionManager[roomId]
	pc := MustCreatePeerConnection()

	session := NewSession(clientID, pc)

	session.SetEmitFunc(func(event string, args ...any) {
		conn.Emit(event, args...)
	})

	// AttachExistingStreams(sessionManager, session)
	RegisterSessionPCListeners(hub, sessionManager, session, conn)

	sessionManager.AddSession(session, conn.ID())

	conn.Emit("joined_room", roomId)
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
		Must(pc.SetLocalDescription(webrtc.SessionDescription{}))
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
	Must(pc.SetRemoteDescription(offer))

	// 4. Create Answer
	answer, err := pc.CreateAnswer(nil)
	Must(err)
	Must(pc.SetLocalDescription(answer))

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

	Must(session.GetPeerConnection().AddICECandidate(candidate))
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
	label := ""

	if trackMetadatMap["label"] != nil {
		label = trackMetadatMap["label"].(string)
	}

	metadata := SessionTrackMetadata{
		trackId:       trackId,
		kind:          kind,
		streamGroupId: streamGroupId,
		clientId:      clientID,
		streamId:      streamId,
		label:         label,
	}

	session, exists := sessionManager.GetSession(clientID)

	if !exists || session == nil {
		return
	}

	_, exist := sessionManager.GetRouter(streamId)

	if !exist {
		newRouter := NewTrackRouter(session.GetPeerConnection(), clientID)
		newRouter.SetMetadata(&metadata)

		// try to start
		newRouter.Start()

		sessionManager.AddRouter(streamId, newRouter)
		return
	}

	hub.EmitTo(sessionManager.GetGroupId(), "new_track", &conn, session.GetClientId(), map[string]interface{}{
		"trackId":       trackId,
		"streamId":      streamId,
		"kind":          kind,
		"clientId":      clientID,
		"streamGroupId": streamGroupId,
		"label":         label,
	})

	logger.Sugar.Infof("Client %s changed track %s (kind=%s) in stream group %s", clientID, streamId, kind, streamGroupId)
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
			Must(session.GetPeerConnection().SetLocalDescription(webrtc.SessionDescription{}))
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
