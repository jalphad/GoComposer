package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddConsumer[I, O, R any](c composer.Composer[I, O], f func(R) error, opts *ConsmrOpts[R]) {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setConsmrOpts(swf, opts)

	this := &taskConsumer[I, O, R]{
		taskBase: taskBase[I, O, I]{
			wf:    swf,
			name:  opts.Name,
			order: opts.order,
		},
		f: f,
	}
	subCh := make(chan R, 1)
	composer.AddSub(swf, opts.Input.Name(), subCh)
	this.sub = subCh
	swf.AddTask(this)
}

type ConsmrOpts[R any] struct {
	Name  string
	Input task.Dependency[R]
	order int
}

func setConsmrOpts[I, O, R any](c *composer.SimpleWorkflow[I, O], o *ConsmrOpts[R]) *ConsmrOpts[R] {
	if o == nil {
		return &ConsmrOpts[R]{
			Name:  fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input: task.WorkflowInput[R](task.Input),
			order: len(c.Tasks),
		}
	}
	if o.Input == nil {
		o.Input = task.WorkflowInput[R](task.Input)
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskConsumer[I, O, R any] struct {
	taskBase[I, O, I]
	f   func(R) error
	sub <-chan R
}

func (t *taskConsumer[I, O, R]) Name() string {
	return t.name
}

func (t *taskConsumer[I, O, R]) Compose() error {
	if t.f != nil {
		if t.sub != nil {
			t.wf.AddTransformFn(func() error {
				err := t.f(<-t.sub)
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

func (t *taskConsumer[I, O, R]) isDependency() {}
