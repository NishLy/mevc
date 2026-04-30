package room

import (
	"fmt"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	socketio "github.com/googollee/go-socket.io"
)

func RoomWsBootstrap(io *socketio.Server) {

	// 1. Handle Connection
	io.OnConnect("/", func(c socketio.Conn) error {
		fmt.Printf("New client connected: %s\n", c.ID())
		logger.Sugar.Infof("Backend connected: %s", c.ID())
		return nil
	})

	// join room
	io.OnEvent("/", "join_room", func(c socketio.Conn, roomID string) {
		logger.Sugar.Infof("Client %s joining room: %s", c.ID(), roomID)
		c.Join(roomID)
	})

	// leave room
	io.OnEvent("/", "leave_room", func(c socketio.Conn, roomID string) {
		logger.Sugar.Infof("Client %s leaving room: %s", c.ID(), roomID)
		c.Leave(roomID)
	})

}
