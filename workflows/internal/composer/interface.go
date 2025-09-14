package composer

import "github.com/jalphad/gocomposer/workflows/internal/task"

type Composer[I, O any] interface {
	Input() task.Dependency[I]
	Compose() (func(I) (O, error), error)
}
