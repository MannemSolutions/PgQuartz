package jobs

import (
	"fmt"
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
	Parallel int                `yaml:"parallel"`
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
		// We want a valid value vor Parallel.
		// Value 0 would number of CPU's.
		// More would be static, less than 0 is invalid.
		// Not using uint, because we only loop through this and don;t want to convert to int in the loop...
		errs = append(errs, fmt.Errorf("invalid value for Parallel %d", c.Parallel))
	} else if len(c.Steps) < 1 {
		errs = append(errs, fmt.Errorf("please define at least one step"))
	} else {
		for stepName, step := range c.Steps {
			if step.Connection == "" {
				if len(c.Conns) == 1 {
					// This is fine. When only one, we use that.
				} else {
					errs = append(errs, fmt.Errorf(
						"please refernce a specific connection for step %s,  or just define only one", stepName))
				}
			} else if _, exists := c.Conns[step.Connection]; ! exists {
				errs = append(errs, fmt.Errorf("step %s references an unknown connection %s", stepName,
					step.Connection))
			}
			for _, dependency := range step.Depends {
				if _, exists := c.Steps[dependency]; ! exists {
					errs = append(errs, fmt.Errorf("step %s depends on unknown step %s", stepName, dependency))
				}
			}
		}
	}
	for _, err := range errs {
		log.Error(err)
	}
	if len(errs) > 0 {
		log.Panicf("config issue(s) prevent me from continuing")
	}
}