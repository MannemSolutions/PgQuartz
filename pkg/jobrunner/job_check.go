package jobrunner

type JobCheck struct {
	JobCheckType string `yaml:"type"`
	Command      string `yaml:"command"`
}

type JobChecks []JobChecks
