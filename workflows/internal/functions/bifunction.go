package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddBiFn[I, O, Q, R, S any](c composer.Composer[I, O], f func(Q, R) (S, error), opts *BiFnOpts[Q, R]) task.Dependency[S] {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setBiFnOpts(swf, opts)
	if _, ok := opts.Input1.(task.WorkflowInput[Q]); ok {
		this := &taskInputFirstBiFn[I, O, R, S]{
			taskBase: taskBase[I, O, S]{
				wf:    swf,
				name:  opts.Name,
				order: opts.order,
			},
		}
		if fi, ok := any(f).(func(I, R) (S, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			swf.AddErr(fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}

		this.pub = composer.SetPub[I, O, S](swf, opts.Name)
		swf.AddTask(this)

		return (&task.TaskDependency[S]{}).SetName(this.name)
	} else if _, ok := opts.Input2.(task.WorkflowInput[R]); ok {
		this := &taskInputSecondBiFn[I, O, Q, S]{
			taskBase: taskBase[I, O, S]{
				wf:    swf,
				name:  opts.Name,
				order: opts.order,
			},
		}
		if fi, ok := any(f).(func(Q, I) (S, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			swf.AddErr(fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}

		this.pub = composer.SetPub[I, O, S](swf, opts.Name)
		swf.AddTask(this)

		return (&task.TaskDependency[S]{}).SetName(this.name)
	}
	this := &taskBiFn[I, O, Q, R, S]{
		taskBase: taskBase[I, O, S]{
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
	this.pub = composer.SetPub[I, O, S](swf, opts.Name)
	swf.AddTask(this)

	return (&task.TaskDependency[S]{}).SetName(this.name)
}

type BiFnOpts[Q, R any] struct {
	Name   string
	Input1 task.Dependency[Q]
	Input2 task.Dependency[R]
	order  int
}

func setBiFnOpts[I, O, Q, R any](c *composer.SimpleWorkflow[I, O], o *BiFnOpts[Q, R]) *BiFnOpts[Q, R] {
	if o == nil {
		return &BiFnOpts[Q, R]{
			Name:   fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input1: nil,
			Input2: nil,
			order:  len(c.Tasks),
		}
	} else if o.Input1 == nil {
		o.Input1 = task.WorkflowInput[Q](task.Input)
	} else if o.Input2 == nil {
		o.Input2 = task.WorkflowInput[R](task.Input)
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskBiFn[I, O, Q, R, S any] struct {
	taskBase[I, O, S]
	f    func(Q, R) (S, error)
	sub1 <-chan Q
	sub2 <-chan R
}

func (t *taskBiFn[I, O, Q, R, S]) Name() string {
	return t.name
}

func (t *taskBiFn[I, O, Q, R, S]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && len(t.pub.Channels) != 0 {
			t.wf.AddTransformFn(func() error {
				res, err := t.f(<-t.sub1, <-t.sub2)
				if err != nil {
					return err
				}
				for _, ch := range t.pub.Channels {
					ch <- res
				}

				return nil
			})
		} else if to, ok := t.toOutputBiFn(); ok {
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

func (t *taskBiFn[I, O, Q, R, S]) isDependency() {}

func (t *taskBiFn[I, O, Q, R, S]) toOutputBiFn() (*taskOutputBiFn[I, O, Q, R], bool) {
	fo, ok := any(t.f).(func(Q, R) (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputBiFn[I, O, Q, R]{
		taskBase: taskBase[I, O, O]{
			name:  t.name,
			order: t.order,
			wf:    t.wf,
		},
		f:    fo,
		sub1: t.sub1,
		sub2: t.sub2,
	}, true
}

type taskInputFirstBiFn[I, O, R, S any] taskBiFn[I, O, I, R, S]

func (t *taskInputFirstBiFn[I, O, R, S]) Name() string {
	return t.name
}

func (t *taskInputFirstBiFn[I, O, R, S]) Compose() error {
	if t.f != nil && len(t.pub.Channels) != 0 {
		t.wf.AddInputFn(func(i I) error {
			res, err := t.f(i, <-t.sub2)
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

func (t *taskInputFirstBiFn[I, O, R, S]) isDependency() {}

type taskInputSecondBiFn[I, O, Q, S any] taskBiFn[I, O, Q, I, S]

func (t *taskInputSecondBiFn[I, O, R, S]) Name() string {
	return t.name
}

func (t *taskInputSecondBiFn[I, O, R, S]) Compose() error {
	if t.f != nil && len(t.pub.Channels) != 0 {
		t.wf.AddInputFn(func(i I) error {
			res, err := t.f(<-t.sub1, i)
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

func (t *taskInputSecondBiFn[I, O, R, S]) isDependency() {}

type taskOutputBiFn[I, O, Q, R any] taskBiFn[I, O, Q, R, O]

func (t *taskOutputBiFn[I, O, Q, R]) Name() string {
	return t.name
}

func (t *taskOutputBiFn[I, O, Q, R]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil {
			if ok := t.wf.SetOutputFn(func() (O, error) {
				return t.f(<-t.sub1, <-t.sub2)
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
