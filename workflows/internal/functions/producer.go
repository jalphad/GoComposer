package functions

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func AddProducer[I, O, S any](c composer.Composer[I, O], f func() (S, error), opts *ProducerOpts) task.Dependency[S] {
	swf := c.(*composer.SimpleWorkflow[I, O])
	opts = setProducerOpts(swf, opts)
	this := &taskProducer[I, O, S]{
		taskBase: taskBase[I, O, S]{
			wf:    swf,
			name:  opts.Name,
			order: opts.order,
		},
		f: f,
	}
	this.pub = composer.SetPub[I, O, S](swf, opts.Name)
	swf.AddTask(this)

	return (&task.TaskDependency[S]{}).SetName(this.name)
}

type ProducerOpts struct {
	Name  string
	order int
}

func setProducerOpts[I, O any](c *composer.SimpleWorkflow[I, O], o *ProducerOpts) *ProducerOpts {
	if o == nil {
		return &ProducerOpts{
			Name:  fmt.Sprintf("Task%d", len(c.Tasks)+1),
			order: len(c.Tasks),
		}
	}
	if o.Name == "" {
		o.Name = fmt.Sprintf("Task%d", len(c.Tasks)+1)
	}
	o.order = len(c.Tasks)
	return o
}

type taskProducer[I, O, S any] struct {
	taskBase[I, O, S]
	f func() (S, error)
}

func (t *taskProducer[I, O, S]) Name() string {
	return t.name
}

func (t *taskProducer[I, O, S]) Compose() error {
	if t.f != nil {
		if len(t.pub.Channels) != 0 {
			t.wf.AddTransformFn(func() error {
				res, err := t.f()
				if err != nil {
					return err
				}
				for _, ch := range t.pub.Channels {
					ch <- res
				}

				return nil
			})
		} else if to, ok := t.toOutputFn(); ok {
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

func (t *taskProducer[I, O, S]) isDependency() {}

func (t *taskProducer[I, O, S]) toOutputFn() (*taskOutputProducer[I, O, O], bool) {
	fo, ok := any(t.f).(func() (O, error))
	if !ok {
		return nil, false
	}
	return &taskOutputProducer[I, O, O]{
		taskBase: taskBase[I, O, O]{
			wf:    t.wf,
			name:  t.name,
			order: t.order,
		},
		f: fo,
	}, true
}

type taskOutputProducer[I, O, R any] taskProducer[I, O, O]

func (t *taskOutputProducer[I, O, S]) Name() string {
	return t.name
}

func (t *taskOutputProducer[I, O, R]) Compose() error {
	if t.f != nil {
		if ok := t.wf.SetOutputFn(func() (O, error) {
			return t.f()
		}, t.order); !ok {
			return fmt.Errorf("%w: error composing task %s, multiple output functions", types.ErrCompose, t.name)
		}
	} else {
		return fmt.Errorf("%w: function for %s is nil", types.ErrCompose, t.name)
	}

	return nil
}
