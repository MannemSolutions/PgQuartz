package jobs

type Step struct {
	StepType     string            `yaml:"type"`
	Command      string            `yaml:"command"`
	Matrix       map[string]string `yaml:"matrix"`
	Depends      []string          `yaml:"depends"`
	parent       *Steps
	dependsSteps []Step
}

type Steps map[string]Step
