package jobs

import (
	"context"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"path/filepath"
	"strings"
	"time"

	"github.com/mannemsolutions/PgQuartz/pkg/etcd"
	"gopkg.in/yaml.v2"
)

type GitConfig struct {
	Remote       string `yaml:"remote"`
	RsaPath      string `yaml:"rsaPath"`
	HttpUser     string `yaml:"httpUser"`
	HttpPassword string `yaml:"httpPassword"`
	Disable      bool   `yaml:"disable"`
}

func (gc *GitConfig) Initialize() {
	if gc.Remote == "" {
		gc.Remote = "origin"
	}
	if gc.RsaPath == "" {
		gc.RsaPath = "~/.ssh.id_rsa"
	}
	if strings.HasPrefix(gc.RsaPath, "~/") {
		if home, err := homedir.Dir(); err != nil {
			panic(fmt.Sprintf("failed to expand homedir: %e", err))
		} else {
			gc.RsaPath = filepath.Join(home, gc.RsaPath[2:])
		}
	}
}

type Config struct {
	Git            GitConfig   `yaml:"git"`
	Steps          Steps       `yaml:"steps"`
	Checks         Checks      `yaml:"checks"`
	Target         Target      `yaml:"target"`
	Conns          Connections `yaml:"connections"`
	Alert          []Alert     `yaml:"alerts"`
	Log            []Log       `yaml:"log"`
	Debug          bool        `yaml:"debug"`
	RunOnRoleError bool        `yaml:"runOnRoleError"`
	LogFile        string      `yaml:"logFile"`
	Parallel       int         `yaml:"parallel"`
	Workdir        string      `yaml:"workdir"`
	EtcdConfig     etcd.Config `yaml:"etcdConfig"`
	Timeout        string      `yaml:"timeout"`
}

func (c Config) String() string {
	if yamlConfig, err := yaml.Marshal(&c); err != nil {
		return ""
	} else {
		return string(yamlConfig)
	}
}

func (c Config) Verify() {
	var errs []error
	if c.Parallel < 0 {
		// We want a valid value for Parallel.
		// Value 0 would number of CPU's.
		// More would be static, less than 0 is invalid.
		// Not using uint, because we only loop through this and don;t want to convert to int in the loop...
		errs = append(errs, fmt.Errorf("invalid value for Parallel %d", c.Parallel))
	} else if len(c.Steps) < 1 {
		errs = append(errs, fmt.Errorf("please define at least one step"))
	} else {
		errs = append(errs, c.Steps.Verify(c.Conns)...)
	}
	for _, err := range errs {
		log.Error(err)
	}
	if len(errs) > 0 {
		log.Panicf("config issue(s) prevent me from continuing")
	}
}

func (c *Config) Initialize() {
	c.Git.Initialize()
	c.Steps.Initialize()
}

func (c Config) GetTimeoutContext(parentContext context.Context) (context.Context, context.CancelFunc) {
	if c.Timeout == "" {
		return parentContext, nil
	}
	lockDuration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		log.Fatal(err)
	}
	return context.WithTimeout(parentContext, lockDuration)
}
