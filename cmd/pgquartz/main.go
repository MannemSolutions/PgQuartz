package main

import (
	"context"

	"github.com/mannemsolutions/PgQuartz/internal"
	"github.com/mannemsolutions/PgQuartz/pkg/etcd"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"github.com/mannemsolutions/PgQuartz/pkg/pg"
)

var (
	jobCtx           context.Context
	jobCtxCancelFunc context.CancelFunc
	config           jobs.Config
)

func initContext() {
	jobCtx, jobCtxCancelFunc = config.GetTimeoutContext(context.Background())
	pg.InitContext(jobCtx)
	etcd.InitContext(jobCtx)
}

func main() {
	var err error
	initLogger()
	if config, err = internal.NewConfig(); err != nil {
		log.Fatal(err)
	} else {
		config.Initialize()
		enableDebug(config.Debug)
		config.Initialize()
		initContext()
		locker := etcd.NewEtcdLocker(config.EtcdConfig)
		locker.Lock()
		defer locker.Close()
		h := jobs.NewHandler(config)
		h.VerifyConfig()
		h.RunSteps()
		locker.Close()
		h.RunChecks()
		jobCtxCancelFunc()
	}
}
