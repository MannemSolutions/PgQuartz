package internal

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mannemsolutions/PgQuartz/pkg/jobs"
	"gopkg.in/yaml.v2"
)

/*
 * This module reads the config file and returns a config object with all entries from the config yaml file.
 */

const (
	envConfName     = "PGQUARTZ_CONFIG"
	defaultConfFile = "/etc/pgquartz/config.yaml"
)

func NewConfig() (config jobs.Config, err error) {
	var debug bool

	var version bool

	var configFile string

	flag.BoolVar(&debug, "d", false, "Add debugging output")
	flag.BoolVar(&version, "v", false, "Show version information")

	flag.StringVar(&configFile, "c", os.Getenv(envConfName), "Path to configfile")

	flag.Parse()

	if version {
		//nolint
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if configFile == "" {
		configFile = defaultConfFile
	}

	configFile, err = filepath.EvalSymlinks(configFile)
	if err != nil {
		return config, err
	}

	// This only parsed as yaml, nothing else
	// #nosec
	yamlConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlConfig, &config)
	config.Initialize()
	dir, fileName := path.Split(configFile)
	if config.Workdir == "" {
		config.Workdir = dir
	}
	if config.EtcdConfig.LockKey == "" {
		config.EtcdConfig.LockKey = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	if debug {
		config.Debug = true
	}

	return config, err
}
