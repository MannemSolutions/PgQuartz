package git

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

type GitFolder string

type GitFolderState int

const (
	GitFolderMissing    GitFolderState = iota
	GitFolderEmpty      GitFolderState = iota
	GitFolderUnexpected GitFolderState = iota
	GitFolderInitiated  GitFolderState = iota
	GitFolderUnknown    GitFolderState = iota
)

func (gf GitFolder) _RunGitCommand(command []string) (string, string, error) {
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	exCommand := exec.Command("git", command...)
	exCommand.Stdout = stdOut
	exCommand.Stderr = stdErr
	exCommand.Dir = string(gf)
	log.Debugf("Running OS command %s", exCommand.String())
	err := exCommand.Run()
	return stdOut.String(), stdErr.String(), err
}

func (gf GitFolder) RunGitCommand(command []string) error {
	stdout, stderr, err := gf._RunGitCommand(command)
	if err != nil {
		log.Error(stdout)
		log.Error(stderr)
	}
	return err
}

func (gf GitFolder) GetCommit(revision string) string {
	if !gf.IsGitRepo() {
		log.Errorf("folder %s is not a git repo", gf)
		return ""
	}
	if out, _, err := gf._RunGitCommand([]string{"rev-list", "-n", "1", revision}); err != nil {
		log.Error("Error occured while retrieving commit %s: %e", revision, err)
		return ""
	} else {
		return out
	}
}

// IsGitRepo checks if the folder is already initialized as a git repo
func (gf GitFolder) IsGitRepo() bool {
	if out, _, err := gf._RunGitCommand([]string{"rev-parse", "--is-inside-work-tree"}); err != nil {
		return false
	} else {
		return strings.TrimSpace(out) == "true"
	}
}

// Exists checks if the folder exists (which coud then still be empty
func (gf GitFolder) Exists() (bool, error) {
	if _, err := os.Stat(string(gf)); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (gf GitFolder) IsEmpty() (bool, error) {
	f, err := os.Open(string(gf))
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func (gf GitFolder) State() GitFolderState {
	/*
		GitFolderMissing   GitFolderState = iota
		GitFolderEmpty     GitFolderState = iota
		GitFolderInitiated GitFolderState = iota
	*/
	if exists, err := gf.Exists(); err != nil {
		log.Debug("Error checkkng if %s Exists: %e", gf, err)
		return GitFolderUnknown
	} else if !exists {
		return GitFolderMissing
	}
	if gf.IsGitRepo() {
		return GitFolderInitiated
	}
	if empty, err := gf.IsEmpty(); err != nil {
		log.Debug("Error checkkng if %s is empty: %e", gf, err)
		return GitFolderUnknown
	} else if empty {
		return GitFolderEmpty
	}
	log.Debug("folder %s exists, is no git folder and is not empty", gf)
	return GitFolderUnexpected
}

func (gf GitFolder) IsPrepared() bool {
	return gf.State() == GitFolderInitiated

}
