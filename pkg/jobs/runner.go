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
		if work, ok := <-r.parent.ToDo; !ok {
			break
		} else if step, sExists := r.parent.Steps[work.Step]; !sExists {
			log.Panicf("Runner %d: Trying to run a step %s that does not exist?", r.index, work.Step)
		} else if instance, iExists := step.Instances[work.ArgKey]; !iExists {
			log.Panicf("Runner %d: Trying to run an instance [%s].[%s] that does not exist?", r.index, work.Step, work.ArgKey)
		} else {
			log.Debugf("Runner %d: Running step [%s].[%s]", r.index, work.Step, work.ArgKey)
			if err := instance.commands.Run(r.config.Conns, instance.args); err != nil {
				log.Errorf("Runner %d: Error occurred while running step instance [%s].[%s]: %e", r.index, work.Step, work.ArgKey, err)
			}
			r.parent.Done <- work
		}
	}
	log.Debugf("Runner %d: Done", r.index)
	r.done = true
}
