package composer

type Composer[I, O any] interface {
	Compose() (func(I) (O, error), error)
}
