package composer

import (
	"errors"
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func NewSimpleWorkflow[I, O any]() *SimpleWorkflow[I, O] {
	swf := &SimpleWorkflow[I, O]{
		pubs: make(map[string]pub),
	}
	swf.inputTask = &InputTask[I]{
		pub: SetPub[I, O, I](swf, task.Input),
	}

	return swf
}

type SimpleWorkflow[I, O any] struct {
	Tasks         []task.Task
	pubs          map[string]pub
	inputTask     *InputTask[I]
	hiddenFns     []func() error
	outputFnOrder int
	outputFn      func() (O, error)
	result        func(I) (O, error)
	errs          []error
}

func (c *SimpleWorkflow[I, O]) AddTask(task task.Task) {
	c.Tasks = append(c.Tasks, task)
}

func (c *SimpleWorkflow[I, O]) AddErr(err error) {
	c.errs = append(c.errs, err)
}

func (c *SimpleWorkflow[I, O]) Input() task.Dependency[I] {
	return task.WorkflowInput[I](task.Input)
}

func (c *SimpleWorkflow[I, O]) AddInputFn(fn func(I) error, order int) {
	// This method is still needed for other function types like BiFn, Consumer, etc.
	// They will continue to use the old approach until they are also updated
	c.hiddenFns = append(c.hiddenFns, func() error {
		return nil // placeholder - this needs to be handled differently
	})
}

func (c *SimpleWorkflow[I, O]) AddTransformFn(fn func() error) {
	c.hiddenFns = append(c.hiddenFns, fn)
}

func (c *SimpleWorkflow[I, O]) SetOutputFn(fn func() (O, error), order int) bool {
	if c.outputFn == nil {
		c.outputFnOrder = order
		c.outputFn = fn
		return true
	}

	return false
}

func (c *SimpleWorkflow[I, O]) Compose() (func(I) (O, error), error) {
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

func compose[I, O any](c *SimpleWorkflow[I, O]) error {
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
	for _, info := range c.Tasks {
		err = info.Compose()
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

func createOutputFn[I, O any](c *SimpleWorkflow[I, O]) func(I) (O, error) {
	return func(in I) (O, error) {
		var (
			o   O
			err error
		)

		// Publish input through the input task if it exists
		if c.inputTask != nil {
			for _, ch := range c.inputTask.pub.Channels {
				ch <- in
			}
		}

		// Execute all hidden functions
		for _, hiddenFn := range c.hiddenFns {
			err = hiddenFn()
			if err != nil {
				return o, err
			}
		}
		return c.outputFn()
	}
}

type InputTask[I any] struct {
	pub *PubImpl[I]
}
