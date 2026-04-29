package app

import (
	"github.com/NishLy/go-fiber-boilerplate/config"
	socketio "github.com/googollee/go-socket.io"
)

type App struct {
	Config *config.Config
	Io     *socketio.Server
}
