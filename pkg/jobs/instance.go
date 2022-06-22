package jobs

type Instances map[string]*Instance

func (is Instances) Clone() (clone Instances) {
	clone = make(Instances)
	for _, i := range is {
		clone[i.name] = i.Clone()
	}
	return clone
}

func (is Instances) Done() bool {
	for _, i := range is {
		if !i.done {
			return false
		}
	}
	return true
}

func (is Instances) StdOut() (r Result) {
	for _, i := range is {
		r = append(r, i.StdOut()...)
	}
	return r
}

func (is Instances) StdErr() (r Result) {
	for _, i := range is {
		r = append(r, i.StdErr()...)
	}
	return r
}

func (is Instances) Rc() (rc int) {
	for _, instance := range is {
		rc += instance.commands.Rc()
	}
	return rc
}

type Instance struct {
	name     string
	args     InstanceArguments
	commands Commands
	done     bool
}

func NewInstance(args InstanceArguments, commands Commands) *Instance {
	return &Instance{
		args:     args,
		name:     args.String(),
		commands: commands,
	}
}

func (i Instance) Clone() (clone *Instance) {
	return &Instance{
		args:     i.args.Clone(),
		commands: i.commands.Clone(),
	}
}

func (i Instance) StdOut() (r Result) {
	return i.commands.StdOut()
}

func (i Instance) StdErr() (r Result) {
	return i.commands.StdErr()
}

func (i Instance) Name() string {
	return i.name
}
