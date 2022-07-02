package main

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mannemsolutions/PgQuartz/internal"
	"github.com/mannemsolutions/PgQuartz/pkg/etcd"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"github.com/mannemsolutions/PgQuartz/pkg/pg"
	"io/ioutil"
	"strings"
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

func getSshAuth(repo string, config jobs.GitConfig) (transport.AuthMethod, error) {
	if strings.HasPrefix(repo, "http") {
		return &http.BasicAuth{
			Username: config.HttpUser,
			Password: config.HttpPassword,
		}, nil
	} else if sshKey, err := ioutil.ReadFile(config.RsaPath); err != nil {
		return nil, err
	} else {
		return ssh.NewPublicKeys("git", []byte(sshKey), "")
	}
}

func pullCurDir(workDir string, config jobs.GitConfig) error {
	if r, err := git.PlainOpenWithOptions(workDir, &git.PlainOpenOptions{DetectDotGit: true}); err != nil {
		return err
	} else if w, err := r.Worktree(); err != nil {
		return err
	} else if remote, err := r.Remote(config.Remote); err != nil {
		return err
	} else if auth, err := getSshAuth(remote.String(), config); err != nil {
		return err
	} else if err = w.Pull(&git.PullOptions{RemoteName: config.Remote, Auth: auth}); err != nil {
		return err
	}
	return nil
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
			if err = pullCurDir(config.Workdir, config.Git); err == git.ErrRepositoryNotExists {
				log.Debugf("could not find a valid repo at %s", config.Workdir)
			} else if err == git.NoErrAlreadyUpToDate {
				log.Debugf("repo %s %s", config.Workdir, err)
			} else if err != nil {
				log.Infof("error while pulling git repo %s: %e", config.Workdir, err)
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
