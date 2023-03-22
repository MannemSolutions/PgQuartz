package git

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

const (
	StartRevision   = "v0.5.1"
	InvalidRevision = "invalid_revision"
	EarlierRevision = "v0.5"
	LaterRevision   = "main"
	URL             = "https://github.com/MannemSolutions/PgQuartz.git"
	Remote          = "pgquartz"
)

func TestMain(m *testing.M) {
	exitcode := m.Run()
	log.Sync()
	os.Exit(exitcode)
}

func TestConfig_Initialize(t *testing.T) {
	log = zaptest.NewLogger(t).Sugar()
	c := Config{}
	c.Initialize("default")
	assert.Equal(t, "default", string(c.Path), "Detect workdir defaults to init argument during Initialize")
	c.Initialize("second_default")

	assert.NotEqual(t, "second_default", c.Path, "Detect initialize does not overwrite workdir if set")

	assert.Equal(t, "origin", c.Remote, "Detect remote defaults to 'origin' after Initialize")
	assert.Equal(t, "main", c.Revision, "Detect revision defaults to 'main' after Initialize")
	if home, err := homedir.Dir(); err != nil {
		panic(fmt.Sprintf("failed to expand homedir: %e", err))
	} else {
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s/.ssh/id_rsa", home)), c.RsaPath,
			"Detect rsa path defaults to id_rsa in homedir after Initialize")
	}
}

func TestCloneCurDir(t *testing.T) {
	log = zaptest.NewLogger(t).Sugar()
	workdir, err := os.MkdirTemp("", "go_test_pgquartz")
	assert.NoError(t, err, "could not create a temp dir")
	defer os.RemoveAll(workdir)

	c := Config{
		Path:     GitFolder(workdir),
		URL:      URL,
		Remote:   Remote,
		Revision: StartRevision,
		Disable:  true,
	}
	c.Initialize(workdir)
	assert.False(t, c.Path.IsGitRepo(), "Check if IsGitRepo works as expected (1)")
	assert.Error(t, c.Clone(), ".Clone should return an error when Config has Disabled set to true")
	c.Disable = false
	assert.Nil(t, c.Clone(), ".Clone() should be able to pull")
	assert.DirExists(t, filepath.Join(workdir, ".git"), "After .Clone() workdir should be a git repo (hold a .git folder)")
	assert.True(t, c.Path.IsGitRepo(), "Check if IsGitRepo works as expected (2)")

	assert.Equal(t, Remote, c.Remote, "Detect remote defaults to 'origin' after Initialize")
	if home, err := homedir.Dir(); err != nil {
		panic(fmt.Sprintf("failed to expand homedir: %e", err))
	} else {
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s/.ssh/id_rsa", home)), c.RsaPath,
			"Detect rsa path defaults to id_rsa in homedir after Initialize")
	}
	assert.Nil(t, c.Checkout(EarlierRevision), ".Checkout should work for an earlier Revision")
	assert.Nil(t, c.Checkout(LaterRevision), ".Checkout should work for a later Revision")
}

func TestSubDir(t *testing.T) {
	log = zaptest.NewLogger(t).Sugar()
	workdir, err := os.MkdirTemp("", "go_test_pgquartz")
	assert.NoError(t, err, "could not create the temp dir for SubDir test")
	//defer os.RemoveAll(workdir)

	c := Config{
		Path:     GitFolder(workdir),
		URL:      URL,
		Remote:   Remote,
		Revision: LaterRevision,
	}
	c.Initialize(workdir)
	log.Debug(c)
	assert.Nil(t, c.Clone(), ".Clone() should be able to clone")
	c.Path = GitFolder(filepath.Join(string(c.Path), "cmd"))
	c.Initialize(workdir)
	assert.True(t, c.Path.IsGitRepo(), "Check if IsGitRepo works as expected (3)")
	assert.Nil(t, c.Checkout(EarlierRevision), ".Checkout should work if this is ")
}
