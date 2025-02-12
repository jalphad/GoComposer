package workflows

import (
	"fmt"
	"github.com/jalphad/gocomposer/types"
)

func AddFn[I, O, R, S any](c *Composer[I, O], f func(R) (S, error), opts *FnOpts[S]) Dependency[S] {
	opts = setOpts(c, opts)
	if _, ok := opts.DependsOn.(WorkflowInput); ok {
		this := &taskInputFn[I, O, S]{
			wf:   c,
			name: opts.Name,
		}
		if fi, ok := any(f).(func(I) (S, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			c.errs = append(c.errs, fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}

		this.pub = setPub[I, O, S](c, opts.Name)
		c.tasks = append(c.tasks, this)

		return this
	}
	this := &taskFn[I, O, R, S]{
		wf:   c,
		f:    f,
		name: opts.Name,
	}
	subCh := make(chan R, 1)
	addSub(c, opts.DependsOn.Name(), subCh)
	this.sub = subCh
	this.pub = setPub[I, O, S](c, opts.Name)
	c.tasks = append(c.tasks, this)

	return this
}

func NewFnOpts[O any](dependsOn Dependency[O]) *FnOpts[O] {
	return &FnOpts[O]{DependsOn: dependsOn}
}

type FnOpts[O any] struct {
	Name      string
	DependsOn Dependency[O]
}

func setOpts[I, O, S any](c *Composer[I, O], o *FnOpts[S]) *FnOpts[S] {
	if o == nil {
		return &FnOpts[S]{
			Name:      fmt.Sprintf("Task%d", len(c.tasks)+1),
			DependsOn: Input,
		}
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.tasks)+1)
	}
	return o
}

type taskFn[I, O, R, S any] struct {
	name string
	wf   *Composer[I, O]
	f    func(R) (S, error)
	pub  *pubImpl[S]
	sub  <-chan R
}

func (t *taskFn[I, O, R, S]) Name() string {
	return t.name
}

func (t *taskFn[I, O, R, S]) compose() error {
	if t.f != nil {
		if t.sub != nil && len(t.pub.channels) != 0 {
			t.wf.transformFns = append(t.wf.transformFns, func() error {
				res, err := t.f(<-t.sub)
				if err != nil {
					return err
				}
				for _, ch := range t.pub.channels {
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
		name: t.name,
		wf:   t.wf,
		f:    fo,
		sub:  t.sub,
	}, true
}

type taskInputFn[I, O, S any] taskFn[I, O, I, S]

func (t *taskInputFn[I, O, S]) Name() string {
	return t.name
}

func (t *taskInputFn[I, O, S]) compose() error {
	if t.f != nil && len(t.pub.channels) != 0 {
		t.wf.inputFns = append(t.wf.inputFns, func(i I) error {
			res, err := t.f(i)
			if err != nil {
				return err
			}
			for _, ch := range t.pub.channels {
				ch <- res
			}

			return nil
		})
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
			if t.wf.outputFn != nil {
				return fmt.Errorf("%w: error composing task %s, multiple output functions", types.ErrCompose, t.name)
			}
			t.wf.outputFn = func() (O, error) {
				return t.f(<-t.sub)
			}
		} else {
			return fmt.Errorf("%w: function for %s is missing input", types.ErrCompose, t.name)
		}
	} else {
		return fmt.Errorf("%w: function for %s is nil", types.ErrCompose, t.name)
	}

	return nil
}
