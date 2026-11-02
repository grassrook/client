package logger

import (
	"path"

	"github.com/amoghe/distillog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(configSrc string) distillog.Logger {
	writer := &lumberjack.Logger{
		Filename:   path.Join(configSrc, "app.log"),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	return distillog.NewStreamLogger("grassrook", writer)
}
