package composer

import (
	"fmt"

	"github.com/jalphad/gocomposer/types"
)

type pub interface {
	isConsumed() bool
	hasPublisher() bool
}

type PubImpl[T any] struct {
	Channels []chan<- T
	claimed  bool
}

func (b *PubImpl[T]) isConsumed() bool {
	return len(b.Channels) > 0
}

func (b *PubImpl[T]) hasPublisher() bool {
	return b.claimed
}

func AddSub[I, O, T any](c *SimpleWorkflow[I, O], name string, ch chan<- T) {
	if got, ok := c.pubs[name]; ok {
		if impl, ok := got.(*PubImpl[T]); ok {
			impl.Channels = append(impl.Channels, ch)
		} else {
			c.errs = append(c.errs, fmt.Errorf("%w: adding subscription for %s, subscription was not for type %T", types.ErrCompose, name, ch))
		}
	} else {
		p := &PubImpl[T]{
			Channels: make([]chan<- T, 0),
		}
		p.Channels = append(p.Channels, ch)
		c.pubs[name] = p
	}
}

func SetPub[I, O, T any](c *SimpleWorkflow[I, O], name string) *PubImpl[T] {
	if got, ok := c.pubs[name]; ok {
		if gota, ok := got.(*PubImpl[T]); ok {
			gota.claimed = true
			return gota
		} else {
			var t T
			c.errs = append(c.errs, fmt.Errorf("%w: claiming pub for %s but pub was not of type %T", types.ErrCompose, name, t))
		}
	}
	newpub := &PubImpl[T]{
		claimed:  true,
		Channels: make([]chan<- T, 0),
	}
	c.pubs[name] = newpub

	return newpub
}
