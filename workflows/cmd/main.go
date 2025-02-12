package main

import (
	"fmt"
	"github.com/jalphad/gocomposer/workflows"
	"strconv"
)

func main() {
	w := workflows.NewComposer[int, string]()
	t1 := workflows.AddFn(w, ToErrFn(strconv.Itoa), nil)
	t2 := workflows.AddFn(w, ToErrFn(DuplicateString), workflows.NewFnOpts[string](t1))
	t3 := workflows.AddFn(w, ToErrFn(AddBar), workflows.NewFnOpts[string](t2))
	//workflows.AddFn(w, ToErrFn(double), workflows.NewFnOpts(t3))
	workflows.AddFn(w, ToErrFn(AddBar), workflows.NewFnOpts[string](t3))
	//workflows.AddFn(w, strconv.Atoi, &workflows.FnOpts{Name: "t5", DependsOn: t3.Name()})
	fn, err := w.Compose()
	if err != nil {
		fmt.Println("oops: " + err.Error())
		return
	}
	fmt.Println(fn(1))
}

func AddBar(in string) string {
	return in + "bar"
}

func DuplicateString(in string) string {
	return in + in
}

func double(n int) int { return int(n) * 2 }

func ToErrFn[I, O any](f func(I) O) func(I) (O, error) {
	return func(i I) (O, error) {
		return f(i), nil
	}
}
