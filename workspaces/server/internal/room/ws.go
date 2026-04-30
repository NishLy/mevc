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

	// send candidate to room
	io.OnEvent("/", "send_candidate", func(c socketio.Conn, roomID string, candidate interface{}) {
		logger.Sugar.Infof("Received candidate from client %s for room %s: %v", c.ID(), roomID, candidate)
		io.BroadcastToRoom("/", roomID, "receive_candidate", candidate)
	})

	io.OnDisconnect("/", func(s socketio.Conn, reason string) {
		logger.Sugar.Infof("Client disconnected: %s, reason: %s", s.ID(), reason)
	})

}
