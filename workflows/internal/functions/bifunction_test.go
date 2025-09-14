package functions

import (
	"fmt"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddBiFnChainedFunctions(t *testing.T) {
	// Create a workflow with BiFunction that chains two inputs
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add first function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add second function: string -> int (parse length - simplified)
	dep2 := AddFn(swf, func(s string) (int, error) {
		return len(s) * 2, nil // just double the length as an example
	}, nil)

	// Add BiFunction: combine both results
	AddBiFn(swf, func(len1, len2 int) (string, error) {
		return fmt.Sprintf("combined: %d + %d = %d", len1, len2, len1+len2), nil
	}, &BiFnOpts[int, int]{
		Input1: dep1,
		Input2: dep2,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with input "hello" (length 5, double length 10, sum 15)
	input := "hello"
	expectedOutput := "combined: 5 + 10 = 15"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddBiFnWithWorkflowInput(t *testing.T) {
	// Test BiFunction where one input comes from workflow input directly
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add BiFunction: workflow input + length
	AddBiFn(swf, func(original string, length int) (string, error) {
		return fmt.Sprintf("%s (length: %d)", original, length), nil
	}, &BiFnOpts[string, int]{
		Input1: nil, // This should default to workflow input
		Input2: dep1,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	expectedOutput := "test (length: 4)"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddBiFnBothWorkflowInputs(t *testing.T) {
	// Test BiFunction where both inputs come from workflow input
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add BiFunction: workflow input + workflow input (same input used twice)
	AddBiFn(swf, func(s1, s2 string) (string, error) {
		return fmt.Sprintf("double: %s + %s", s1, s2), nil
	}, &BiFnOpts[string, string]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input (should default to workflow input)
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "hello"
	expectedOutput := "double: hello + hello"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}