package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type GitFolderState int

const (
	GitFolderMissing   GitFolderState = iota
	GitFolderEmpty     GitFolderState = iota
	GitFolderInitiated GitFolderState = iota
)

// IsGitRepo checks if the folder is already initialized as a git repo
func (gc Config) IsGitRepo() bool {
	if fileInfo, err := os.Stat(filepath.Join(gc.Path, ".git")); err != nil {
		log.Debugf("Error checking if it is a git repo: %e", err)
		return false
	} else {
		return fileInfo.IsDir()
	}
}

// Exists checks if the folder exists (which coud then still be empty
func (gc Config) Exists() (bool, error) {
	if _, err := os.Stat(gc.Path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (gc Config) IsEmpty() (bool, error) {
	f, err := os.Open(gc.Path)
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

func (gc Config) IsPrepared() (bool, error) {
	if exists, err := gc.Exists(); err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}
	if gc.IsGitRepo() {
		return true, nil
	}
	if empty, err := gc.IsEmpty(); err != nil {
		return false, err
	} else if !empty {
		return false, fmt.Errorf("folder exists, is no git folder and is not empty")
	}
	return false, nil
}

