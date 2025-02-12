package workflows

type Task interface {
	Name() string
	compose() error
}

type Dependency[I any] interface {
	Name() string
	isDependency()
}

type WorkflowInput string

func (w WorkflowInput) Name() string {
	return string(w)
}

func (w WorkflowInput) isDependency() {}

const Input WorkflowInput = "INPUT"
