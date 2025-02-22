package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jalphad/gocomposer/workflows"
)

func main() {
	w := workflows.NewComposer[int, string]()
	t1 := workflows.Fn(w, ToErrFn(strconv.Itoa)).Add()
	t2 := workflows.Fn(w, ToErrFn(IgnoreInt)).Add()
	t3 := workflows.BiFn(w, Combine2Strings).Params(t1, t2).Add()
	t4 := workflows.Fn(w, ToErrFn(DuplicateString)).Param(t3).Add()
	t5 := workflows.Fn(w, ToErrFn(AddBar)).Param(t4).Add()
	workflows.Fn(w, ToErrFn(AddBar)).Param(t5).Add()
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

func TimeCompose[I, O any](c workflows.Composer[I, O]) (func(I) (O, error), error) {
	defer timer("compose")()
	return c.Compose()
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
