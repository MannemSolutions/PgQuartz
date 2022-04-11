package jobrunner

type JobAlert struct {
	JobCheckType string `yaml:"type"`
	Command      string `yaml:"command"`
}

type JobAlerts []JobAlert
