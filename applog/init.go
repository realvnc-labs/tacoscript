package applog

import (
	log "github.com/sirupsen/logrus"
)

//Init logging entry point
func Init(isVerbose bool) {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000000",
		FullTimestamp:   true,
	})

	logLevel := log.WarnLevel
	if isVerbose {
		logLevel = log.DebugLevel
	}

	log.SetLevel(logLevel)

	log.Debugf("will set log level to %v", logLevel)
}
