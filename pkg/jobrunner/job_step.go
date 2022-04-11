package jobrunner

type JobStep struct {
	JobStepType  string            `yaml:"type"`
	Command      string            `yaml:"command"`
	Matrix       map[string]string `yaml:"matrix"`
	Depends      []string          `yaml:"depends"`
	parent       *JobSteps
	dependsSteps []JobStep
}

type JobSteps map[string]JobStep
