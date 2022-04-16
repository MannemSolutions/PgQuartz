package jobs

import "time"

type Runners []*Runner

type Runner struct {
	index  int
	config Config
	Steps Steps
	parent *Handler
	done bool
}

func (rs Runners) Done() bool {
	for _, r := range rs {
		if ! r.done {
			return false
		}
	}
	return true
}

func NewRunner(h *Handler, index int) *Runner {
	return &Runner{
		index: index,
		parent: h,
		config: h.config,
	}
}

func (r *Runner) Run () {
	// !!!! Locking !!!!!
	for {
		if stepName, ok := <-r.parent.toDo; ! ok {
			break
		} else if step, exists := r.parent.config.Steps[stepName]; !exists {
			log.Panicf("Runner %d: Trying to run a step %s that does not exist?", r.index, stepName)
		} else {
			log.Infof("Runner %d: Running the following command: %s", r.index, step.Command)
			time.Sleep(3 * time.Second)
			r.parent.done <- stepName
		}
	}
	log.Infof("Runner %d: Done", r.index)
	r.done = true
}
