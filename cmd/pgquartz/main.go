package main

import (
	"context"

	"github.com/mannemsolutions/PgQuartz/internal"
	"github.com/mannemsolutions/PgQuartz/pkg/etcd"
	"github.com/mannemsolutions/PgQuartz/pkg/git"
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
	if config, err = internal.NewConfig(); err != nil {
		initLogger("")
		log.Fatal(err)
	} else {
		initLogger(config.LogFile)
		initRemoteLoggers()
		enableDebug(config.Debug)
		config.Initialize()
		defer log.Sync() //nolint:errcheck
		if config.Git.Disable {
			log.Debug("Git pull functionality is disabled")
		} else {
			if err = git.PullCurDir(config.Workdir, config.Git); err == git.ErrRepositoryNotExists {
				log.Debugf("could not find a valid repo at %s", config.Workdir)
			} else if err == git.NoErrAlreadyUpToDate {
				log.Debugf("repo %s already up to date", config.Workdir)
			} else if err != nil {
				log.Infof("error while pulling git repo %s: %e", config.Workdir, err)
			} else {
				log.Debugf("git repo at %s updated, reapplying config", config.Workdir)
				if config, err = internal.NewConfig(); err != nil {
					log.Fatal(err)
				}
			}
		}
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
