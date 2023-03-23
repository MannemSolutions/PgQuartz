package git

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strings"
	"testing"
)

func TestInitiated(t *testing.T) {
	assert.True(t, strings.HasSuffix(InitiatedFolder.String(), "/initiated"), "folder should end with /initiated")
	assert.True(t, InitiatedFolder.IsPrepared(), "InitedFolder should have Prepared state")
	empty, err := InitiatedFolder.IsEmpty()
	assert.NoError(t, err, "Checking if folder is empty should succeed")
	assert.False(t, empty, fmt.Sprintf("Folder %s should not be empty", InitiatedFolder))
	exists, err := InitiatedFolder.Exists()
	assert.NoError(t, err, "Checking if folder exists should succeed")
	assert.True(t, exists, "Folder exists should exist")
	assert.True(t, InitiatedFolder.IsGitRepo(), "InitedFolder should be detected as git repo")
	assert.Equal(t, folderInitiated, InitiatedFolder.state(), "state should be folderInitiated")
	commit := InitiatedFolder.GetCommit("tag1")
	assert.Regexp(t, regexp.MustCompile("[a-f\\d]{40}"), commit,
		fmt.Sprintf("GetCommit should be returning a commit for tag1, but retruned %s", commit))
	assert.NoError(t, InitiatedFolder.RunGitCommand([]string{"status"}), "")
	stdout, stderr, err := InitiatedFolder._RunGitCommand([]string{"branch", "--show-current"})
	assert.Equal(t, "main\n", stdout, "_RunGitCommand should return 'main' on stdout")
	assert.Equal(t, "", stderr, "_RunGitCommand should return '' on stderr")
	assert.NoError(t, err, "_RunGitCommand should return no error")
}

func TestMissing(t *testing.T) {
	assert.True(t, strings.HasSuffix(MissingFolder.String(), "/missing"), "folder should end with /missing")
	assert.False(t, MissingFolder.IsPrepared(), fmt.Sprintf("Folder %s should not have Prepared state",
		MissingFolder))
	empty, err := MissingFolder.IsEmpty()
	assert.Error(t, err, "Checking if folder is empty should not succeed")
	assert.False(t, empty, fmt.Sprintf("Folder %s should not be empty", MissingFolder))
	exists, err := MissingFolder.Exists()
	assert.NoError(t, err, "Checking if folder exists should succeed")
	assert.False(t, exists, "Folder exists should exist")
	assert.False(t, MissingFolder.IsGitRepo(), "InitedFolder should be detected as git repo")
	assert.Equal(t, folderMissing, MissingFolder.state(), "state should be folderMissing")
	commit := MissingFolder.GetCommit("tag1")
	assert.Equal(t, "", commit,
		fmt.Sprintf("GetCommit should be returning eptystring as commit for tag1, but retruned %s", commit))
	assert.Error(t, MissingFolder.RunGitCommand([]string{"status"}), "")
	stdout, stderr, err := MissingFolder._RunGitCommand([]string{"branch", "--show-current"})
	assert.Equal(t, "", stdout, "_RunGitCommand should return 'main' on stdout")
	assert.Equal(t, "", stderr, "_RunGitCommand should return '' on stderr")
	assert.Error(t, err, "_RunGitCommand should return no error")
}

func TestEmpty(t *testing.T) {
	assert.True(t, strings.HasSuffix(EmptyFolder.String(), "/empty"), "folder should end with /empty")
	assert.False(t, EmptyFolder.IsPrepared(), "InitedFolder should not have Prepared state")
	empty, err := EmptyFolder.IsEmpty()
	assert.NoError(t, err, "Checking if folder is empty should succeed")
	assert.True(t, empty, fmt.Sprintf("Folder %s should be empty", EmptyFolder))
	exists, err := EmptyFolder.Exists()
	assert.NoError(t, err, "Checking if folder exists should succeed")
	assert.True(t, exists, "Folder should exist")
	assert.False(t, EmptyFolder.IsGitRepo(), "InitedFolder should be detected as git repo")
	assert.Equal(t, folderEmpty, EmptyFolder.state(), "state should be folderEmpty")
	commit := EmptyFolder.GetCommit("tag1")
	assert.Equal(t, "", commit,
		fmt.Sprintf("GetCommit should be returning an emptystring for tag1, but returned %s", commit))
	assert.Error(t, EmptyFolder.RunGitCommand([]string{"status"}), "")
	stdout, stderr, err := EmptyFolder._RunGitCommand([]string{"branch", "--show-current"})
	assert.Equal(t, "", stdout, "_RunGitCommand should return '' on stdout")
	assert.Equal(t, "fatal: not a git repository (or any of the parent directories): .git\n",
		stderr, "_RunGitCommand should return errormsg on stderr")
	assert.Error(t, err, "_RunGitCommand should return no error")
}

func TestUnexpected(t *testing.T) {
	//Unexpected state means that folder has files, but is not a repo
	assert.True(t, strings.HasSuffix(UnexpectedFolder.String(), "/unexpected"), "folder should end with /unexpected")
	assert.False(t, UnexpectedFolder.IsPrepared(), "InitedFolder should not have Prepared state")
	empty, err := UnexpectedFolder.IsEmpty()
	assert.NoError(t, err, "Checking if folder is empty should succeed")
	assert.False(t, empty, fmt.Sprintf("Folder %s should be empty", UnexpectedFolder))
	exists, err := UnexpectedFolder.Exists()
	assert.NoError(t, err, "Checking if folder exists should succeed")
	assert.True(t, exists, "Folder exists should exist")
	assert.False(t, UnexpectedFolder.IsGitRepo(), "InitedFolder should not be detected as git repo")
	assert.Equal(t, folderUnexpected, UnexpectedFolder.state(), "state should be folderUnexpected")
	commit := UnexpectedFolder.GetCommit("tag1")
	assert.Equal(t, "", commit,
		fmt.Sprintf("GetCommit should be returning an emptystring for tag1, but retruned %s", commit))
	assert.Error(t, UnexpectedFolder.RunGitCommand([]string{"status"}), "")
	stdout, stderr, err := UnexpectedFolder._RunGitCommand([]string{"branch", "--show-current"})
	assert.Equal(t, "", stdout, "_RunGitCommand should return 'main' on stdout")
	assert.Equal(t, "fatal: not a git repository (or any of the parent directories): .git\n",
		stderr, "_RunGitCommand should return '' on stderr")
	assert.Error(t, err, "_RunGitCommand should return no error")
}
