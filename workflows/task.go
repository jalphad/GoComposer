package workflows

type Task interface {
	Name() string
	compose() error
}

type Dependency[I any] interface {
	Name() string
	val(I)
}

type dependency[O any] struct {
	name string
}

func (d *dependency[O]) val(_ O) {}

func (d *dependency[O]) Name() string {
	return d.name
}

type workflowInput[I any] string

func (w workflowInput[I]) Name() string {
	return string(w)
}

func (w workflowInput[I]) val(_ I) {}

const Input = "INPUT"
