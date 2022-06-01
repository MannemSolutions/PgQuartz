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
		if instance, ok := <-r.parent.ToDo; !ok {
			break
		} else if step, exists := r.parent.Steps[instance.Step]; !exists {
			log.Panicf("Runner %d: Trying to run a step %s that does not exist?", r.index, instance.Step)
		} else {
			args := step.GetInstanceArgs(instance.Index)
			log.Debugf("Runner %d: Running step %s, instance %d with args %s", r.index, instance.Step,
				instance.Index, args.String())
			if err := step.Commands.Run(r.config.Conns, args); err != nil {
				log.Errorf("error occurred while running step %s: %e", instance.Step, err)
			}
			r.parent.Done <- instance
		}
	}
	log.Debugf("Runner %d: Done", r.index)
	r.done = true
}
