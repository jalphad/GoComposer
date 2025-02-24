package workflows

import (
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	"github.com/jalphad/gocomposer/workflows/types"
)

func NewComposer[I, O any]() composer.Composer[I, O] {
	return composer.NewSimpleWorkflow[I, O]()
}

func Fn[I, O, S, T any](c composer.Composer[I, O], f func(S) (T, error)) *types.Function[I, O, S, T] {
	return types.NewFunction(c, f)
}

func BiFn[I, O, R, S, T any](c composer.Composer[I, O], f func(R, S) (T, error)) *types.BiFunction[I, O, R, S, T] {
	return types.NewBiFunction(c, f)
}

func TriFn[I, O, Q, R, S, T any](c composer.Composer[I, O], f func(Q, R, S) (T, error)) *types.TriFunction[I, O, Q, R, S, T] {
	return types.NewTriFunction(c, f)
}

func Producer[I, O, T any](c composer.Composer[I, O], f func() (T, error)) *types.Producer[I, O, T] {
	return types.NewProducer(c, f)
}

func Consumer[I, O, S any](c composer.Composer[I, O], f func(S) error) *types.Consumer[I, O, S] {
	return types.NewConsumer(c, f)
}

func BiConsumer[I, O, R, S any](c composer.Composer[I, O], f func(R, S) error) *types.BiConsumer[I, O, R, S] {
	return types.NewBiConsumer(c, f)
}
