package main

import (
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
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
	jobs.InitLogger(log)
}
