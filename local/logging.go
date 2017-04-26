package local

import (
	log "github.com/Sirupsen/logrus"
)

func SetLoggingLevel(verbosity int) {
	switch verbosity {
	case 0:
		log.SetLevel(log.ErrorLevel)
	case 1:
		log.SetLevel(log.WarnLevel)
	case 2:
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}
