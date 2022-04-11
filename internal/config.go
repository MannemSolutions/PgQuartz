package internal

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mannemsolutions/PgQuartz/pkg/jobrunner"
	"github.com/mannemsolutions/PgQuartz/pkg/pg"
	"gopkg.in/yaml.v2"
)

/*
 * This module reads the config file and returns a config object with all entries from the config yaml file.
 */

const (
	envConfName     = "PGQUARTZ_CONFIG"
	defaultConfFile = "/etc/pgquartz/config.yaml"
)

type QuartzConfig struct {
	Steps    jobrunner.JobSteps   `yaml:"steps"`
	Checks   []jobrunner.JobCheck `yaml:"checks"`
	Conns    map[string]pg.Conn   `yaml:"connections"`
	Alert    []jobrunner.JobAlert `yaml:"alerts"`
	Log      []jobrunner.JobLog   `yaml:"log"`
	Debug    bool
	Parallel uint `yaml:"parallel"`
}

func NewConfig() (config QuartzConfig, err error) {
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

	if debug {
		config.Debug = true
	}

	return config, err
}

func (c QuartzConfig) String() string {
	if yamlConfig, err := yaml.Marshal(&c); err != nil {
		return ""
	} else {
		return string(yamlConfig)
	}
}
