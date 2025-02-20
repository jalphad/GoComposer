package workflows

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
)

func AddBiFn[I, O, Q, R, S any](c *Composer[I, O], f func(Q, R) (S, error), opts *BiFnOpts[Q, R]) Dependency[S] {
	opts = setBiFnOpts(c, opts)
	if _, ok := opts.DependsOn1st.(workflowInput[Q]); ok {
		this := &taskInputFirstBiFn[I, O, R, S]{
			wf:   c,
			name: opts.Name,
		}
		if fi, ok := any(f).(func(I, R) (S, error)); ok {
			this.f = fi
		} else {
			var dummy func(I) S
			c.errs = append(c.errs, fmt.Errorf("%w: function was not of expected type, expected %T, got %T", types.ErrCompose, dummy, f))
		}

		this.pub = setPub[I, O, S](c, opts.Name)
		c.tasks = append(c.tasks, this)

		return &dependency[S]{name: this.name}
	}
	this := &taskBiFn[I, O, Q, R, S]{
		wf:   c,
		f:    f,
		name: opts.Name,
	}
	sub1Ch := make(chan Q, 1)
	addSub(c, opts.DependsOn1st.Name(), sub1Ch)
	this.sub1 = sub1Ch
	sub2Ch := make(chan R, 1)
	addSub(c, opts.DependsOn2nd.Name(), sub2Ch)
	this.sub2 = sub2Ch
	this.pub = setPub[I, O, S](c, opts.Name)
	c.tasks = append(c.tasks, this)

	return &dependency[S]{name: this.name}
}

func NewBiFnOpts[Q, R any](dependsOn1st Dependency[Q], dependsOn2nd Dependency[R]) *BiFnOpts[Q, R] {
	return &BiFnOpts[Q, R]{
		DependsOn1st: dependsOn1st,
		DependsOn2nd: dependsOn2nd,
	}
}

type BiFnOpts[Q, R any] struct {
	Name         string
	DependsOn1st Dependency[Q]
	DependsOn2nd Dependency[R]
}

func setBiFnOpts[I, O, Q, R any](c *Composer[I, O], o *BiFnOpts[Q, R]) *BiFnOpts[Q, R] {
	if o == nil {
		return &BiFnOpts[Q, R]{
			Name:         fmt.Sprintf("Task%d", len(c.tasks)+1),
			DependsOn1st: workflowInput[Q](Input),
			DependsOn2nd: workflowInput[R](Input),
		}
	} else if o.DependsOn1st == nil {
		return &BiFnOpts[Q, R]{
			Name:         fmt.Sprintf("Task%d", len(c.tasks)+1),
			DependsOn1st: workflowInput[Q](Input),
		}
	} else if o.DependsOn2nd == nil {
		return &BiFnOpts[Q, R]{
			Name:         fmt.Sprintf("Task%d", len(c.tasks)+1),
			DependsOn2nd: workflowInput[R](Input),
		}
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.tasks)+1)
	}
	return o
}

type taskBiFn[I, O, Q, R, S any] struct {
	name string
	wf   *Composer[I, O]
	f    func(Q, R) (S, error)
	pub  *pubImpl[S]
	sub1 <-chan Q
	sub2 <-chan R
}

func (t *taskBiFn[I, O, Q, R, S]) Name() string {
	return t.name
}

func (t *taskBiFn[I, O, Q, R, S]) compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil && len(t.pub.channels) != 0 {
			t.wf.transformFns = append(t.wf.transformFns, func() error {
				res, err := t.f(<-t.sub1, <-t.sub2)
				if err != nil {
					return err
				}
				for _, ch := range t.pub.channels {
					ch <- res
				}

				return nil
			})
		} else if to, ok := t.toOutputBiFn(); ok {
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

func (t *taskBiFn[I, O, Q, R, S]) isDependency() {}

func (t *taskBiFn[I, O, Q, R, S]) toOutputBiFn() (*taskOutputBiFn[I, O, Q, R], bool) {
	fo, ok := any(t.f).(func(Q, R) (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputBiFn[I, O, Q, R]{
		name: t.name,
		wf:   t.wf,
		f:    fo,
		sub1: t.sub1,
		sub2: t.sub2,
	}, true
}

type taskInputFirstBiFn[I, O, R, S any] taskBiFn[I, O, I, R, S]

func (t *taskInputFirstBiFn[I, O, R, S]) Name() string {
	return t.name
}

func (t *taskInputFirstBiFn[I, O, R, S]) compose() error {
	if t.f != nil && len(t.pub.channels) != 0 {
		t.wf.inputFns = append(t.wf.inputFns, func(i I) error {
			res, err := t.f(i, <-t.sub2)
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

func (t *taskInputFirstBiFn[I, O, R, S]) isDependency() {}

type taskOutputBiFn[I, O, Q, R any] taskBiFn[I, O, Q, R, O]

func (t *taskOutputBiFn[I, O, Q, R]) Name() string {
	return t.name
}

func (t *taskOutputBiFn[I, O, Q, R]) compose() error {
	if t.f != nil {
		if t.sub1 != nil && t.sub2 != nil {
			if t.wf.outputFn != nil {
				return fmt.Errorf("%w: error composing task %s, multiple output functions", types.ErrCompose, t.name)
			}
			t.wf.outputFn = func() (O, error) {
				return t.f(<-t.sub1, <-t.sub2)
			}
		} else {
			return fmt.Errorf("%w: function for %s is missing input", types.ErrCompose, t.name)
		}
	} else {
		return fmt.Errorf("%w: function for %s is nil", types.ErrCompose, t.name)
	}

	return nil
}
