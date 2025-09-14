package functions

import (
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddFnChainedFunctions(t *testing.T) {
	// Create a new SimpleWorkflow that takes a string input and produces an int output
	swf := composer.NewSimpleWorkflow[string, int]()

	// Add first function: string -> int (length of string)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add second function: int -> int (multiply by 2)
	AddFn(swf, func(length int) (int, error) {
		return length * 2, nil
	}, &FnOpts[int]{
		Input: dep1,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with input "hello" (length 5, should return 10)
	input := "hello"
	expectedOutput := 10
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %d, got %d", expectedOutput, actualOutput)
	}
}

func TestAddFnInputFunction(t *testing.T) {
	// Create a workflow that processes input directly and produces output
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add function that takes workflow input, transforms it, and produces workflow output
	AddFn(swf, func(s string) (string, error) {
		return "processed: " + s, nil
	}, nil) // nil means it uses workflow input and produces workflow output

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	expectedOutput := "processed: test"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}
