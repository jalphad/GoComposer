package composer

import (
	"errors"
	"fmt"

	"github.com/jalphad/gocomposer/types"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

func NewSimpleWorkflow[I, O any]() *SimpleWorkflow[I, O] {
	return &SimpleWorkflow[I, O]{
		pubs: make(map[string]pub),
	}
}

type SimpleWorkflow[I, O any] struct {
	Tasks         []task.Task
	pubs          map[string]pub
	inputFnOrder  []int
	inputFns      []func(I) error
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
	c.inputFnOrder = append(c.inputFnOrder, order)
	c.inputFns = append(c.inputFns, fn)
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
		j := 0
		k := 0
		for i, fn := range c.inputFns {
			// if some functions (consumer functions) are added after the output function we need to adjust
			// j by one to keep intended order of execution
			if c.inputFnOrder[i] > c.outputFnOrder {
				k = 1
			}
			for c.inputFnOrder[i]-j-k > i {
				err = c.hiddenFns[j]()
				if err != nil {
					return o, err
				}
				j++
			}
			err = fn(in)
			if err != nil {
				return o, err
			}
		}
		for _, hiddenFn := range c.hiddenFns[j:] {
			err = hiddenFn()
			if err != nil {
				return o, err
			}
		}
		return c.outputFn()
	}
}
