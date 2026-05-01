package app

import (
	"github.com/NishLy/go-fiber-boilerplate/config"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/ws"
)

type App struct {
	Config *config.Config
	WsHub  ws.WsHub
}
