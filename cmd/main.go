package main

import (
	"errors"
	"fmt"
	composer "github.com/jalphad/gocomposer"
	"log"
	"strconv"
)

func main() {
	c := composer.New[int, string]()
	composer.AddFn(c, "addOne", composer.Wrap(logVal, addOne))
	composer.AddFn(c, "double", double)
	composer.AddErrFn(c, "error", errFunc[int])
	composer.AddFn(c, "toString", toString)
	f, err := c.Compose()
	if err != nil {
		log.Fatal("functions failed to compose")
	}
	fmt.Println(f(8))

	css := composer.New[string, string]()
	composer.New[string, string](
		composer.WithErrFn(css, "Atoi", strconv.Atoi),
		composer.WithFn(css, "double", composer.Wrap(logVal, double)),
		composer.WithFn(css, "Itoa", strconv.Itoa)).
		MustComposeNoErr(false)("2")
}

func addOne(n int) int      { return n + 1 }
func double(n int) int      { return int(n) * 2 }
func toString(n int) string { return strconv.Itoa(n) }
func errFunc[T any](n T) (T, error) {
	return n, errors.New("an error occurred")
}

func logVal[T any](f func(T) T) func(T) T {
	return func(t T) T {
		log.Println("value before: ", t)
		defer func() { log.Println("value after: ", t) }()
		t = f(t)
		return t
	}
}
