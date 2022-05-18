package pg

import (
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
	ValidRoles = map[string]bool{
		"primary": true,
		"standby": true,
	}
)

func Initialize(logger *zap.SugaredLogger) {
	log = logger
}