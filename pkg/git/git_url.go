package git

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reGitHttpUrl        = regexp.MustCompile(`(https?)://(\S*?@)?(\S*?)/(\S*.git)`)
	reGitSshUrl         = regexp.MustCompile(`(ssh)://(\S*?@)?(\S*?)(/\S*.git)`)
	reGitUrl            = regexp.MustCompile(`(git)@(\S*?)(/~\S*?)?(/\S*.git)`)
	invalidGitUrlFormat = fmt.Errorf("could not parse url as this format")
)

type gitUrl struct {
	url      string
	protocol string
	user     string
	password string
	host     string
	port     int
	repo     string
}

func newGitUrlFromHttp(remoteUrl string) (url gitUrl, err error) {
	var password string
	var port int
	urlParts := reGitHttpUrl.FindStringSubmatch(remoteUrl)
	if len(urlParts) < 5 {
		return gitUrl{}, invalidGitUrlFormat
	}
	userPassword := strings.SplitN(strings.TrimSuffix(urlParts[2], "@"), ":", 2)
	user := userPassword[0]
	if len(userPassword) > 1 {
		password = userPassword[1]
	} else {
		password = ""
	}
	hostPort := strings.SplitN(urlParts[3], ":", 2)
	host := hostPort[0]
	if len(hostPort) > 1 {
		port, err = strconv.Atoi(hostPort[1])
		if err != nil {
			return gitUrl{}, err
		}
	} else if urlParts[1] == "https" {
		port = 443
	} else {
		port = 80
	}
	return gitUrl{
		url:      urlParts[0],
		protocol: urlParts[1],
		user:     user,
		password: password,
		host:     host,
		port:     port,
		repo:     urlParts[4],
	}, nil
}

func newGitUrlFromSsh(remoteUrl string) (url gitUrl, err error) {
	var port int
	urlParts := reGitSshUrl.FindStringSubmatch(remoteUrl)
	if len(urlParts) < 4 {
		return gitUrl{}, invalidGitUrlFormat
	}
	hostPort := strings.SplitN(urlParts[3], ":", 2)
	host := hostPort[0]
	if len(hostPort) > 1 {
		port, err = strconv.Atoi(hostPort[1])
		if err != nil {
			return gitUrl{}, err
		}
	} else {
		port = 22
	}
	return gitUrl{
		url:      urlParts[0],
		protocol: urlParts[1],
		user:     strings.TrimSuffix(urlParts[2], "@"),
		host:     host,
		port:     port,
		repo:     strings.TrimPrefix(urlParts[4], "/"),
	}, nil
}

func newGitUrlFromGit(remoteUrl string) (url gitUrl, err error) {
	urlParts := reGitUrl.FindStringSubmatch(remoteUrl)
	if len(urlParts) < 4 {
		return gitUrl{}, invalidGitUrlFormat
	}
	return gitUrl{
		url:      urlParts[0],
		protocol: urlParts[1],
		host:     urlParts[2],
		user:     strings.TrimPrefix(urlParts[3], "/~"),
		port:     22,
		repo:     strings.TrimPrefix(urlParts[4], "/"),
	}, nil
}

func newGitUrl(remoteUrl string) (url gitUrl, err error) {
	if url, err = newGitUrlFromHttp(remoteUrl); err != invalidGitUrlFormat {
		return
	} else if url, err = newGitUrlFromSsh(remoteUrl); err != invalidGitUrlFormat {
		return
	} else if url, err = newGitUrlFromGit(remoteUrl); err != invalidGitUrlFormat {
		return
	}
	return gitUrl{}, invalidGitUrlFormat
}

type gitUrls []gitUrl

func newGitUrls(remoteUrls []string) (urls gitUrls, err error) {
	for _, remoteUrl := range remoteUrls {
		if url, err := newGitUrl(remoteUrl); err != nil {
			return gitUrls{}, err
		} else {
			urls = append(urls, url)
		}
	}
	return urls, nil
}
