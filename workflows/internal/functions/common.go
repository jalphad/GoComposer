package functions

import "github.com/jalphad/gocomposer/workflows/internal/composer"

type taskBase[I, O, S any] struct {
	name  string
	order int
	wf    *composer.SimpleWorkflow[I, O]
	pub   *composer.PubImpl[S]
}
