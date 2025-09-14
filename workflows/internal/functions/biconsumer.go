package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddBiConsumer[I, O, R, S any](c composer.Composer[I, O], f func(R, S) error, opts *BiConsmrOpts[R, S]) {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setBiConsmrOpts(swf, opts)

	this := &taskBiConsumer[I, O, R, S]{
		taskBase: taskBase[I, O, I]{
			wf:    swf,
			name:  opts.Name,
			order: opts.order,
		},
		f: f,
	}
	sub1Ch := make(chan R, 1)
	composer.AddSub(swf, opts.Input1.Name(), sub1Ch)
	this.sub1 = sub1Ch
	sub2Ch := make(chan S, 1)
	composer.AddSub(swf, opts.Input2.Name(), sub2Ch)
	this.sub2 = sub2Ch
	swf.AddTask(this)
}

type BiConsmrOpts[R, S any] struct {
	Name   string
	Input1 task.Dependency[R]
	Input2 task.Dependency[S]
	order  int
}

func setBiConsmrOpts[I, O, R, S any](c *composer.SimpleWorkflow[I, O], o *BiConsmrOpts[R, S]) *BiConsmrOpts[R, S] {
	if o == nil {
		return &BiConsmrOpts[R, S]{
			Name:   fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input1: task.WorkflowInput[R](task.Input),
			Input2: task.WorkflowInput[S](task.Input),
			order:  len(c.Tasks),
		}
	}
	if o.Input1 == nil {
		o.Input1 = task.WorkflowInput[R](task.Input)
	}
	if o.Input2 == nil {
		o.Input2 = task.WorkflowInput[S](task.Input)
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskBiConsumer[I, O, R, S any] struct {
	taskBase[I, O, I]
	f    func(R, S) error
	sub1 <-chan R
	sub2 <-chan S
}

func (t *taskBiConsumer[I, O, R, S]) Name() string {
	return t.name
}

func (t *taskBiConsumer[I, O, R, S]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil {
			t.wf.AddTransformFn(func() error {
				err := t.f(<-t.sub1, <-t.sub2)
				if err != nil {
					return err
				}

				return nil
			})
		} else {
			return fmt.Errorf("%w: function for %s is missing input", types.ErrCompose, t.name)
		}
	} else {
		return fmt.Errorf("%w: function for %s is nil", types.ErrCompose, t.name)
	}

	return nil
}

func (t *taskBiConsumer[I, O, R, S]) isDependency() {}
