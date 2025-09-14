package functions

import (
	"fmt"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddTriFnChainedFunctions(t *testing.T) {
	// Create a workflow with TriFunction that chains three inputs
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add first function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add second function: string -> int (double length)
	dep2 := AddFn(swf, func(s string) (int, error) {
		return len(s) * 2, nil
	}, nil)

	// Add third function: string -> int (triple length)
	dep3 := AddFn(swf, func(s string) (int, error) {
		return len(s) * 3, nil
	}, nil)

	// Add TriFunction: combine all three results
	AddTriFn(swf, func(len1, len2, len3 int) (string, error) {
		total := len1 + len2 + len3
		return fmt.Sprintf("tri-sum: %d + %d + %d = %d", len1, len2, len3, total), nil
	}, &TriFnOpts[int, int, int]{
		Input1: dep1,
		Input2: dep2,
		Input3: dep3,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with input "hello" (length 5, double 10, triple 15, sum 30)
	input := "hello"
	expectedOutput := "tri-sum: 5 + 10 + 15 = 30"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddTriFnWithWorkflowInputs(t *testing.T) {
	// Test TriFunction where some inputs come from workflow input directly
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add TriFunction: workflow input + workflow input + length
	AddTriFn(swf, func(s1, s2 string, length int) (string, error) {
		return fmt.Sprintf("%s-%s (len:%d)", s1, s2, length), nil
	}, &TriFnOpts[string, string, int]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input
		Input3: dep1,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	expectedOutput := "test-test (len:4)"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddTriFnAllWorkflowInputs(t *testing.T) {
	// Test TriFunction where all inputs come from workflow input
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add TriFunction: workflow input used three times
	AddTriFn(swf, func(s1, s2, s3 string) (string, error) {
		return fmt.Sprintf("triple: %s + %s + %s", s1, s2, s3), nil
	}, &TriFnOpts[string, string, string]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input
		Input3: nil, // workflow input
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "hello"
	expectedOutput := "triple: hello + hello + hello"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddTriFnMixedInputs(t *testing.T) {
	// Test TriFunction with mixed input sources
	swf := composer.NewSimpleWorkflow[int, string]()

	// Add function: int -> string
	dep1 := AddFn(swf, func(i int) (string, error) {
		return fmt.Sprintf("num-%d", i), nil
	}, nil)

	// Add function: int -> bool
	dep2 := AddFn(swf, func(i int) (bool, error) {
		return i%2 == 0, nil
	}, nil)

	// Add TriFunction: workflow input + string + bool
	AddTriFn(swf, func(original int, processed string, isEven bool) (string, error) {
		evenStr := "odd"
		if isEven {
			evenStr = "even"
		}
		return fmt.Sprintf("original: %d, processed: %s, is %s", original, processed, evenStr), nil
	}, &TriFnOpts[int, string, bool]{
		Input1: nil, // workflow input
		Input2: dep1,
		Input3: dep2,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := 42
	expectedOutput := "original: 42, processed: num-42, is even"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}