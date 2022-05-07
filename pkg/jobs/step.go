package jobs

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type stepState int

const (
	stepStateWaiting stepState = iota
	stepStateSkipped
	stepStateReady
	stepStateScheduled
	stepStateRunning
	stepStateDone
	stepStateUnknown
)

var (
	stepStateStrings = map[stepState]string{
		stepStateWaiting:   "Waiting",
		stepStateSkipped:   "Skipped",
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
			} else if subStep.state != stepStateDone && subStep.state != stepStateSkipped {
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

func (ss Steps) CheckWhen(all Handler, stepName string) (bool, error) {
	var numChecks int
	if step, exists := ss[stepName]; ! exists {
		return false, fmt.Errorf("checking a 'when' on an undefined step %s", stepName)
	} else {
		numChecks = len(step.When)
		for _, whenCheck := range step.When {
			if ! strings.Contains(whenCheck, "{{") || ! strings.Contains(whenCheck, "}}") {
				whenCheck = fmt.Sprintf("{{if %s }}True{{end}}", whenCheck)
			}
			if t, err := template.New("when").Parse(whenCheck); err !=  nil {
				return false, err
			} else {
				log.Debugf("Processing WhenCheck '%s' for step %s", whenCheck, stepName)
				var parsed bytes.Buffer
				err = t.Execute(&parsed, all)
				log.Debugf("WhenCheck '%s' returned %s for step %s", whenCheck, parsed.String(), stepName)
				if err != nil {
					return false, err
				} else if parsed.String() != "True" {
					return false, nil
				}
			}
		}
	}
	log.Debugf("All %d WhenChecks for step %s are OK", numChecks, stepName)
	return true, nil
}

type Step struct {
	Commands Commands `yaml:"commands"`
	Depends  []string `yaml:"depends,omitempty"`
	state    stepState
	When     []string `yaml:"when,omitempty"`
	stdOut   Result   `yaml:"-"`
	stdErr   Result   `yaml:"-"`
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
		When:     s.When,
	}
}

func (s Step) StdOut() Result {
	if s.stdOut == nil {
		s.stdOut = s.Commands.StdOut()
	}
	return s.stdOut
}

func (s Step) StdErr() Result {
	if s.stdErr == nil {
		s.stdErr = s.Commands.StdErr()
	}
	return s.stdErr
}

func (s Step) Rc() int {
	return s.Commands.Rc()
}