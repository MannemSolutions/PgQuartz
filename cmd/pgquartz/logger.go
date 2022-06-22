package main

import (
	"os"

	"github.com/mannemsolutions/PgQuartz/pkg/etcd"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"github.com/mannemsolutions/PgQuartz/pkg/pg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.SugaredLogger
	atom zap.AtomicLevel
)

func initLogger() {
	atom = zap.NewAtomicLevel()
	//encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	log = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	)).Sugar()
	jobs.InitLogger(log, atom)
	pg.InitLogger(log)
	etcd.InitLogger(log)
}

func enableDebug(debug bool) {
	if debug {
		atom.SetLevel(zap.DebugLevel)
	}
	log.Debug("Debug logging enabled by config")
}
