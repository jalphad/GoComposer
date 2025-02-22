package workflows

import (
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/functions"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

type (
	Composer[I, O any] = composer.Composer[I, O]
)

func NewComposer[I, O any]() Composer[I, O] {
	return composer.NewSimpleWorkflow[I, O]()
}

func AddFn[I, O, R, S any](c composer.Composer[I, O], f func(R) (S, error), d task.Dependency[R]) task.Dependency[S] {
	opts := &functions.FnOpts[R]{
		Input: d,
	}
	return functions.AddFn(c, f, opts)
}

func AddBiFn[I, O, Q, R, S any](c composer.Composer[I, O], f func(Q, R) (S, error), d1 task.Dependency[Q], d2 task.Dependency[R]) task.Dependency[S] {
	opts := &functions.BiFnOpts[Q, R]{
		Input1: d1,
		Input2: d2,
	}
	return functions.AddBiFn(c, f, opts)
}
