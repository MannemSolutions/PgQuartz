package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/martian/log"
)

var (
	// ErrRepositoryNotExists is reexported, so we can handle this error differently from caller without importing go-git there
	ErrRepositoryNotExists = git.ErrRepositoryNotExists
	NoErrAlreadyUpToDate   = git.NoErrAlreadyUpToDate
)

type Config struct {
	Remote       string `yaml:"remote"`
	RsaPath      string `yaml:"rsaPath"`
	HttpUser     string `yaml:"httpUser"`
	HttpPassword string `yaml:"httpPassword"`
	Disable      bool   `yaml:"disable"`
}

func (gc *Config) Initialize() {
	if gc.Remote == "" {
		gc.Remote = "origin"
	}
	if gc.RsaPath == "" {
		gc.RsaPath = "~/.ssh/id_rsa"
	}
	if strings.HasPrefix(gc.RsaPath, "~/") {
		if home, err := homedir.Dir(); err != nil {
			panic(fmt.Sprintf("failed to expand homedir: %e", err))
		} else {
			gc.RsaPath = filepath.Join(home, gc.RsaPath[2:])
		}
	}
}

func (gc Config) getGitAuth(remoteUrls []string) (transport.AuthMethod, error) {
	if urls, err := newGitUrls(remoteUrls); err != nil {
		return nil, err
	} else {
		for _, url := range urls {
			if strings.HasPrefix(url.protocol, "http") {
				if gc.HttpUser != "" {
					return &http.BasicAuth{
						Username: gc.HttpUser,
						Password: gc.HttpPassword,
					}, nil
				} else {
					return &http.BasicAuth{
						Username: url.user,
						Password: url.password,
					}, nil
				}
			} else if url.protocol == "ssh" || url.protocol == "git" {
				if sshKey, err := os.ReadFile(gc.RsaPath); err != nil {
					return nil, err
				} else {
					return ssh.NewPublicKeys("git", []byte(sshKey), "")
				}
			}
		}
	}
	return nil, invalidGitUrlFormat
}

func (gc Config) PullCurDir(workDir string) (err error) {
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
	} else if remote, err = repo.Remote(gc.Remote); err != nil {
		return
	} else if auth, err = gc.getGitAuth(remote.Config().URLs); err != nil {
		return
	} else if err = workTree.Pull(&git.PullOptions{RemoteName: gc.Remote, Auth: auth}); err == git.NoErrAlreadyUpToDate {
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
