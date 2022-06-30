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
	defer log.Sync()
	if config, err = internal.NewConfig(); err != nil {
		log.Fatal(err)
	} else {
		enableDebug(config.Debug)
		config.Initialize()
		initContext()
		locker := etcd.NewEtcdLocker(config.EtcdConfig)
		locker.Lock()
		defer locker.Close()
		h := jobs.NewHandler(config)
		h.VerifyConfig()
		if err = h.VerifyRoles(); err == pg.UnexpctedRole {
			log.Infof("%s", err)
			locker.Close()
			return
		} else if err != nil {
			log.Panicf("error during role verification: %e", err)
		}
		h.RunSteps()
		locker.Close()
		h.RunChecks()
		jobCtxCancelFunc()
	}
}
