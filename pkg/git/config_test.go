package git

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

const (
	StartRevision = "v0.5.1"
	NextRevision  = "v0.5"
	FinalRevision = "main"
)

func TestConfig_Initialize(t *testing.T) {
	c := Config{}
	c.Initialize("default")
	assert.Equal(t, "default", c.Workdir, "Detect workdir defaults to init argument during Initialize")
	c.Initialize("second_default")

	assert.NotEqual(t, "second_default", c.Workdir, "Detect initialize does not overwrite workdir if set")

	assert.Equal(t, "origin", c.Remote, "Detect remote defaults to 'origin' after Initialize")
	assert.Equal(t, "main", c.Revision, "Detect revision defaults to 'main' after Initialize")
	if home, err := homedir.Dir(); err != nil {
		panic(fmt.Sprintf("failed to expand homedir: %e", err))
	} else {
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s/.ssh/id_rsa", home)), c.RsaPath,
			"Detect rsa path defaults to id_rsa in homedir after Initialize")
	}
}

func TestPullCurDir(t *testing.T) {
	workdir, err := os.MkdirTemp("", "go_test_pgquartz")
	assert.NoError(t, err, "could not create a temp dir")
	defer os.RemoveAll(workdir)

	c := Config{
		Workdir:  workdir,
		Remote:   "https://github.com/MannemSolutions/PgQuartz.git",
		Revision: StartRevision,
		Disable:  true,
	}
	c.Initialize(workdir)
	assert.False(t, c.IsGitRepo(), "Check if IsGitRepo works as expected (1)")
	assert.Error(t, c.Pull(), "PullCurDir should return an error when Config has Disabled set to true")
	c.Disable = false
	assert.Nil(t, c.Pull(), "PullCurDir should be able to pull")
	assert.DirExists(t, filepath.Join(workdir, ".git"), "After PullCurDir workdir should be a git repo (hold a .git folder)")
	assert.True(t, c.IsGitRepo(), "Check if IsGitRepo works as expected (2)")

	assert.Equal(t, "origin", c.Remote, "Detect remote defaults to 'origin' after Initialize")
	if home, err := homedir.Dir(); err != nil {
		panic(fmt.Sprintf("failed to expand homedir: %e", err))
	} else {
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s/.ssh/id_rsa", home)), c.RsaPath,
			"Detect rsa path defaults to id_rsa in homedir after Initialize")
	}
}

func TestSubDir(t *testing.T) {
	workdir, err := os.MkdirTemp("", "go_test_pgquartz")
	assert.NoError(t, err, "could not create the temp dir for SubDir test")
	defer os.RemoveAll(workdir)

	c := Config{
		Workdir:  workdir,
		Remote:   "https://github.com/MannemSolutions/PgQuartz.git",
		Revision: FinalRevision,
	}
	c.Initialize(workdir)
	assert.Nil(t, c.Pull(), "PullCurDir should be able to pull")
	c.Workdir = filepath.Join(c.Workdir, "cmd")
	c.Initialize(workdir)
	assert.True(t, c.IsGitRepo(), "Check if IsGitRepo works as expected (3)")
}
