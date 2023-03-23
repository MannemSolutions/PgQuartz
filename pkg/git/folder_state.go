package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Folder string

type FolderState int

const (
	folderMissing    FolderState = iota
	folderEmpty      FolderState = iota
	folderUnexpected FolderState = iota
	folderInitiated  FolderState = iota
	folderUnknown    FolderState = iota
)

func (gf Folder) _RunGitCommand(command []string) (string, string, error) {
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

func (gf Folder) SubFolder(name string) (Folder, error) {
	subFolder := Folder(filepath.Join(gf.String(), name))
	if exists, err := subFolder.Exists(); err != nil {
		return "", fmt.Errorf("could not check if folder %s already exists", subFolder)
	} else if exists {
		return "", fmt.Errorf("folder %s already exists", subFolder)
	}
	if err := os.MkdirAll(subFolder.String(), 0775); err != nil {
		return "", err
	}
	return subFolder, nil
}

func (gf Folder) String() string {
	return string(gf)
}

func (gf Folder) RunGitCommand(command []string) error {
	stdout, stderr, err := gf._RunGitCommand(command)
	if err != nil {
		log.Error(stdout)
		log.Error(stderr)
	}
	return err
}

func (gf Folder) GetCommit(revision string) string {
	if !gf.IsGitRepo() {
		log.Errorf("folder %s is not a git repo", gf)
		return ""
	}
	if out, _, err := gf._RunGitCommand([]string{"rev-list", "-n", "1", revision}); err != nil {
		log.Error("Error occured while retrieving commit %s: %e", revision, err)
		return ""
	} else {
		return strings.TrimSpace(out)
	}
}

// IsGitRepo checks if the folder is already initialized as a git repo
func (gf Folder) IsGitRepo() bool {
	if out, _, err := gf._RunGitCommand([]string{"rev-parse", "--is-inside-work-tree"}); err != nil {
		return false
	} else {
		return strings.TrimSpace(out) == "true"
	}
}

// Exists checks if the folder exists (which coud then still be empty
func (gf Folder) Exists() (bool, error) {
	if _, err := os.Stat(string(gf)); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (gf Folder) IsEmpty() (bool, error) {
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

func (gf Folder) state() FolderState {
	/*
		folderMissing   GitFolderState = iota
		folderEmpty     GitFolderState = iota
		folderInitiated GitFolderState = iota
	*/
	if exists, err := gf.Exists(); err != nil {
		log.Debug("Error checking if %s Exists: %e", gf, err)
		return folderUnknown
	} else if !exists {
		return folderMissing
	}
	if gf.IsGitRepo() {
		return folderInitiated
	}
	if empty, err := gf.IsEmpty(); err != nil {
		log.Debug("Error checking if %s is empty: %e", gf, err)
		return folderUnknown
	} else if empty {
		return folderEmpty
	}
	log.Debug("folder %s exists, is no git folder and is not empty", gf)
	return folderUnexpected
}

func (gf Folder) IsPrepared() bool {
	return gf.state() == folderInitiated
}
