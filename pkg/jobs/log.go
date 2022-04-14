package jobs

type Log struct {
	LogType string `yaml:"type"`
	Command string `yaml:"command"`
}

type Logs []Log
