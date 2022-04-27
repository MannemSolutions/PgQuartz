package jobs

import (
	"fmt"
)

type stepState int

const (
	stepStateWaiting stepState = iota
	stepStateReady
	stepStateScheduled
	stepStateRunning
	stepStateDone
	stepStateUnknown
)

var (
	stepStateStrings = map[stepState]string{
		stepStateWaiting:   "Waiting",
		stepStateReady:     "Ready",
		stepStateScheduled: "Scheduled",
		stepStateRunning:   "Running",
		stepStateDone:      "Done",
		stepStateUnknown:   "Unknown",
	}
)

func (sst stepState) String() string {
	stepStateString, exists := stepStateStrings[sst]
	if !exists {
		log.Panicf("Unknown stepStateStrings")
	}
	return stepStateString
}

type Steps map[string]*Step

func (ss Steps) Verify(conns Connections) (errs []error) {
	for stepName, step := range ss {
		errs = append(errs, step.Commands.Verify(stepName, conns)...)
		for _, dependency := range step.Depends {
			if _, exists := ss[dependency]; !exists {
				errs = append(errs, fmt.Errorf("step %s depends on unknown step %s", stepName, dependency))
			}
		}
	}
	return errs
}

func (ss Steps) Clone() Steps {
	clone := make(Steps)
	for name, step := range ss {
		clone[name] = step.Clone()
	}
	return clone
}

func (ss Steps) stepState(stepName string) stepState {
	if step, exists := ss[stepName]; !exists {
		log.Panicf("Looking for a step %s that does not exist???", stepName)
	} else {
		return step.state
	}
	return stepStateUnknown
}

func (ss Steps) setStepState(stepName string, newState stepState) {
	if step, exists := ss[stepName]; !exists {
		log.Panicf("Looking for a step %s that does not exist???", stepName)
	} else if err := step.setState(newState); err != nil {
		log.Panicf("Error while changing state for step %s: %e", stepName, err)
	} else {
		step.state = newState
	}
}

func (ss Steps) GetReadySteps() (ready []string) {
	var isReady bool
	for stepName, step := range ss {
		if step.state != stepStateWaiting {
			continue
		}
		isReady = true
		for _, dependency := range step.Depends {
			if subStep, exists := ss[dependency]; !exists {
				log.Panicf("step %s depends on unknows step %s", stepName, dependency)
			} else if subStep.state != stepStateDone {
				isReady = false
				break
			}
		}
		if isReady {
			ready = append(ready, stepName)
		}
	}
	return ready
}

func (ss Steps) NumWaiting() (numWaiting int) {
	for _, step := range ss {
		if step.state != stepStateWaiting {
			continue
		}
		numWaiting += 1
	}
	return numWaiting
}

type Step struct {
	Commands Commands `yaml:"commands"`
	Depends  []string `yaml:"depends"`
	state    stepState
}

func (s *Step) setState(newState stepState) error {
	if s.state > newState {
		return fmt.Errorf("invalid step transition from %s to %s", s.state.String(), newState.String())
	} else {
		s.state = newState
		return nil
	}
}

func (s Step) Clone() *Step {
	return &Step{
		Commands: s.Commands.Clone(),
		Depends:  s.Depends,
		state:    stepStateWaiting,
	}
}
