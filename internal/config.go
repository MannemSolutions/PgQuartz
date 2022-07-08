package internal

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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

var (
	debug      bool
	version    bool
	configFile string
)

func ProcessFlags() (err error) {
	if configFile != "" {
		return
	}

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
	return err
}

func NewConfig() (config jobs.Config, err error) {
	if err = ProcessFlags(); err != nil {
		return
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
	jobName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if config.Workdir == "" {
		config.Workdir = dir
	}

	if config.LogFile == "" {
		// If it is emptystring, then don't do fancy stuff with stat on it
	} else if fileInfo, err := os.Stat(config.LogFile); err != nil {
		return config, err
	} else if fileInfo.IsDir() {
		// is a directory
		t := time.Now()
		logFileName := fmt.Sprintf("%s_%s.log", t.Format("2006-01-02"), jobName)
		config.LogFile = filepath.Join(config.LogFile, logFileName)
	}
	if config.EtcdConfig.LockKey == "" {
		config.EtcdConfig.LockKey = jobName
	}

	if debug {
		config.Debug = true
	}

	return config, err
}
