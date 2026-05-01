package room

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
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

}
