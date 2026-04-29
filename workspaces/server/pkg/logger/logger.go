package logger

import (
	"github.com/NishLy/go-fiber-boilerplate/config"
	"go.uber.org/zap"
)

var Log *zap.Logger
var Sugar *zap.SugaredLogger

func Init() {
	cfg := config.Get()
	logger, err := func() (*zap.Logger, error) {
		if cfg.ENV == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}()

	if err != nil {
		panic(err)
	}

	Log = logger
	Sugar = logger.Sugar()
}
