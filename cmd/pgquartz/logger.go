package main

import (
	"github.com/mannemsolutions/PgQuartz/pkg/etcd"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"github.com/mannemsolutions/PgQuartz/pkg/pg"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.SugaredLogger
	atom zap.AtomicLevel
)

func initLogger(logFilePath string) {

	atom = zap.NewAtomicLevel()
	// First, define our level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= atom.Level()
	})

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Optimize the Kafka output for machine consumption and the console output
	// for human operators.
	//encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the cores together.
	var core zapcore.Core
	if logFilePath != "" {
		fileEncoder := zapcore.NewConsoleEncoder(encoderCfg)
		// #nosec G304,G302 -- path from variable is ok in this case (pgquartz is run by a user with low OS permissions)
		if logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			initLogger("")
			log.Panicf("error while opening logfile: %s", err)
		} else {
			writer := zapcore.AddSync(logFile)
			core = zapcore.NewTee(
				zapcore.NewCore(fileEncoder, writer, atom),
				zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
				zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
			)
		}
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
		)
	}

	log = zap.New(core).Sugar()
}

func initRemoteLoggers() {
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
