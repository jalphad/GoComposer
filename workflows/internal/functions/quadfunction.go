package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddQuadFn[I, O, Q, R, S, T, U any](c composer.Composer[I, O], f func(Q, R, S, T) (U, error), opts *QuadFnOpts[Q, R, S, T]) task.Dependency[U] {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setQuadFnOpts(swf, opts)

	this := &taskQuadFn[I, O, Q, R, S, T, U]{
		taskBase: taskBase[I, O, U]{
			wf:    swf,
			name:  opts.Name,
			order: opts.order,
		},
		f: f,
	}
	sub1Ch := make(chan Q, 1)
	composer.AddSub(swf, opts.Input1.Name(), sub1Ch)
	this.sub1 = sub1Ch
	sub2Ch := make(chan R, 1)
	composer.AddSub(swf, opts.Input2.Name(), sub2Ch)
	this.sub2 = sub2Ch
	sub3Ch := make(chan S, 1)
	composer.AddSub(swf, opts.Input3.Name(), sub3Ch)
	this.sub3 = sub3Ch
	sub4Ch := make(chan T, 1)
	composer.AddSub(swf, opts.Input4.Name(), sub4Ch)
	this.sub4 = sub4Ch
	this.pub = composer.SetPub[I, O, U](swf, opts.Name)
	swf.AddTask(this)

	return (&task.TaskDependency[U]{}).SetName(this.name)
}

type QuadFnOpts[Q, R, S, T any] struct {
	Name   string
	Input1 task.Dependency[Q]
	Input2 task.Dependency[R]
	Input3 task.Dependency[S]
	Input4 task.Dependency[T]
	order  int
}

func setQuadFnOpts[I, O, Q, R, S, T any](c *composer.SimpleWorkflow[I, O], o *QuadFnOpts[Q, R, S, T]) *QuadFnOpts[Q, R, S, T] {
	if o == nil {
		return &QuadFnOpts[Q, R, S, T]{
			Name:   fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input1: task.WorkflowInput[Q](task.Input),
			Input2: task.WorkflowInput[R](task.Input),
			Input3: task.WorkflowInput[S](task.Input),
			Input4: task.WorkflowInput[T](task.Input),
			order:  len(c.Tasks),
		}
	}
	if o.Input1 == nil {
		o.Input1 = task.WorkflowInput[Q](task.Input)
	}
	if o.Input2 == nil {
		o.Input2 = task.WorkflowInput[R](task.Input)
	}
	if o.Input3 == nil {
		o.Input3 = task.WorkflowInput[S](task.Input)
	}
	if o.Input4 == nil {
		o.Input4 = task.WorkflowInput[T](task.Input)
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskQuadFn[I, O, Q, R, S, T, U any] struct {
	taskBase[I, O, U]
	f    func(Q, R, S, T) (U, error)
	sub1 <-chan Q
	sub2 <-chan R
	sub3 <-chan S
	sub4 <-chan T
}

func (t *taskQuadFn[I, O, Q, R, S, T, U]) Name() string {
	return t.name
}

func (t *taskQuadFn[I, O, Q, R, S, T, U]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && t.sub3 != nil && t.sub4 != nil && len(t.pub.Channels) != 0 {
			t.wf.AddTransformFn(func() error {
				res, err := t.f(<-t.sub1, <-t.sub2, <-t.sub3, <-t.sub4)
				if err != nil {
					return err
				}
				for _, ch := range t.pub.Channels {
					ch <- res
				}

				return nil
			})
		} else if to, ok := t.toOutputQuadFn(); ok {
			err := to.Compose()
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%w: function for %s is missing input or output is not used", types.ErrCompose, t.name)
		}
	} else {
		return fmt.Errorf("%w: function for %s is nil", types.ErrCompose, t.name)
	}

	return nil
}

func (t *taskQuadFn[I, O, Q, R, S, T, U]) isDependency() {}

func (t *taskQuadFn[I, O, Q, R, S, T, U]) toOutputQuadFn() (*taskOutputQuadFn[I, O, Q, R, S, T], bool) {
	fo, ok := any(t.f).(func(Q, R, S, T) (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputQuadFn[I, O, Q, R, S, T]{
		taskBase: taskBase[I, O, O]{
			name:  t.name,
			order: t.order,
			wf:    t.wf,
		},
		f:    fo,
		sub1: t.sub1,
		sub2: t.sub2,
		sub3: t.sub3,
		sub4: t.sub4,
	}, true
}

type taskOutputQuadFn[I, O, Q, R, S, T any] taskQuadFn[I, O, Q, R, S, T, O]

func (t *taskOutputQuadFn[I, O, Q, R, S, T]) Name() string {
	return t.name
}

func (t *taskOutputQuadFn[I, O, Q, R, S, T]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && t.sub3 != nil && t.sub4 != nil {
			if ok := t.wf.SetOutputFn(func() (O, error) {
				return t.f(<-t.sub1, <-t.sub2, <-t.sub3, <-t.sub4)
			}, t.order); !ok {
				return fmt.Errorf("%w: error composing task %s, multiple output functions", types.ErrCompose, t.name)
			}
		} else {
			return fmt.Errorf("%w: function for %s is missing input", types.ErrCompose, t.name)
		}
	} else {
		return fmt.Errorf("%w: function for %s is nil", types.ErrCompose, t.name)
	}

	return nil
}
