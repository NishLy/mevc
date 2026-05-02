package rtc

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/webrtc/v4"
)

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

	if _, exists := GlobalSessionManager[roomId]; !exists {
		GlobalSessionManager[roomId] = NewSessionManager(roomId)
	}

	GlobalSessionManager[roomId].AddSession(NewSession(clientID))
	conn.Emit("joined_room", roomId)
	WSIDtoClientID[conn.ID()] = clientID
}

func HandleLeaveRoom(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	if hub.GetRoom(conn) == nil {
		return
	}

	hub.Leave(*hub.GetRoom(conn), conn)
	HandleDisconnectByWsClient(conn)
}

func HandleOffer(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()

	if groupID == nil {
		return
	}

	if _, exists := GlobalSessionManager[*groupID]; !exists {
		return
	}

	clientID, ok := data[0].(string)
	if !ok {
		return
	}

	session, exists := GlobalSessionManager[*groupID].GetSession(clientID)

	if !exists {
		return
	}

	WSIDtoClientID[conn.ID()] = clientID

	if !session.IsInitialized() {
		InitPeerConnectionForSession(hub, conn, session)
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

	GlobalSessionManager[*groupID].AddSession(session)

	conn.Emit("receive_answer", clientID, answer)
}

func HandleIceCandidate(conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()

	if groupID == nil {
		return
	}

	clientID, ok := data[0].(string)
	if !ok {
		return
	}

	if _, exists := GlobalSessionManager[*groupID]; !exists {
		return
	}

	session, exists := GlobalSessionManager[*groupID].GetSession(clientID)
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

func handleTrackChanged(hub ws.WsHub, conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()

	if groupID == nil {
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

	sessionManager, exists := GlobalSessionManager[*groupID]

	if !exists {
		// logger.Sugar.Warnf("No session manager found for group %s when handling track change", *groupID)
		return
	}

	metadata := SessionTrackMetadata{
		trackId:       trackId,
		kind:          kind,
		streamGroupId: streamGroupId,
		clientId:      clientID,
	}

	// session.AddRemoteTrackMeta(trackId, metadata)
	// session.SetOwnerSessionIdForTrack(trackId, session.GetClientId())
	sessionManager.AddRemoteTrackMeta(trackId, metadata)
	sessionManager.SetOwnerSessionIdForTrack(trackId, metadata.clientId)

	hub.EmitTo(*groupID, "new_track", &conn, map[string]interface{}{
		"clientId":      clientID,
		"trackId":       trackId,
		"kind":          kind,
		"streamGroupId": streamGroupId,
	})
}

// HandleRenegotiateAnswer processes the subscriber client's SDP answer
// to a server-initiated renegotiation offer (sent after forwardTrack).
func HandleRenegotiateAnswer(conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()
	if groupID == nil {
		return
	}

	clientID, ok := data[0].(string)
	if !ok || clientID == "" {
		return
	}

	if _, exists := GlobalSessionManager[*groupID]; !exists {
		return
	}

	session, exists := GlobalSessionManager[*groupID].GetSession(clientID)
	if !exists || session == nil {
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
}

func HandleLateJoin(conn ws.WebSocketConnection, data ...any) {
	groupID := conn.GetGroupId()
	clientId := data[0].(string)

	if groupID == nil || clientId == "" {
		logger.Sugar.Warnf("Invalid late join data: groupID=%v, clientId=%s", groupID, clientId)
		return
	}

	if sessionManager, exists := GlobalSessionManager[*groupID]; exists {
		for _, track := range sessionManager.GetSubscribedTracks() {
			if track.Metadata == nil {
				continue
			}

			if track.Metadata.clientId == clientId {
				continue
			}

			conn.Emit("new_track", map[string]interface{}{
				"clientId":      track.Metadata.clientId,
				"trackId":       track.Metadata.trackId,
				"kind":          track.Metadata.kind,
				"streamGroupId": track.Metadata.streamGroupId,
			})
		}
	}
}
