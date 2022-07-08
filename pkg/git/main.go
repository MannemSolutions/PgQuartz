package git

import (
	"github.com/go-git/go-git/v5/plumbing"
	"io/ioutil"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/martian/log"
	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
)

var (
	// ErrRepositoryNotExists is reexported, so we can handle this error differently from caller without importing go-git there
	ErrRepositoryNotExists = git.ErrRepositoryNotExists
	NoErrAlreadyUpToDate   = git.NoErrAlreadyUpToDate
)

func getGitAuth(remoteUrls []string, config jobs.GitConfig) (transport.AuthMethod, error) {
	if urls, err := newGitUrls(remoteUrls); err != nil {
		return nil, err
	} else {
		for _, url := range urls {
			if strings.HasPrefix(url.protocol, "http") {
				if config.HttpUser != "" {
					return &http.BasicAuth{
						Username: config.HttpUser,
						Password: config.HttpPassword,
					}, nil
				} else {
					return &http.BasicAuth{
						Username: url.user,
						Password: url.password,
					}, nil
				}
			} else if url.protocol == "ssh" || url.protocol == "git" {
				if sshKey, err := ioutil.ReadFile(config.RsaPath); err != nil {
					return nil, err
				} else {
					return ssh.NewPublicKeys("git", []byte(sshKey), "")
				}
			}
		}
	}
	return nil, invalidGitUrlFormat
}

func PullCurDir(workDir string, config jobs.GitConfig) (err error) {
	var repo *git.Repository
	var workTree *git.Worktree
	var remote *git.Remote
	var auth transport.AuthMethod
	var ref *plumbing.Reference
	var updated bool
	if repo, err = git.PlainOpenWithOptions(workDir, &git.PlainOpenOptions{DetectDotGit: true}); err != nil {
		return
	} else if workTree, err = repo.Worktree(); err != nil {
		return
	} else if remote, err = repo.Remote(config.Remote); err != nil {
		return
	} else if auth, err = getGitAuth(remote.Config().URLs, config); err != nil {
		return
	} else if err = workTree.Pull(&git.PullOptions{RemoteName: config.Remote, Auth: auth}); err == git.NoErrAlreadyUpToDate {
		// no updates, which is fine
	} else if err != nil {
		return
	} else {
		updated = true
	}
	ref, err = repo.Head()
	if err != nil {
		return
	}
	log.Infof("running with branch %s (commit %s)", ref.Name(), ref.Hash())
	if !updated {
		return NoErrAlreadyUpToDate
	}
	return
}
