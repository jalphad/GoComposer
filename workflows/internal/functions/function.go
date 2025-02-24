package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddFn[I, O, R, S any](c composer.Composer[I, O], f func(R) (S, error), opts *FnOpts[R]) task.Dependency[S] {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setOpts(swf, opts)
	if _, ok := opts.Input.(task.WorkflowInput[R]); ok {
		this := &taskInputFn[I, O, S]{
			taskBase: taskBase[I, O, S]{
				wf:    swf,
				name:  opts.Name,
				order: opts.order,
			},
		}
		if fi, ok := any(f).(func(I) (S, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			swf.AddErr(fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}
		this.pub = composer.SetPub[I, O, S](swf, opts.Name)
		swf.AddTask(this)

		return (&task.TaskDependency[S]{}).SetName(this.name)
	}
	this := &taskFn[I, O, R, S]{
		taskBase: taskBase[I, O, S]{
			wf:    swf,
			name:  opts.Name,
			order: opts.order,
		},
		f: f,
	}
	subCh := make(chan R, 1)
	composer.AddSub(swf, opts.Input.Name(), subCh)
	this.sub = subCh
	this.pub = composer.SetPub[I, O, S](swf, opts.Name)
	swf.AddTask(this)

	return (&task.TaskDependency[S]{}).SetName(this.name)
}

type FnOpts[R any] struct {
	Name  string
	Input task.Dependency[R]
	order int
}

func setOpts[I, O, R any](c *composer.SimpleWorkflow[I, O], o *FnOpts[R]) *FnOpts[R] {
	if o == nil {
		return &FnOpts[R]{
			Name:  fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input: task.WorkflowInput[R](task.Input),
			order: len(c.Tasks),
		}
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	if o.Input == nil {
		o.Input = task.WorkflowInput[R](task.Input)
	}
	o.order = len(c.Tasks)
	return o
}

type taskFn[I, O, R, S any] struct {
	taskBase[I, O, S]
	f   func(R) (S, error)
	sub <-chan R
}

func (t *taskFn[I, O, R, S]) Name() string {
	return t.name
}

func (t *taskFn[I, O, R, S]) Compose() error {
	if t.f != nil {
		if t.sub != nil && len(t.pub.Channels) != 0 {
			t.wf.AddTransformFn(func() error {
				res, err := t.f(<-t.sub)
				if err != nil {
					return err
				}
				for _, ch := range t.pub.Channels {
					ch <- res
				}

				return nil
			})
		} else if to, ok := t.toOutputFn(); ok {
			err := to.compose()
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

func (t *taskFn[I, O, R, S]) isDependency() {}

func (t *taskFn[I, O, R, S]) toOutputFn() (*taskOutputFn[I, O, R], bool) {
	fo, ok := any(t.f).(func(R) (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputFn[I, O, R]{
		taskBase: taskBase[I, O, O]{
			wf:    t.wf,
			name:  t.name,
			order: t.order,
		},
		f:   fo,
		sub: t.sub,
	}, true
}

type taskInputFn[I, O, S any] taskFn[I, O, I, S]

func (t *taskInputFn[I, O, S]) Name() string {
	return t.name
}

func (t *taskInputFn[I, O, S]) Compose() error {
	if t.f != nil && len(t.pub.Channels) != 0 {
		t.wf.AddInputFn(func(i I) error {
			res, err := t.f(i)
			if err != nil {
				return err
			}
			for _, ch := range t.pub.Channels {
				ch <- res
			}

			return nil
		}, t.order)
	} else {
		return fmt.Errorf("%w: function for %s is nil or output is not used", types.ErrCompose, t.name)
	}

	return nil
}

func (t *taskInputFn[I, O, S]) isDependency() {}

type taskOutputFn[I, O, R any] taskFn[I, O, R, O]

func (t *taskOutputFn[I, O, S]) Name() string {
	return t.name
}

func (t *taskOutputFn[I, O, R]) compose() error {
	if t.f != nil {
		if t.sub != nil {
			if ok := t.wf.SetOutputFn(func() (O, error) {
				return t.f(<-t.sub)
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
