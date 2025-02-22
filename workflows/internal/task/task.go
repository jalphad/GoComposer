package task

type Task interface {
	Name() string
	Compose() error
}

type Dependency[I any] interface {
	Name() string
	ofType(I)
}

type TaskDependency[O any] struct {
	name string
}

func (d *TaskDependency[O]) SetName(name string) *TaskDependency[O] {
	d.name = name
	return d
}

func (d *TaskDependency[O]) ofType(_ O) {}

func (d *TaskDependency[O]) Name() string {
	return d.name
}

type WorkflowInput[I any] string

func (w WorkflowInput[I]) Name() string {
	return string(w)
}

func (w WorkflowInput[I]) ofType(_ I) {}

const Input = "INPUT"
