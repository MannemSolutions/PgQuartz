package jobs

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Steps    Steps       `yaml:"steps"`
	Checks   Commands    `yaml:"checks"`
	Target   Target      `yaml:"target"`
	Conns    Connections `yaml:"connections"`
	Alert    []Alert     `yaml:"alerts"`
	Log      []Log       `yaml:"log"`
	Debug    bool        `yaml:"debug"`
	Parallel int         `yaml:"parallel"`
	Workdir  string      `yaml:"workdir"`
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
