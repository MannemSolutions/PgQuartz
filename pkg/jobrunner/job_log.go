package jobrunner

type JobLog struct {
	JobCheckType string `yaml:"type"`
	Command      string `yaml:"command"`
}

type JobLogs []JobLog
