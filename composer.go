package composer

import (
	"errors"
	"fmt"
)

func New[I, O any](opts ...Option[I, O]) *Composer[I, O] {
	wf := &Composer[I, O]{}
	for _, opt := range opts {
		opt(wf)
	}
	return wf
}

type Option[I, O any] func(c *Composer[I, O])

func WithFn[I, O, R, S any](name string, f func(R) S) Option[I, O] {
	return func(c *Composer[I, O]) {
		c.tasks = append(c.tasks, &taskImpl[I, O, R, S]{
			wf: c,
			f:  f,
			n:  name,
		})
	}
}

func WithErrFn[I, O, R, S any](name string, f func(R) (S, error)) Option[I, O] {
	return func(c *Composer[I, O]) {
		c.tasks = append(c.tasks, &taskImpl[I, O, R, S]{
			wf:   c,
			ferr: f,
			n:    name,
		})
	}
}

type Composer[I, O any] struct {
	tasks []task
}

func (w *Composer[I, O]) Compose() (func(I) (O, error), error) {
	r, err := compose(w)
	if err != nil {
		return nil, err
	}
	if r.f != nil {
		return func(i I) (O, error) {
			return r.f(i), nil
		}, nil
	}
	return r.ferr, nil
}

func (w *Composer[I, O]) MustCompose() func(I) (O, error) {
	r, err := w.Compose()
	if err != nil {
		panic(fmt.Errorf("composer failed to compose with error: %w", err))
	}
	return r
}

func (w *Composer[I, O]) ComposeNoErr(panicOnErr bool) (func(I) O, error) {
	rf, err := compose(w)
	if err != nil {
		return nil, err
	}
	if rf.f != nil {
		return rf.f, nil
	}
	if panicOnErr {
		return func(i I) O {
			r, err := rf.ferr(i)
			if err != nil {
				panic(err)
			}
			return r
		}, nil
	}
	return func(i I) O {
		r, _ := rf.ferr(i)
		return r
	}, nil
}

func (w *Composer[I, O]) MustComposeNoErr(panicOnErr bool) func(I) O {
	r, err := w.ComposeNoErr(panicOnErr)
	if err != nil {
		panic(fmt.Errorf("composer failed to compose with error: %w", err))
	}
	return r
}

func AddFn[I, O, R, S any](c *Composer[I, O], name string, f func(R) S) {
	c.tasks = append(c.tasks, &taskImpl[I, O, R, S]{
		wf: c,
		f:  f,
		n:  name,
	})
}

func AddErrFn[I, O, R, S any](c *Composer[I, O], name string, f func(R) (S, error)) {
	c.tasks = append(c.tasks, &taskImpl[I, O, R, S]{
		wf:   c,
		ferr: f,
		n:    name,
	})
}

type task interface {
	name() string
	compose(info task) (task, error)
}

type taskImpl[I, O, R, S any] struct {
	wf   *Composer[I, O]
	f    func(R) S
	ferr func(R) (S, error)
	n    string
}

func (t *taskImpl[I, O, R, S]) name() string {
	return t.n
}

func (t *taskImpl[I, O, R, S]) compose(info task) (task, error) {
	prev, ok := info.(*taskImpl[I, O, I, R])
	if !ok {
		return nil, errors.New("failed to compose")
	}
	next := &taskImpl[I, O, I, S]{}
	if t.f != nil {
		if prev.f != nil {
			next.f = func(i I) S {
				return t.f(prev.f(i))
			}
		} else if prev.ferr != nil {
			next.ferr = func(i I) (S, error) {
				var s S
				r, err := prev.ferr(i)
				if err != nil {
					return s, err
				}
				return t.f(r), nil
			}
		}
	}
	if t.ferr != nil {
		if prev.f != nil {
			next.ferr = func(i I) (S, error) {
				return t.ferr(prev.f(i))
			}
		} else if prev.ferr != nil {
			next.ferr = func(i I) (S, error) {
				var s S
				r, err := prev.ferr(i)
				if err != nil {
					return s, err
				}
				return t.ferr(r)
			}
		}
	}
	return next, nil
}

func Wrap[R, S any](wrapper func(func(R) S) func(R) S, fn func(R) S) func(R) S {
	return func(r R) S {
		return wrapper(fn)(r)
	}
}

func compose[I, O any](w *Composer[I, O]) (*taskImpl[I, O, I, O], error) {
	var (
		ti  task
		err error
	)
	for _, info := range w.tasks {
		if ti == nil {
			ti = info
			continue
		}
		ti, err = info.compose(ti)
		if err != nil {
			return nil, err
		}
	}
	return ti.(*taskImpl[I, O, I, O]), nil
}
