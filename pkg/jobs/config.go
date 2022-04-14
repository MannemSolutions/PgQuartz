package jobs

import (
	"github.com/mannemsolutions/PgQuartz/pkg/pg"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Steps    Steps              `yaml:"steps"`
	Checks   []Check            `yaml:"checks"`
	Target   Target             `yaml:"target"`
	Conns    map[string]pg.Conn `yaml:"connections"`
	Alert    []Alert            `yaml:"alerts"`
	Log      []Log              `yaml:"log"`
	Debug    bool               `yaml:"debug"`
	Parallel uint               `yaml:"parallel"`
}

func (c Config) String() string {
	if yamlConfig, err := yaml.Marshal(&c); err != nil {
		return ""
	} else {
		return string(yamlConfig)
	}
}
