package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

const (
	LF byte = 10
)

type Config struct {
	Path         Folder `yaml:"dir"`
	URL          string `yaml:"url"`
	Remote       string `yaml:"remote"`
	Revision     string `yaml:"revision"`
	RsaPath      string `yaml:"rsaPath"`
	HttpUser     string `yaml:"httpUser"`
	HttpPassword string `yaml:"httpPassword"`
	Disable      bool   `yaml:"disable"`
}

// Initialize the git config with defaults
func (gc *Config) Initialize(workdir Folder) {
	if gc.Path == "" {
		gc.Path = workdir
	}
	if gc.Remote == "" {
		gc.Remote = "origin"
	}
	if gc.Revision == "" {
		gc.Revision = "main"
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

func (gc *Config) Checkout(revision string) error {
	if gc.Disable {
		return fmt.Errorf("git pull functionality is disabled")
	}
	log.Debug("git checkout")
	currentCommit := gc.Path.GetCommit("HEAD")
	log.Debugf("HEAD: %s", currentCommit)
	wantedCommit := gc.Path.GetCommit(revision)
	log.Debugf("Wanted: %s", wantedCommit)

	if currentCommit == wantedCommit {
		log.Debugf("Already at revision %s", revision)
		gc.Revision = revision
		return nil
	} else if wantedCommit == "" {
		return fmt.Errorf("revision (%s) is not found in this repo", revision)
	} else if err := gc.Path.RunGitCommand([]string{"checkout", revision}); err != nil {
		return err
	} else if gc.Path.GetCommit("HEAD") != wantedCommit {
		return fmt.Errorf("revision (%s => %s) not as expected (%s) after checkout", gc.Revision, wantedCommit, currentCommit)
	}
	gc.Revision = revision
	return nil
}

func (gc Config) Clean() error {
	if err := gc.Path.RunGitCommand([]string{"clean", "-f"}); err != nil {
		return err
	}
	if err := gc.Path.RunGitCommand([]string{"restore", "--staged", "."}); err != nil {
		return err
	}
	if err := gc.Path.RunGitCommand([]string{"restore", "."}); err != nil {
		return err
	}
	return nil
}

func (gc Config) Clone() error {
	if gc.Disable {
		return fmt.Errorf("git pull functionality is disabled")
	}
	if gc.Path.IsPrepared() {
		log.Debug("Repo already is cloned, pulling instead")
		return gc.Pull()
	}

	if exists, err := gc.Path.Exists(); err != nil {
		log.Debug("Repo %s is not a git repo, error getting more info", gc.Path)
		return err
	} else if !exists {
		log.Debug("Creating repo folder %s", gc.Path)
		if err = os.MkdirAll(string(gc.Path), os.ModePerm); err != nil {
			return err
		}
		// We need to create the folder!!!
	}

	log.Debug("Running git clone for %s", gc.Path)
	if err := gc.Path.RunGitCommand([]string{"clone", "-o", gc.Remote, gc.URL, string(gc.Path)}); err != nil {
		return err
	}
	return gc.Checkout(gc.Revision)
}

func (gc Config) Pull() error {
	if gc.Disable {
		return fmt.Errorf("git pull functionality is disabled")
	}
	if err := gc.Clean(); err != nil {
		return err
	}
	if err := gc.Checkout(gc.Revision); err != nil {
		return err
	}
	return gc.Path.RunGitCommand([]string{"pull"})
}
