package jobs

import (
	"go.uber.org/zap"
)

var (
	log  *zap.SugaredLogger
	atom zap.AtomicLevel
)

func InitLogger(logger *zap.SugaredLogger, level zap.AtomicLevel) {
	log = logger
	atom = level
}

func debug() bool {
	return atom.Level() == zap.DebugLevel
}
