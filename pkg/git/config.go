package git

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	LF byte = 10
)

type Config struct {
	Path         string `yaml:"dir"`
	URL          string `yaml:"url"`
	Remote       string `yaml:"remote"`
	Revision     string `yaml:"revision"`
	RsaPath      string `yaml:"rsaPath"`
	HttpUser     string `yaml:"httpUser"`
	HttpPassword string `yaml:"httpPassword"`
	Disable      bool   `yaml:"disable"`
}

// Initialize the git config with defaults
func (gc *Config) Initialize(workdir string) {
	if gc.Path == "" {
		gc.Path = workdir
	}
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

func (gc Config) GetBranchName() (string, error) {
	if !gc.IsGitRepo() {
		return "", fmt.Errorf("folder %s is not a git repo", gc.Path)
	}
	var stdOut bytes.Buffer
	exCommand := exec.Command("git", "branch", "--show-current")
	exCommand.Stdout = io.MultiWriter(&stdOut)
	exCommand.Dir = gc.Path
	if err := exCommand.Run(); err != nil {
		for _, line := range strings.Split(stdOut.String(), "\n") {
			log.Info(line)
		}
		return "", err
	} else {
		return stdOut.String(), nil
	}
}

func (gc Config) RunGitCommand(command []string) error {
	exCommand := exec.Command("git", command...)
	exCommand.Dir = gc.Path
	if err := exCommand.Run(); err != nil {
		return err
	} else {
		return nil
	}
}

func (gc Config) Checkout() error {
	log.Debug("git checkout")
	if name, err := gc.GetBranchName(); err != nil {
		return err
	} else if name == gc.Revision {
		log.Debugf("Already at revision %s", name)
		return nil
	} else if err = gc.RunGitCommand([]string{"checkout", gc.Revision}); err != nil {
		return nil
	} else if name, err = gc.GetBranchName(); err != nil {
		return err
	} else if name != gc.Revision {
		return fmt.Errorf("revision (%s) not as expected (%s) after checkout", name, gc.Revision)
	}
	return nil
}

func (gc Config) Clean() error {
	if err := gc.RunGitCommand([]string{"clean"}); err != nil {
		return err
	}
	if err := gc.RunGitCommand([]string{"restore", "--staged", "."}); err != nil {
		return err
	}
	if err := gc.RunGitCommand([]string{"restore", "."}); err != nil {
		return err
	}
	return nil
}

func (gc Config) Clone() error {
	if prepared, err := gc.IsPrepared(); err != nil {
		return err
	} else if prepared {
		log.Debug("Repo already is cloned, pulling instead")
		return gc.Pull()
	}

	if gc.Exists()
	os.MkdirAll(, os.ModePerm)
	// We need to create the folder!!!

	if err := gc.RunGitCommand([]string{"clone", "-o", gc.Remote, gc.URL, gc.Path}); err != nil {
		return err
	}
	return gc.Checkout()
}

func (gc Config) Pull() error {
	if gc.Disable {
		return fmt.Errorf("git pull functionality is disabled")
	}
	if err := gc.Clean(); err != nil {
		return err
	}
	if err := gc.Checkout(); err != nil {
		return err
	}
	return gc.RunGitCommand([]string{"pull"})
}
