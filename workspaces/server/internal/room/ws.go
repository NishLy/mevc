package room

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	rtc "github.com/NishLy/go-fiber-boilerplate/internal/webrtc"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/webrtc/v4"
)

func Bootstrap(hub ws.WsHub) {

	hub.On("connect", func(conn ws.WebSocketConnection, data ...any) {
		logger.Sugar.Infof("New connection: %s", conn.ID())
	})

	hub.On("disconnect", func(conn ws.WebSocketConnection, data ...any) {
		logger.Sugar.Infof("Disconnected: %s", conn.ID())
	})

	hub.On("join_room", func(conn ws.WebSocketConnection, data ...any) {
		roomId, ok := data[0].(string)
		if !ok || roomId == "" {
			return
		}

		hub.Join(roomId, conn)
		logger.Sugar.Infof("Connection %s joined room %s", conn.ID(), roomId)
	})

	hub.On("leave_room", func(conn ws.WebSocketConnection, data ...any) {
		if hub.GetRoom(conn) == nil {
			return
		}

		hub.Leave(*hub.GetRoom(conn), conn)
		logger.Sugar.Infof("Connection %s left room %s", conn.ID(), *hub.GetRoom(conn))
	})

	hub.On("send_offer", func(conn ws.WebSocketConnection, data ...any) {
		roomId := hub.GetRoom(conn)
		if roomId == nil {
			return
		}

		offer := &webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
		}

		offerMap, ok := data[0].(map[string]interface{})
		if !ok {
			logger.Sugar.Errorf("Invalid offer format from client %s", conn.ID())
			return
		}

		if sdp, ok := offerMap["sdp"].(string); ok {
			offer.SDP = sdp
		}

		logger.Sugar.Infof("Received offer from client %s in room %s", conn.ID(), *roomId)

		pc := rtc.MustCreatePeerConnection()
		rtc.MustAddTransceivers(pc)

		rtc.Must(pc.SetRemoteDescription(*offer))

		answer, err := pc.CreateAnswer(nil)
		rtc.Must(err)
		rtc.Must(pc.SetLocalDescription(answer))

		go func() {
			<-webrtc.GatheringCompletePromise(pc)
			logger.Sugar.Infof("Sending answer to client %s in room %s", conn.ID(), *roomId)
			conn.Emit("receive_answer", pc.LocalDescription()) // send back to same client
		}()
	})

}
