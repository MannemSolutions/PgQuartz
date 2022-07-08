package git

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHttps(t *testing.T) {
	url, err := newGitUrl("https://me:secret@host.local:8443/folder/to/repo.git")
	assert.Nil(t, err, "Convert https url without error")
	assert.Equal(t, "https", url.protocol, "Detect protocol from https url")
	assert.Equal(t, "me", url.user, "Detect user from https url")
	assert.Equal(t, "secret", url.password, "Detect password from https url")
	assert.Equal(t, "host.local", url.host, "Detect password from https url")
	assert.Equal(t, 8443, url.port, "Detect password from https url")
	assert.Equal(t, "folder/to/repo.git", url.repo, "Detect repo from https url")

	url, err = newGitUrl("http://me@host.local/folder/to/repo.git")
	assert.Nil(t, err, "Convert http url without error")
	assert.Equal(t, "http", url.protocol, "Detect protocol from http url without port")
	assert.Equal(t, 80, url.port, "Detect protocol from http url without port")
	assert.Equal(t, "me", url.user, "Detect user without password from http url")
	assert.Equal(t, "", url.password, "Detect http url without password")

	url, err = newGitUrl("https://host.local/folder/to/repo.git")
	assert.Nil(t, err, "Convert https url without error")
	assert.Equal(t, "https", url.protocol, "Detect protocol from https url without user")
	assert.Equal(t, 443, url.port, "Detect port from https url without port setting")
	assert.Equal(t, "", url.user, "Detect https url without user")
	assert.Equal(t, "", url.password, "Detect password from https url without user")

	_, err = newGitUrl("https://me@host.local:port/folder/to/repo.git")
	assert.EqualError(t, err, "strconv.Atoi: parsing \"port\": invalid syntax")
}

func TestSsh(t *testing.T) {
	url, err := newGitUrl("ssh://myself@example.local:1234/folder/to/repo.git")
	assert.Nil(t, err, "Convert ssh url without error")
	assert.Equal(t, "ssh", url.protocol, "Detect protocol from ssh url")
	assert.Equal(t, "myself", url.user, "Detect user from ssh url")
	assert.Equal(t, "", url.password, "Cannot get password from ssh url")
	assert.Equal(t, "example.local", url.host, "Cannot detect password from ssh url")
	assert.Equal(t, 1234, url.port, "Detect port from ssh url")
	assert.Equal(t, "folder/to/repo.git", url.repo, "Detect repo from ssh url")

	url, err = newGitUrl("ssh://host/folder/to/repo.git")
	assert.Nil(t, err, "Convert ssh url without error")
	assert.Equal(t, "ssh", url.protocol, "Detect protocol from ssh url")
	assert.Equal(t, 22, url.port, "Detect port from ssh url without port setting")
	assert.Equal(t, "", url.user, "Convert ssh url without user")
}

func TestGit(t *testing.T) {
	url, err := newGitUrl("git@host/folder/to/repo.git")
	assert.Nil(t, err, "Convert git url without error")
	assert.Equal(t, "git", url.protocol, "Detect protocol from git url")
	assert.Equal(t, "", url.user, "Detect user from git url")
	assert.Equal(t, 22, url.port, "Detect port from git url")
	assert.Equal(t, "folder/to/repo.git", url.repo, "Detect repo from git url")

	url, err = newGitUrl("git@host/~onlyme/folder/to/repo.git")
	assert.Nil(t, err, "Convert git url with user without error")
	assert.Equal(t, "git", url.protocol, "Detect protocol from git url")
	assert.Equal(t, "onlyme", url.user, "Convert git url without user")
}
