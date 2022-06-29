package jobs

import "os"

type Work struct {
	Step   string
	ArgKey string
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
	log.Debug("This is my config:\n", h.Config.String())
	log.Debugf("Jumping to workdir %s", h.Config.Workdir)
	if err := os.Chdir(h.Config.Workdir); err != nil {
		log.Panicf("could not jump to dir %s", h.Config.Workdir)
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
	log.Info("Checking job results")
	h.Config.Checks.Run(h.Config.Conns)
	log.Info("Job finished successfully")
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
		log.Infof("Scheduling step %s", name)
		if result, err := h.Steps.CheckWhen(*h, name); err != nil {
			log.Errorf("Error while checking step %s: %e", name, err)
			h.Steps.setStepState(name, stepStateSkipped)
		} else if result {
			instances := h.Steps[name].GetInstances()
			log.Debugf("Scheduling %d instances for step %s", len(instances), name)
			for _, i := range instances {
				instanceName := i.Name()
				log.Debugf("Scheduling instance [%s].[%s]", name, instanceName)
				h.ToDo <- Work{name, instanceName}
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
			log.Debugf("This step instance is done: [%s].[%s]", doneInstance.Step, doneInstance.ArgKey)
			h.Steps.InstanceFinished(doneInstance)
		}
	default:
		//log.Infof("break")
	}
}

func (h *Handler) checkAllDone() (done bool) {
	return h.Runners.Done()
}
