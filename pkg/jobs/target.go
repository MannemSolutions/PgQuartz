package jobs

type Target struct {
	Role         string `yaml:"role"`
	Distribution string `yaml:"distribution"`
	Repeat       int    `yaml:"repeat"`
	Delay        int    `yaml:"delay"`
}
