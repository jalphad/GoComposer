package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/jalphad/GoComposer.git"
)

func main() {
	c := composer.New[int, string]()
	composer.WithFn(c, "addOne", composer.Wrap(logVal, addOne))
	composer.WithFn(c, "double", double)
	composer.WithErrFn(c, "error", errFunc[int])
	composer.WithFn(c, "toString", toString)
	f, err := c.Compose()
	if err != nil {
		log.Fatal("functions failed to compose")
	}
	fmt.Print(f(8))
}

func addOne(n int) int { return n + 1 }
func double(n int) int { return int(n) * 2 }
func logVal[T any](f func(T) T) func(T) T {
	return func(t T) T {
		log.Println("value before: ", t)
		defer func() { log.Println("value after: ", t) }()
		t = f(t)
		return t
	}
}
func toString(n int) string { return strconv.Itoa(n) }
func errFunc[T any](n T) (T, error) {
	return n, errors.New("an error occurred")
}
