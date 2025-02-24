package types

import (
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/internal/functions"
	"github.com/jalphad/gocomposer/workflows/internal/task"
)

type (
	Composer[I, O any] = composer.Composer[I, O]
)

func NewFunction[I, O, S, T any](c Composer[I, O], f func(S) (T, error)) *Function[I, O, S, T] {
	return &Function[I, O, S, T]{
		c: c,
		f: f,
	}
}

type Function[I, O, S, T any] struct {
	c Composer[I, O]
	f func(S) (T, error)
	d task.Dependency[S]
}

func (f *Function[I, O, S, T]) Param(d task.Dependency[S]) *Function[I, O, S, T] {
	if f == nil {
		return nil
	}
	f.d = d
	return f
}

func (f *Function[I, O, S, T]) Add() task.Dependency[T] {
	opts := &functions.FnOpts[S]{
		Input: f.d,
	}
	return functions.AddFn(f.c, f.f, opts)
}

func NewBiFunction[I, O, R, S, T any](c Composer[I, O], f func(R, S) (T, error)) *BiFunction[I, O, R, S, T] {
	return &BiFunction[I, O, R, S, T]{
		c: c,
		f: f,
	}
}

type BiFunction[I, O, R, S, T any] struct {
	c  Composer[I, O]
	f  func(R, S) (T, error)
	d1 task.Dependency[R]
	d2 task.Dependency[S]
}

func (f *BiFunction[I, O, R, S, T]) Params(d1 task.Dependency[R], d2 task.Dependency[S]) *BiFunction[I, O, R, S, T] {
	if f == nil {
		return nil
	}
	f.d1 = d1
	f.d2 = d2

	return f
}

func (f *BiFunction[I, O, R, S, T]) Add() task.Dependency[T] {
	opts := &functions.BiFnOpts[R, S]{
		Input1: f.d1,
		Input2: f.d2,
	}
	return functions.AddBiFn(f.c, f.f, opts)
}

func NewTriFunction[I, O, Q, R, S, T any](c Composer[I, O], f func(Q, R, S) (T, error)) *TriFunction[I, O, Q, R, S, T] {
	return &TriFunction[I, O, Q, R, S, T]{
		c: c,
		f: f,
	}
}

type TriFunction[I, O, Q, R, S, T any] struct {
	c  Composer[I, O]
	f  func(Q, R, S) (T, error)
	d1 task.Dependency[Q]
	d2 task.Dependency[R]
	d3 task.Dependency[S]
}

func (f *TriFunction[I, O, Q, R, S, T]) Params(d1 task.Dependency[Q], d2 task.Dependency[R], d3 task.Dependency[S]) *TriFunction[I, O, Q, R, S, T] {
	if f == nil {
		return nil
	}
	f.d1 = d1
	f.d2 = d2
	f.d3 = d3

	return f
}

func (f *TriFunction[I, O, Q, R, S, T]) Add() task.Dependency[T] {
	opts := &functions.TriFnOpts[Q, R, S]{
		Input1: f.d1,
		Input2: f.d2,
		Input3: f.d3,
	}
	return functions.AddTriFn(f.c, f.f, opts)
}

func NewProducer[I, O, T any](c Composer[I, O], f func() (T, error)) *Producer[I, O, T] {
	return &Producer[I, O, T]{
		c: c,
		f: f,
	}
}

type Producer[I, O, T any] struct {
	c Composer[I, O]
	f func() (T, error)
}

func (f *Producer[I, O, T]) Add() task.Dependency[T] {
	return functions.AddProducer(f.c, f.f, &functions.ProducerOpts{})
}

func NewConsumer[I, O, S any](c Composer[I, O], f func(S) error) *Consumer[I, O, S] {
	return &Consumer[I, O, S]{
		c: c,
		f: f,
	}
}

type Consumer[I, O, S any] struct {
	c Composer[I, O]
	f func(S) error
	d task.Dependency[S]
}

func (f *Consumer[I, O, S]) Param(d task.Dependency[S]) *Consumer[I, O, S] {
	if f == nil {
		return nil
	}
	f.d = d
	return f
}

func (f *Consumer[I, O, S]) Add() {
	opts := &functions.ConsmrOpts[S]{
		Input: f.d,
	}
	functions.AddConsumer(f.c, f.f, opts)
}

func NewBiConsumer[I, O, R, S any](c Composer[I, O], f func(R, S) error) *BiConsumer[I, O, R, S] {
	return &BiConsumer[I, O, R, S]{
		c: c,
		f: f,
	}
}

type BiConsumer[I, O, R, S any] struct {
	c  Composer[I, O]
	f  func(R, S) error
	d1 task.Dependency[R]
	d2 task.Dependency[S]
}

func (f *BiConsumer[I, O, R, S]) Params(d1 task.Dependency[R], d2 task.Dependency[S]) *BiConsumer[I, O, R, S] {
	if f == nil {
		return nil
	}
	f.d1 = d1
	f.d2 = d2
	return f
}

func (f *BiConsumer[I, O, R, S]) Add() {
	opts := &functions.BiConsmrOpts[R, S]{
		Input1: f.d1,
		Input2: f.d2,
	}
	functions.AddBiConsumer(f.c, f.f, opts)
}
