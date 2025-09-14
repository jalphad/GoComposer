package functions

import (
	"fmt"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddQuintFnChainedFunctions(t *testing.T) {
	// Create a workflow with QuintFunction that chains five inputs
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

	// Add fifth function: string -> int (quintuple length)
	dep5 := AddFn(swf, func(s string) (int, error) {
		return len(s) * 5, nil
	}, nil)

	// Add QuintFunction: combine all five results
	AddQuintFn(swf, func(len1, len2, len3, len4, len5 int) (string, error) {
		total := len1 + len2 + len3 + len4 + len5
		return fmt.Sprintf("quint-sum: %d + %d + %d + %d + %d = %d", len1, len2, len3, len4, len5, total), nil
	}, &QuintFnOpts[int, int, int, int, int]{
		Input1: dep1,
		Input2: dep2,
		Input3: dep3,
		Input4: dep4,
		Input5: dep5,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with input "hello" (length 5, double 10, triple 15, quadruple 20, quintuple 25, sum 75)
	input := "hello"
	expectedOutput := "quint-sum: 5 + 10 + 15 + 20 + 25 = 75"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddQuintFnWithWorkflowInputs(t *testing.T) {
	// Test QuintFunction where some inputs come from workflow input directly
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add function: string -> bool (is longer than 3)
	dep2 := AddFn(swf, func(s string) (bool, error) {
		return len(s) > 3, nil
	}, nil)

	// Add function: string -> rune (first character)
	dep3 := AddFn(swf, func(s string) (rune, error) {
		if len(s) == 0 {
			return 0, fmt.Errorf("empty string")
		}
		return rune(s[0]), nil
	}, nil)

	// Add QuintFunction: workflow input + workflow input + length + bool + rune
	AddQuintFn(swf, func(s1, s2 string, length int, isLong bool, firstChar rune) (string, error) {
		longStr := "short"
		if isLong {
			longStr = "long"
		}
		return fmt.Sprintf("%s-%s (len:%d, %s, first:'%c')", s1, s2, length, longStr, firstChar), nil
	}, &QuintFnOpts[string, string, int, bool, rune]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input
		Input3: dep1,
		Input4: dep2,
		Input5: dep3,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	expectedOutput := "test-test (len:4, long, first:'t')"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddQuintFnAllWorkflowInputs(t *testing.T) {
	// Test QuintFunction where all inputs come from workflow input
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add QuintFunction: workflow input used five times
	AddQuintFn(swf, func(s1, s2, s3, s4, s5 string) (string, error) {
		return fmt.Sprintf("quint: %s + %s + %s + %s + %s", s1, s2, s3, s4, s5), nil
	}, &QuintFnOpts[string, string, string, string, string]{
		Input1: nil, // workflow input
		Input2: nil, // workflow input
		Input3: nil, // workflow input
		Input4: nil, // workflow input
		Input5: nil, // workflow input
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "hello"
	expectedOutput := "quint: hello + hello + hello + hello + hello"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}

func TestAddQuintFnMixedInputs(t *testing.T) {
	// Test QuintFunction with mixed input sources
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

	// Add function: int -> byte
	dep4 := AddFn(swf, func(i int) (byte, error) {
		return byte(i % 256), nil
	}, nil)

	// Add QuintFunction: workflow input + string + bool + float64 + byte
	AddQuintFn(swf, func(original int, processed string, isEven bool, half float64, b byte) (string, error) {
		evenStr := "odd"
		if isEven {
			evenStr = "even"
		}
		return fmt.Sprintf("original: %d, processed: %s, is %s, half: %.1f, byte: %d", original, processed, evenStr, half, b), nil
	}, &QuintFnOpts[int, string, bool, float64, byte]{
		Input1: nil, // workflow input
		Input2: dep1,
		Input3: dep2,
		Input4: dep3,
		Input5: dep4,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := 42
	expectedOutput := "original: 42, processed: num-42, is even, half: 21.0, byte: 42"
	actualOutput, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	if actualOutput != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, actualOutput)
	}
}