package workflows

import (
	"fmt"
	"github.com/jalphad/gocomposer/types"
)

type pub interface {
	isConsumed() bool
	hasPublisher() bool
}

type pubImpl[T any] struct {
	channels []chan<- T
	claimed  bool
}

func (b *pubImpl[T]) isConsumed() bool {
	return len(b.channels) > 0
}

func (b *pubImpl[T]) hasPublisher() bool {
	return b.claimed
}

func addSub[I, O, T any](c *Composer[I, O], name string, ch chan<- T) {
	if got, ok := c.pubs[name]; ok {
		if impl, ok := got.(*pubImpl[T]); ok {
			impl.channels = append(impl.channels, ch)
		} else {
			c.errs = append(c.errs, fmt.Errorf("%w: adding subscription for %s, subscription was not for type %T", types.ErrCompose, name, ch))
		}
	} else {
		p := &pubImpl[T]{
			channels: make([]chan<- T, 0),
		}
		p.channels = append(p.channels, ch)
		c.pubs[name] = p
	}
}

func setPub[I, O, T any](c *Composer[I, O], name string) *pubImpl[T] {
	if got, ok := c.pubs[name]; ok {
		if gota, ok := got.(*pubImpl[T]); ok {
			gota.claimed = true
			return gota
		} else {
			var t T
			c.errs = append(c.errs, fmt.Errorf("%w: claiming pub for %s but pub was not of type %T", types.ErrCompose, name, t))
		}
	}
	newpub := &pubImpl[T]{
		claimed:  true,
		channels: make([]chan<- T, 0),
	}
	c.pubs[name] = newpub

	return newpub
}
