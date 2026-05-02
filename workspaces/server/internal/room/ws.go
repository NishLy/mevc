package room

import (
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
)

func Bootstrap(hub ws.WsHub) {

	hub.On("connect", func(conn ws.WebSocketConnection, data ...any) {
	})

}
