package workflows

type Task interface {
	Name() string
	compose() error
}

type Dependency[O any] struct {
	name string
}

func (d *Dependency[O]) isDependency() {}

func (d *Dependency[O]) Name() string {
	return d.name
}

type iDependency[I any] interface {
	Name() string
	isDependency()
}

type WorkflowInput string

func (w WorkflowInput) Name() string {
	return string(w)
}

func (w WorkflowInput) isDependency() {}

const Input WorkflowInput = "INPUT"
