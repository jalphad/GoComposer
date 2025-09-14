package functions

import (
	"fmt"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddQuadFnChainedFunctions(t *testing.T) {
	// Create a workflow with QuadFunction that chains four inputs
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

	// Add fourth function: string -> int (quadruple length)
	dep4 := AddFn(swf, func(s string) (int, error) {
		return len(s) * 4, nil
	}, nil)

	// Add QuadFunction: combine all four results
	AddQuadFn(swf, func(len1, len2, len3, len4 int) (string, error) {
		total := len1 + len2 + len3 + len4
		return fmt.Sprintf("quad-sum: %d + %d + %d + %d = %d", len1, len2, len3, len4, total), nil
	}, &QuadFnOpts[int, int, int, int]{
		Input1: dep1,
		Input2: dep2,
		Input3: dep3,
		Input4: dep4,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with input "hello" (length 5, double 10, triple 15, quadruple 20, sum 50)
	input := "hello"
	expectedOutput := "quad-sum: 5 + 10 + 15 + 20 = 50"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddQuadFnWithWorkflowInputs(t *testing.T) {
	// Test QuadFunction where some inputs come from workflow input directly
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add function: string -> bool (is longer than 3)
	dep2 := AddFn(swf, func(s string) (bool, error) {
		return len(s) > 3, nil
	}, nil)

	// Add QuadFunction: workflow input + workflow input + length + bool
	AddQuadFn(swf, func(s1, s2 string, length int, isLong bool) (string, error) {
		longStr := "short"
		if isLong {
			longStr = "long"
		}
		return fmt.Sprintf("%s-%s (len:%d, %s)", s1, s2, length, longStr), nil
	}, &QuadFnOpts[string, string, int, bool]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input
		Input3: dep1,
		Input4: dep2,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	expectedOutput := "test-test (len:4, long)"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddQuadFnAllWorkflowInputs(t *testing.T) {
	// Test QuadFunction where all inputs come from workflow input
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add QuadFunction: workflow input used four times
	AddQuadFn(swf, func(s1, s2, s3, s4 string) (string, error) {
		return fmt.Sprintf("quad: %s + %s + %s + %s", s1, s2, s3, s4), nil
	}, &QuadFnOpts[string, string, string, string]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input
		Input3: nil, // workflow input
		Input4: nil, // workflow input
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "hello"
	expectedOutput := "quad: hello + hello + hello + hello"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddQuadFnMixedInputs(t *testing.T) {
	// Test QuadFunction with mixed input sources
	swf := composer.NewSimpleWorkflow[int, string]()

	// Add function: int -> string
	dep1 := AddFn(swf, func(i int) (string, error) {
		return fmt.Sprintf("num-%d", i), nil
	}, nil)

	// Add function: int -> bool
	dep2 := AddFn(swf, func(i int) (bool, error) {
		return i%2 == 0, nil
	}, nil)

	// Add function: int -> float64
	dep3 := AddFn(swf, func(i int) (float64, error) {
		return float64(i) / 2.0, nil
	}, nil)

	// Add QuadFunction: workflow input + string + bool + float64
	AddQuadFn(swf, func(original int, processed string, isEven bool, half float64) (string, error) {
		evenStr := "odd"
		if isEven {
			evenStr = "even"
		}
		return fmt.Sprintf("original: %d, processed: %s, is %s, half: %.1f", original, processed, evenStr, half), nil
	}, &QuadFnOpts[int, string, bool, float64]{
		Input1: nil, // workflow input
		Input2: dep1,
		Input3: dep2,
		Input4: dep3,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := 42
	expectedOutput := "original: 42, processed: num-42, is even, half: 21.0"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}