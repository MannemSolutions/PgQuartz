package etcd

import (
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func InitLogger(logger *zap.SugaredLogger) {
	log = logger
}
