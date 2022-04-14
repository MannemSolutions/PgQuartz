package jobs

type Alert struct {
	AlertType string `yaml:"type"`
	Command   string `yaml:"command"`
}

type Alerts []Alert
