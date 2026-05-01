package room

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
)

func Bootstrap(hub ws.WsHub) {

	hub.On("connect", func(conn ws.WebSocketConnection, data ...any) {
		logger.Sugar.Infof("New connection ss: %s", conn.ID())
	})
}
