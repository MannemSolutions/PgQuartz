package jobs

import "os"

type Handler struct {
	config  Config
	steps   Steps
	runners Runners
	work    []string
	toDo    chan string
	done    chan string
}

func NewHandler(c Config) Handler {
	return Handler{
		config: c,
		steps:  c.Steps.Clone(),
		toDo:   make(chan string, len(c.Steps)),
		done:   make(chan string, len(c.Steps)),
	}
}

func (h *Handler) VerifyConfig() {
	log.Info("This is my config:\n", h.config.String())
	if h.config.Workdir != "" {
		log.Infof("Jumping to workdir %s", h.config.Workdir)
		os.Chdir(h.config.Workdir)
	}
	log.Info("Verifying config")
	h.config.Verify()
}

func (h *Handler) RunSteps() {
	log.Info("Initializing runners")
	h.initRunners()
	log.Info("Waiting for all work to be scheduled")
	for {
		if !h.newWork() {
			break
		}
		select {
		case doneStep := <-h.done:
			log.Infof("This step is done: %s", doneStep)
			h.steps.setStepState(doneStep, stepStateDone)
		default:
			//log.Infof("break")
		}
	}
	close(h.toDo)
	log.Info("Waiting for all work to be done")
	for {
		if h.checkAllDone() {
			log.Info("break")
			break
		}
	}
	close(h.done)
	h.processDone()
	log.Info("All work is done")
	log.Sync()
}

func (h *Handler) RunChecks() {
	if len(h.config.Checks) == 0 {
		return
	}
	log.Info("Checking results for these jobs")
	if err := h.config.Checks.Run(h.config.Conns); err != nil {
		log.Errorf("error occurred while running checks: %e", err)
	}
}

func (h *Handler) initRunners() {
	for i := 0; i < h.config.Parallel; i++ {
		r := NewRunner(h, i)
		h.runners = append(h.runners, r)
		go r.Run()
	}
}

func (h *Handler) newWork() (done bool) {
	done = true
	for _, name := range h.steps.GetReadySteps() {
		log.Infof("scheduling step %s", name)
		h.toDo <- name
		h.steps[name].state = stepStateScheduled
	}
	return h.steps.NumWaiting() > 0
}

func (h *Handler) processDone() {
	select {
	case doneStep := <-h.done:
		log.Infof("This step is done: %s", doneStep)
		h.steps.setStepState(doneStep, stepStateDone)
	default:
		//log.Infof("break")
	}
}

func (h *Handler) checkAllDone() (done bool) {
	return h.runners.Done()
}
