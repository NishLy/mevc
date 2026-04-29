package room

import (
	"log"

	socketio "github.com/googollee/go-socket.io"
)

func RoomWsBootstrap(io *socketio.Server) {

	io.OnConnect("/", func(s socketio.Conn) error {
		log.Println("connected:", s.ID())
		return nil
	})

	io.OnEvent("/", "join", func(s socketio.Conn, room string) {
		s.Join(room)
		log.Println("join room:", room)
	})

	io.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("disconnect:", s.ID())
	})

}
