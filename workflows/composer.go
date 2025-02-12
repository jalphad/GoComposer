package workflows

import (
	"errors"
	"fmt"
	"github.com/jalphad/gocomposer/types"
)

func NewComposer[I, O any]() *Composer[I, O] {
	return &Composer[I, O]{
		pubs: make(map[string]pub),
	}
}

type Composer[I, O any] struct {
	tasks        []Task
	pubs         map[string]pub
	inputFns     []func(I) error
	transformFns []func() error
	outputFn     func() (O, error)
	result       func(I) (O, error)
	errs         []error
}

func (c *Composer[I, O]) Compose() (func(I) (O, error), error) {
	if c.errs != nil {
		return nil, errors.Join(c.errs...)
	}
	err := compose(c)
	if err != nil {
		return nil, err
	}
	if c.result == nil {
		return nil, fmt.Errorf("%w: resulting function is nil", types.ErrCompose)
	}

	return c.result, nil
}

func compose[I, O any](c *Composer[I, O]) error {
	if c == nil {
		return fmt.Errorf("%w: composer cannot be nil", types.ErrInvalidArgument)
	}
	var (
		err error
	)
	for k, v := range c.pubs {
		if !v.hasPublisher() {
			return fmt.Errorf("%w: task %s is not found", types.ErrCompose, k)
		}
	}
	for _, info := range c.tasks {
		err = info.compose()
		if err != nil {
			return err
		}
	}
	for k, v := range c.pubs {
		var outputTasks = make([]string, 0, 1)
		if !v.isConsumed() {
			outputTasks = append(outputTasks, k)
		}
		if len(outputTasks) > 1 {
			return fmt.Errorf("%w: multiple outputs %v", types.ErrCompose, outputTasks)
		}
	}
	c.result = createOutputFn[I, O](c)
	return nil
}

func createOutputFn[I, O any](c *Composer[I, O]) func(I) (O, error) {
	return func(s I) (O, error) {
		var (
			o   O
			err error
		)
		for _, consumer := range c.inputFns {
			err = consumer(s)
			if err != nil {
				return o, err
			}
		}
		for _, intermediate := range c.transformFns {
			err = intermediate()
			if err != nil {
				return o, err
			}
		}
		return c.outputFn()
	}
}
