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

func (ss *Steps) InstanceFinished(w Work) {
	if s, sExists := (*ss)[w.Step]; !sExists {
		log.Panicf("Unknown step %s is finished", w.Step)
	} else if i, iExists := s.Instances[w.ArgKey]; !iExists {
		log.Panicf("Unknown instance %s for step %s is finished", w.ArgKey, w.Step)
	} else {
		i.done = true
		log.Debugf("Set instance [%s].[%s] to done", w.Step, w.ArgKey)
	}
}

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

func (ss *Steps) Initialize() {
	for _, step := range *ss {
		step.Initialize()
	}
}

func (ss Steps) GetNumInstances() int {
	var count int
	for name, step := range ss {
		step.SetInstances()
		num := len(step.GetInstances())
		log.Debugf("step %s has %d instances", name, num)
		count += num
	}
	log.Debugf("counting %d instances", count)
	return count
}

func (ss Steps) Clone() Steps {
	clone := make(Steps)
	for name, step := range ss {
		newStep := step.Clone()
		newStep.Initialize()
		clone[name] = newStep
	}
	return clone
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
		if step.Ready() {
			ready = append(ready, stepName)
		}
		if !step.Waiting() {
			continue
		}
		isReady = true
		for _, dependency := range step.Depends {
			if subStep, exists := ss[dependency]; !exists {
				log.Panicf("step %s depends on unknows step %s", stepName, dependency)
			} else if !subStep.Done() && subStep.state != stepStateSkipped {
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
		if !step.Waiting() {
			continue
		}
		numWaiting += 1
	}
	return numWaiting
}

func (ss Steps) CheckWhen(all Handler, stepName string) (bool, error) {
	var numChecks int
	if step, exists := ss[stepName]; !exists {
		return false, fmt.Errorf("checking a 'when' on an undefined step %s", stepName)
	} else {
		numChecks = len(step.When)
		for _, whenCheck := range step.When {
			if !strings.Contains(whenCheck, "{{") || !strings.Contains(whenCheck, "}}") {
				whenCheck = fmt.Sprintf("{{if %s }}True{{end}}", whenCheck)
			}
			if t, err := template.New("when").Parse(whenCheck); err != nil {
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
	Commands  Commands `yaml:"commands"`
	Depends   []string `yaml:"depends,omitempty"`
	state     stepState
	When      []string   `yaml:"when,omitempty"`
	Matrix    MatrixArgs `yaml:"matrix,omitempty"`
	Instances Instances  `yaml:"-"`
}

func (s Step) Waiting() bool {
	return s.state == stepStateWaiting
}

func (s Step) Ready() bool {
	return s.state == stepStateReady
}

func (s *Step) Done() bool {
	if s.state == stepStateDone {
		return true
	}
	if s.Instances.Done() {
		if err := s.setState(stepStateDone); err != nil {
			log.Panicf("Could not issue done state for step %e", err)
		}
		return true
	}
	return false
}

func (s *Step) InstanceFinished(instance string) bool {
	if s.Done() {
		log.Fatalf("calling instanceFinished on a step that is already finished")
	}
	log.Debugf("instance done: %s", s.Instances[instance].Name())
	if i, exists := s.Instances[instance]; !exists {
		log.Fatalf("calling instanceFinished on an instance that does not exist")
	} else {
		i.done = true
	}
	return s.Done()
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
		Commands:  s.Commands.Clone(),
		Instances: s.Instances.Clone(),
		Depends:   s.Depends,
		state:     stepStateWaiting,
		When:      s.When,
		Matrix:    s.Matrix,
	}
}

func (s Step) StdOut() Result {
	return s.Instances.StdOut()
}

func (s Step) StdErr() Result {
	return s.Instances.StdErr()
}

func (s Step) Rc() int {
	return s.Instances.Rc()
}

func (s *Step) Initialize() {
	s.SetInstances()
}

func (s *Step) SetInstances() {
	if len(s.Instances) > 0 {
		return
	}
	s.Instances = make(Instances)
	for _, args := range s.Matrix.Instances() {
		s.Instances[args.String()] = NewInstance(args, s.Commands.Clone())
	}
}

func (s Step) GetInstances() Instances {
	//log.Debugf("Instances: %s", s.Instances.String())
	return s.Instances
}
