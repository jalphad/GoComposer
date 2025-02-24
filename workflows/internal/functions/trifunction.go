package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddTriFn[I, O, Q, R, S, T any](c composer.Composer[I, O], f func(Q, R, S) (T, error), opts *TriFnOpts[Q, R, S]) task.Dependency[T] {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setTriFnOpts(swf, opts)
	if _, ok := opts.Input1.(task.WorkflowInput[Q]); ok {
		this := &taskInputFirstTriFn[I, O, R, S, T]{
			taskBase: taskBase[I, O, T]{
				wf:    swf,
				name:  opts.Name,
				order: opts.order,
			},
		}
		if fi, ok := any(f).(func(I, R, S) (T, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			swf.AddErr(fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}
		sub2Ch := make(chan R, 1)
		composer.AddSub(swf, opts.Input2.Name(), sub2Ch)
		this.sub2 = sub2Ch
		sub3Ch := make(chan S, 1)
		composer.AddSub(swf, opts.Input3.Name(), sub3Ch)
		this.sub3 = sub3Ch

		this.pub = composer.SetPub[I, O, T](swf, opts.Name)
		swf.AddTask(this)

		return (&task.TaskDependency[T]{}).SetName(this.name)
	} else if _, ok := opts.Input2.(task.WorkflowInput[R]); ok {
		this := &taskInputSecondTriFn[I, O, Q, S, T]{
			taskBase: taskBase[I, O, T]{
				wf:    swf,
				name:  opts.Name,
				order: opts.order,
			},
		}
		if fi, ok := any(f).(func(Q, I, S) (T, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			swf.AddErr(fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}
		sub1Ch := make(chan Q, 1)
		composer.AddSub(swf, opts.Input1.Name(), sub1Ch)
		this.sub1 = sub1Ch
		sub3Ch := make(chan S, 1)
		composer.AddSub(swf, opts.Input3.Name(), sub3Ch)
		this.sub3 = sub3Ch

		this.pub = composer.SetPub[I, O, T](swf, opts.Name)
		swf.AddTask(this)

		return (&task.TaskDependency[T]{}).SetName(this.name)
	} else if _, ok := opts.Input3.(task.WorkflowInput[S]); ok {
		this := &taskInputThirdTriFn[I, O, Q, R, T]{
			taskBase: taskBase[I, O, T]{
				wf:    swf,
				name:  opts.Name,
				order: opts.order,
			},
		}
		if fi, ok := any(f).(func(Q, R, I) (T, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			swf.AddErr(fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}
		sub1Ch := make(chan Q, 1)
		composer.AddSub(swf, opts.Input1.Name(), sub1Ch)
		this.sub1 = sub1Ch
		sub2Ch := make(chan R, 1)
		composer.AddSub(swf, opts.Input2.Name(), sub2Ch)
		this.sub2 = sub2Ch

		this.pub = composer.SetPub[I, O, T](swf, opts.Name)
		swf.AddTask(this)

		return (&task.TaskDependency[T]{}).SetName(this.name)
	}
	this := &taskTriFn[I, O, Q, R, S, T]{
		taskBase: taskBase[I, O, T]{
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
	this.pub = composer.SetPub[I, O, T](swf, opts.Name)
	swf.AddTask(this)

	return (&task.TaskDependency[T]{}).SetName(this.name)
}

type TriFnOpts[Q, R, S any] struct {
	Name   string
	Input1 task.Dependency[Q]
	Input2 task.Dependency[R]
	Input3 task.Dependency[S]
	order  int
}

func setTriFnOpts[I, O, Q, R, S any](c *composer.SimpleWorkflow[I, O], o *TriFnOpts[Q, R, S]) *TriFnOpts[Q, R, S] {
	if o == nil {
		return &TriFnOpts[Q, R, S]{
			Name:   fmt.Sprintf("Task%d", len(c.Tasks)+1),
			Input1: nil,
			Input2: nil,
			Input3: nil,
			order:  len(c.Tasks),
		}
	} else if o.Input1 == nil {
		o.Input1 = task.WorkflowInput[Q](task.Input)
	} else if o.Input2 == nil {
		o.Input2 = task.WorkflowInput[R](task.Input)
	} else if o.Input3 == nil {
		o.Input3 = task.WorkflowInput[S](task.Input)
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskTriFn[I, O, Q, R, S, T any] struct {
	taskBase[I, O, T]
	f    func(Q, R, S) (T, error)
	sub1 <-chan Q
	sub2 <-chan R
	sub3 <-chan S
}

func (t *taskTriFn[I, O, Q, R, S, T]) Name() string {
	return t.name
}

func (t *taskTriFn[I, O, Q, R, S, T]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && t.sub3 != nil && len(t.pub.Channels) != 0 {
			t.wf.AddTransformFn(func() error {
				res, err := t.f(<-t.sub1, <-t.sub2, <-t.sub3)
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

func (t *taskTriFn[I, O, Q, R, S, T]) isDependency() {}

func (t *taskTriFn[I, O, Q, R, S, T]) toOutputBiFn() (*taskOutputTriFn[I, O, Q, R, S], bool) {
	fo, ok := any(t.f).(func(Q, R, S) (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputTriFn[I, O, Q, R, S]{
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

type taskInputFirstTriFn[I, O, R, S, T any] taskTriFn[I, O, I, R, S, T]

func (t *taskInputFirstTriFn[I, O, R, S, T]) Name() string {
	return t.name
}

func (t *taskInputFirstTriFn[I, O, R, S, T]) Compose() error {
	if t.f != nil && len(t.pub.Channels) != 0 {
		t.wf.AddInputFn(func(i I) error {
			res, err := t.f(i, <-t.sub2, <-t.sub3)
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

func (t *taskInputFirstTriFn[I, O, R, S, T]) isDependency() {}

type taskInputSecondTriFn[I, O, Q, S, T any] taskTriFn[I, O, Q, I, S, T]

func (t *taskInputSecondTriFn[I, O, Q, S, T]) Name() string {
	return t.name
}

func (t *taskInputSecondTriFn[I, O, Q, S, T]) Compose() error {
	if t.f != nil && len(t.pub.Channels) != 0 {
		t.wf.AddInputFn(func(i I) error {
			res, err := t.f(<-t.sub1, i, <-t.sub3)
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

func (t *taskInputSecondTriFn[I, O, Q, S, T]) isDependency() {}

type taskInputThirdTriFn[I, O, Q, R, T any] taskTriFn[I, O, Q, R, I, T]

func (t *taskInputThirdTriFn[I, O, Q, R, T]) Name() string {
	return t.name
}

func (t *taskInputThirdTriFn[I, O, Q, R, T]) Compose() error {
	if t.f != nil && len(t.pub.Channels) != 0 {
		t.wf.AddInputFn(func(i I) error {
			res, err := t.f(<-t.sub1, <-t.sub2, i)
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

func (t *taskInputThirdTriFn[I, O, Q, S, T]) isDependency() {}

type taskOutputTriFn[I, O, Q, R, S any] taskTriFn[I, O, Q, R, S, O]

func (t *taskOutputTriFn[I, O, Q, R, S]) Name() string {
	return t.name
}

func (t *taskOutputTriFn[I, O, Q, R, S]) Compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil {
			if ok := t.wf.SetOutputFn(func() (O, error) {
				return t.f(<-t.sub1, <-t.sub2, <-t.sub3)
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
