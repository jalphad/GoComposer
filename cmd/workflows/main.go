package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jalphad/gocomposer/workflows"
	"github.com/jalphad/gocomposer/workflows/internal/composer"
	functions2 "github.com/jalphad/gocomposer/workflows/internal/functions"
)

func main() {
	w := workflows.NewComposer[int, string]()
	t1 := workflows.AddFn(w, ToErrFn(strconv.Itoa), nil)
	t2 := functions2.AddFn(w, ToErrFn(IgnoreInt), nil)
	t3 := functions2.AddBiFn(w, Combine2Strings, functions2.NewBiFnOpts(t1, t2))
	t4 := functions2.AddFn(w, ToErrFn(DuplicateString), functions2.NewFnOpts(t3))
	t5 := functions2.AddFn(w, ToErrFn(AddBar), functions2.NewFnOpts(t4))
	functions2.AddFn(w, ToErrFn(AddBar), functions2.NewFnOpts(t5))
	//workflows.AddFn(w, strconv.Atoi, &workflows.FnOpts{Name: "t5", DependsOn: t3.Name()})
	fn, err := TimeCompose(w)
	if err != nil {
		fmt.Println("oops: " + err.Error())
		return
	}
	defer timer("func")()
	fmt.Println(fn(1))
}

func AddBar(in string) string {
	return in + "bar"
}

func DuplicateString(in string) string {
	return in + in
}

func IgnoreInt(_ int) string {
	return "Ignored the input"
}

func Combine2Strings(in1 string, in2 string) (string, error) {
	return in1 + in2, nil
}

func double(n int) int { return int(n) * 2 }

func ToErrFn[I, O any](f func(I) O) func(I) (O, error) {
	return func(i I) (O, error) {
		return f(i), nil
	}
}

func TimeCompose[I, O any](c composer.Composer[I, O]) (func(I) (O, error), error) {
	defer timer("compose")()
	return c.Compose()
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
