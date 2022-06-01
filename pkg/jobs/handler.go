package jobs

import "os"

type Work struct {
	Step string
	Index int
}

type Handler struct {
	Config  Config
	Steps   Steps
	Runners Runners
	ToDo    chan Work
	Done    chan Work
}

func NewHandler(c Config) Handler {
	return Handler{
		Config: c,
		Steps:  c.Steps,
		ToDo:   make(chan Work, c.Steps.GetNumInstances()),
		Done:   make(chan Work, c.Steps.GetNumInstances()),
	}
}

func (h *Handler) VerifyConfig() {
	log.Info("This is my config:\n", h.Config.String())
	if h.Config.Workdir != "" {
		log.Infof("Jumping to workdir %s", h.Config.Workdir)
		if err := os.Chdir(h.Config.Workdir); err != nil {
			log.Panicf("could not jump to dir %s", h.Config.Workdir)
		}
	}
	log.Info("Verifying config")
	h.Config.Verify()
	//if ! h.VerifyRoles() {
	//	log.Panicf("one or more connections are to an instance with different role")
	//}
}

func (h Handler) VerifyRoles() bool {
	for name, con := range h.Config.Conns {
		if ok, err := con.VerifyRole(); err != nil {
			log.Errorf("Error while verifying %s: %e", name, err)
			return false
		} else {
			if ok {
				log.Debugf("connection %s has expected role", name)
			} else {
				log.Debugf("connection %s has different role", name)
				return false
			}
		}
	}
	return true
}

func (h *Handler) RunSteps() {
	log.Info("Initializing runners")
	h.initRunners()
	log.Info("Waiting for all work to be scheduled")
	for {
		if !h.newWork() {
			break
		}
		h.processDone()
	}
	close(h.ToDo)
	log.Info("Waiting for all work to be done")
	for {
		if h.checkAllDone() {
			log.Debug("RunSteps: break")
			break
		}
	}
	close(h.Done)
	h.processDone()
	log.Info("All work is done")
}

func (h *Handler) RunChecks() {
	if len(h.Config.Checks) == 0 {
		return
	}
	log.Debug("Checking job results")
	if err := h.Config.Checks.Run(h.Config.Conns, nil); err != nil {
		log.Errorf("error occurred while running checks: %e", err)
	}
}

func (h *Handler) initRunners() {
	for i := 0; i < h.Config.Parallel; i++ {
		r := NewRunner(h, i)
		h.Runners = append(h.Runners, r)
		go r.Run()
	}
}

func (h *Handler) newWork() (done bool) {
	done = true
	for _, name := range h.Steps.GetReadySteps() {
		log.Infof("scheduling step %s", name)
		if result, err := h.Steps.CheckWhen(*h, name); err != nil {
			log.Errorf("error while checking step %s: %e", name, err)
			h.Steps.setStepState(name, stepStateSkipped)
		} else if result {
			instances := h.Steps[name].GetInstances()
			log.Debugf("scheduling %d instances for step %s", len(instances), name)
			for i, args := range instances {
				log.Debugf("scheduling step %s, instance %d (%s)", name, i, args.String())
				h.ToDo <- Work{name, i}
			}
			h.Steps.setStepState(name, stepStateScheduled)
		} else {
			h.Steps.setStepState(name, stepStateDone)
		}
	}
	return h.Steps.NumWaiting() > 0
}

func (h *Handler) processDone() {
	select {
	case doneInstance := <-h.Done:
		if doneInstance.Step != "" {
			log.Debugf("This step instance is done: %s.%d", doneInstance.Step, doneInstance.Index)
			h.Steps[doneInstance.Step].InstanceFinished(doneInstance.Index)
		}
	default:
		//log.Infof("break")
	}
}

func (h *Handler) checkAllDone() (done bool) {
	return h.Runners.Done()
}
