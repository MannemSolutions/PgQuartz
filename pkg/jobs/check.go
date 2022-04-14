package jobs

type Check struct {
	CheckType string `yaml:"type"`
	Command   string `yaml:"command"`
}

type Checks []Check
