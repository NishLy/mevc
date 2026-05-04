package rtc

import (
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

	go RemoveFromSessionManager(hub, sessionManager, session.GetClientId())
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

	RemoveFromSessionManager(hub, sessionManager, session.GetClientId())
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
	session.SetRemoteSet(true)

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

func handleTrackChanged(conn ws.WebSocketConnection, data ...any) {
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
	kind := trackMetadatMap["kind"].(string)
	streamGroupId := trackMetadatMap["streamGroupId"].(string)

	metadata := SessionTrackMetadata{
		trackId:       trackId,
		kind:          kind,
		streamGroupId: streamGroupId,
		clientId:      clientID,
	}

	session, exists := sessionManager.GetSession(clientID)

	if !exists || session == nil {
		logger.Sugar.Warnf("Session not found for client %s when handling track change", clientID)
		return
	}

	session.AddSelfTrackMetadata(trackId, metadata)
	sessionManager.AddRemoteTrackMeta(trackId, metadata)
	sessionManager.SetOwnerSessionIdForTrack(trackId, metadata.clientId)

	var sessionsToBeRenegotiated []Session
	for _, otherSession := range sessionManager.GetSessions() {
		if otherSession.GetClientId() == clientID {
			continue
		}

		otherSession.AddRemoteTrackMeta(trackId, metadata)
		streamedSession := otherSession.StartRTPStream(trackId, metadata.clientId)

		if streamedSession != nil {
			sessionsToBeRenegotiated = append(sessionsToBeRenegotiated, streamedSession)
		}
	}

	for _, otherSession := range sessionsToBeRenegotiated {
		if err := otherSession.Renegotiate(nil); err != nil {
			logger.Sugar.Errorf("Failed to renegotiate for session %s: %v", otherSession.GetClientId(), err)
		}
	}

	logger.Sugar.Infof("Client %s added track %s (kind=%s) to stream group %s", clientID, trackId, kind, streamGroupId)
}

func HandleRenegotiateAnswer(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	answerMap, ok := data[1].(map[string]interface{})
	if !ok {
		return
	}

	sdp, ok := answerMap["sdp"].(string)
	if !ok {
		return
	}

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}

	Must(session.GetPeerConnection().SetRemoteDescription(answer))

	<-session.GetOfferWaitChan()
}

func HandleRequestMeta(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	metaMap := data[1].(map[string]interface{})

	if metaMap["transceiverMid"] == nil {
		logger.Sugar.Warnf("Invalid track metadata from client %s: %v", session.GetClientId(), metaMap)
		return
	}

	if session.GetClientId() == "" {
		logger.Sugar.Warnf("Invalid late join data: clientId=%s", session.GetClientId())
		return
	}

	pc := session.GetPeerConnection()

	for _, track := range pc.GetTransceivers() {
		if track.Mid() == metaMap["transceiverMid"].(string) {
			sessionTrack, exist := session.GetRemoteTrack(track.Sender().Track().ID())

			if !exist {
				logger.Sugar.Warnf("No track found for MID %s in session of client %s", track.Mid(), session.GetClientId())
				return
			}

			conn.Emit("new_track", sessionTrack.Metadata.clientId, map[string]interface{}{
				"clientId":       sessionTrack.Metadata.clientId,
				"trackId":        sessionTrack.Metadata.trackId,
				"kind":           sessionTrack.Metadata.kind,
				"streamGroupId":  sessionTrack.Metadata.streamGroupId,
				"transceiverMid": track.Mid(),
			})
		}
	}

}

func HandleRemoveTrack(conn ws.WebSocketConnection, data ...any) {
	sessionManager := GetGroupManagerFromConn(conn)
	if sessionManager == nil {
		return
	}

	session, exists := sessionManager.GetSessionByWsID(conn.ID())
	if !exists {
		return
	}

	streamGroupId := data[1].(string)

	metadata, exists := session.GetSelfTracksMetadata(streamGroupId)
	if !exists {
		logger.Sugar.Warnf("No track metadata found for streamGroupId %s in session of client %s", streamGroupId, session.GetClientId())
		return
	}

	session.RemoveSelfTrackMetadata(metadata.trackId)

	sessionManager.RemoveSubscribedTrack(metadata.trackId)
	for _, otherSession := range sessionManager.GetSessions() {
		if otherSession.GetClientId() == session.GetClientId() {
			continue
		}

		otherSession.RemoveRemoteTrack(metadata.trackId)

		if err := otherSession.Renegotiate(nil); err != nil {
			logger.Sugar.Errorf("Failed to renegotiate for session %s: %v", otherSession.GetClientId(), err)
		}
	}
}
