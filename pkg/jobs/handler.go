package jobs

import "os"

type Handler struct {
	Config  Config
	Steps   Steps
	Runners Runners
	Work    []string
	ToDo    chan string
	Done    chan string
}

func NewHandler(c Config) Handler {
	return Handler{
		Config: c,
		Steps:  c.Steps.Clone(),
		ToDo:   make(chan string, len(c.Steps)),
		Done:   make(chan string, len(c.Steps)),
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
	if err := h.Config.Checks.Run(h.Config.Conns); err != nil {
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
			h.ToDo <- name
			h.Steps.setStepState(name, stepStateScheduled)
		} else {
			h.Steps.setStepState(name, stepStateDone)
		}
	}
	return h.Steps.NumWaiting() > 0
}

func (h *Handler) processDone() {
	select {
	case doneStep := <-h.Done:
		if doneStep != "" {
			log.Debugf("This step is done: %s", doneStep)
			h.Steps.setStepState(doneStep, stepStateDone)
		}
	default:
		//log.Infof("break")
	}
}

func (h *Handler) checkAllDone() (done bool) {
	return h.Runners.Done()
}
