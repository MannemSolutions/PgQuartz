package pg

import (
	"context"

	"go.uber.org/zap"
)

var (
	log        *zap.SugaredLogger
	ctx        context.Context
	ValidRoles = map[string]bool{
		"primary": true,
		"standby": true,
	}
)

func InitLogger(logger *zap.SugaredLogger) {
	log = logger
}
func InitContext(c context.Context) {
	ctx = c
}
