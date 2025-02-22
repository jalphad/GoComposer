package workflows

import (
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/functions"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

type (
	Composer[I, O any] = composer.Composer[I, O]
)

func NewComposer[I, O any]() Composer[I, O] {
	return composer.NewSimpleWorkflow[I, O]()
}

func Fn[I, O, R, S any](c composer.Composer[I, O], f func(R) (S, error)) *Function[I, O, R, S] {
	return &Function[I, O, R, S]{
		c: c,
		f: f,
	}
}

func BiFn[I, O, Q, R, S any](c composer.Composer[I, O], f func(Q, R) (S, error)) *BiFunction[I, O, Q, R, S] {
	return &BiFunction[I, O, Q, R, S]{
		c: c,
		f: f,
	}
}

type Function[I, O, R, S any] struct {
	c composer.Composer[I, O]
	f func(R) (S, error)
	d task.Dependency[R]
}

func (f *Function[I, O, R, S]) Param(d task.Dependency[R]) *Function[I, O, R, S] {
	if f == nil {
		return nil
	}
	f.d = d
	return f
}

func (f *Function[I, O, R, S]) Add() task.Dependency[S] {
	opts := &functions.FnOpts[R]{
		Input: f.d,
	}
	return functions.AddFn(f.c, f.f, opts)
}

type BiFunction[I, O, Q, R, S any] struct {
	c  composer.Composer[I, O]
	f  func(Q, R) (S, error)
	d1 task.Dependency[Q]
	d2 task.Dependency[R]
}

func (f *BiFunction[I, O, Q, R, S]) Params(d1 task.Dependency[Q], d2 task.Dependency[R]) *BiFunction[I, O, Q, R, S] {
	if f == nil {
		return nil
	}
	f.d1 = d1
	f.d2 = d2
	return f
}

func (f *BiFunction[I, O, Q, R, S]) Add() task.Dependency[S] {
	opts := &functions.BiFnOpts[Q, R]{
		Input1: f.d1,
		Input2: f.d2,
	}
	return functions.AddBiFn(f.c, f.f, opts)
}
