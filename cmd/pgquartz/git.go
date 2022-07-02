package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"io/ioutil"
	"strings"
)

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
	var repo *git.Repository
	var workTree *git.Worktree
	var remote *git.Remote
	var auth transport.AuthMethod
	var err error
	if repo, err = git.PlainOpenWithOptions(workDir, &git.PlainOpenOptions{DetectDotGit: true}); err != nil {
		return err
	} else if workTree, err = repo.Worktree(); err != nil {
		return err
	} else if remote, err = repo.Remote(config.Remote); err != nil {
		return err
	} else if auth, err = getSshAuth(remote.String(), config); err != nil {
		return err
	} else if err = workTree.Pull(&git.PullOptions{RemoteName: config.Remote, Auth: auth}); err == git.NoErrAlreadyUpToDate {
		// no updates, which is fine
	} else if err != nil {
		return err
	}
	if ref, err := repo.Head(); err != nil {
		return err
	} else {
		log.Infof("running with branch %s (commit %s)", ref.Name(), ref.Hash())
	}
	return nil
}
