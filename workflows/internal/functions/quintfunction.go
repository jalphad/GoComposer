package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddQuintFn[I, O, Q, R, S, T, U, V any](c composer.Composer[I, O], f func(Q, R, S, T, U) (V, error), opts *QuintFnOpts[Q, R, S, T, U]) task.Dependency[V] {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setQuintFnOpts(swf, opts)

	this := &taskQuintFn[I, O, Q, R, S, T, U, V]{
		taskBase: taskBase[I, O, V]{
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
	sub5Ch := make(chan U, 1)
	composer.AddSub(swf, opts.Input5.Name(), sub5Ch)
	this.sub5 = sub5Ch
	this.pub = composer.SetPub[I, O, V](swf, opts.Name)
	swf.AddTask(this)

	return (&task.TaskDependency[V]{}).SetName(this.name)
}

type QuintFnOpts[Q, R, S, T, U any] struct {
	Name   string
	Input1 task.Dependency[Q]
	Input2 task.Dependency[R]
	Input3 task.Dependency[S]
	Input4 task.Dependency[T]
	Input5 task.Dependency[U]
	order  int
}

func setQuintFnOpts[I, O, Q, R, S, T, U any](c *composer.SimpleWorkflow[I, O], o *QuintFnOpts[Q, R, S, T, U]) *QuintFnOpts[Q, R, S, T, U] {
	if o == nil {
		return &QuintFnOpts[Q, R, S, T, U]{
			Name:   fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input1: task.WorkflowInput[Q](task.Input),
			Input2: task.WorkflowInput[R](task.Input),
			Input3: task.WorkflowInput[S](task.Input),
			Input4: task.WorkflowInput[T](task.Input),
			Input5: task.WorkflowInput[U](task.Input),
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
	if o.Input5 == nil {
		o.Input5 = task.WorkflowInput[U](task.Input)
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskQuintFn[I, O, Q, R, S, T, U, V any] struct {
	taskBase[I, O, V]
	f    func(Q, R, S, T, U) (V, error)
	sub1 <-chan Q
	sub2 <-chan R
	sub3 <-chan S
	sub4 <-chan T
	sub5 <-chan U
}

func (t *taskQuintFn[I, O, Q, R, S, T, U, V]) Name() string {
	return t.name
}

func (t *taskQuintFn[I, O, Q, R, S, T, U, V]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && t.sub3 != nil && t.sub4 != nil && t.sub5 != nil && len(t.pub.Channels) != 0 {
			t.wf.AddTransformFn(func() error {
				res, err := t.f(<-t.sub1, <-t.sub2, <-t.sub3, <-t.sub4, <-t.sub5)
				if err != nil {
					return err
				}
				for _, ch := range t.pub.Channels {
					ch <- res
				}

				return nil
			})
		} else if to, ok := t.toOutputQuintFn(); ok {
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

func (t *taskQuintFn[I, O, Q, R, S, T, U, V]) isDependency() {}

func (t *taskQuintFn[I, O, Q, R, S, T, U, V]) toOutputQuintFn() (*taskOutputQuintFn[I, O, Q, R, S, T, U], bool) {
	fo, ok := any(t.f).(func(Q, R, S, T, U) (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputQuintFn[I, O, Q, R, S, T, U]{
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
		sub5: t.sub5,
	}, true
}

type taskOutputQuintFn[I, O, Q, R, S, T, U any] taskQuintFn[I, O, Q, R, S, T, U, O]

func (t *taskOutputQuintFn[I, O, Q, R, S, T, U]) Name() string {
	return t.name
}

func (t *taskOutputQuintFn[I, O, Q, R, S, T, U]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && t.sub3 != nil && t.sub4 != nil && t.sub5 != nil {
			if ok := t.wf.SetOutputFn(func() (O, error) {
				return t.f(<-t.sub1, <-t.sub2, <-t.sub3, <-t.sub4, <-t.sub5)
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
