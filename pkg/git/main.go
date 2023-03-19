package git

import (
	//"context"

	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
	//ctx context.Context
)

func InitLogger(logger *zap.SugaredLogger) {
	log = logger
}

//func InitContext(c context.Context) {
//	ctx = c
//}
