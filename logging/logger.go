package logging

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

//Logger is a logger
var Logger *logrus.Logger = logrus.New()

//LogConfig is the config for logger
type LogConfig struct {
	LogLevel   logrus.Level
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// ConfigLogger configure log settings
func ConfigLogger(config *LogConfig) {
	Logger.SetLevel(config.LogLevel)
	if len(config.Filename) > 0 {
		Logger.SetOutput(&lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize, // megabytes
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge, // days
			Compress:   false,         // disabled by default
		})
	}
}

func init() {
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	formatter.FullTimestamp = true
	Logger.SetFormatter(formatter)
	Logger.SetLevel(logrus.DebugLevel)
}
