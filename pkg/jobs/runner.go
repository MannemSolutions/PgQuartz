package jobs

type Runners []*Runner

type Runner struct {
	index  int
	config Config
	Steps  Steps
	parent *Handler
	done   bool
}

func (rs Runners) Done() bool {
	for _, r := range rs {
		if !r.done {
			return false
		}
	}
	return true
}

func NewRunner(h *Handler, index int) *Runner {
	return &Runner{
		index:  index,
		parent: h,
		config: h.Config,
	}
}

func (r *Runner) Run() {
	for {
		if stepName, ok := <-r.parent.ToDo; !ok {
			break
		} else if step, exists := r.parent.Config.Steps[stepName]; !exists {
			log.Panicf("Runner %d: Trying to run a step %s that does not exist?", r.index, stepName)
		} else {
			if err := step.Commands.Run(r.config.Conns); err != nil {
				log.Errorf("error occurred while running step %s: %e", stepName, err)
			}
			r.parent.Done <- stepName
		}
	}
	log.Debugf("Runner %d: Done", r.index)
	r.done = true
}
