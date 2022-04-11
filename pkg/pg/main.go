package pg

import (
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func Initialize(logger *zap.SugaredLogger) {
	log = logger
}
