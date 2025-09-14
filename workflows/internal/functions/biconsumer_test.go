package functions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddBiConsumerWithWorkflowInputs(t *testing.T) {
	// Test BiConsumer that takes workflow input for both parameters
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed string

	// Add BiConsumer: consume workflow input twice
	AddBiConsumer(swf, func(s1, s2 string) error {
		consumed = fmt.Sprintf("consumed: %s + %s", s1, s2)
		return nil
	}, nil)

	// Add output function to return some result
	AddFn(swf, func(s string) (string, error) {
		return fmt.Sprintf("processed: %s", s), nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	// Check that biconsumer was called
	expectedConsumed := "consumed: test + test"
	if consumed != expectedConsumed {
		t.Errorf("Expected consumed %q, got %q", expectedConsumed, consumed)
	}

	// Check output
	expectedOutput := "processed: test"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddBiConsumerWithFunctionDependencies(t *testing.T) {
	// Test BiConsumer that takes outputs from other functions
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed string

	// Add first function: string -> int (length)
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add second function: string -> bool (is longer than 3)
	dep2 := AddFn(swf, func(s string) (bool, error) {
		return len(s) > 3, nil
	}, nil)

	// Add BiConsumer: consume both dependencies
	AddBiConsumer(swf, func(length int, isLong bool) error {
		longStr := "short"
		if isLong {
			longStr = "long"
		}
		consumed = fmt.Sprintf("consumed: length=%d, %s", length, longStr)
		return nil
	}, &BiConsmrOpts[int, bool]{
		Input1: dep1,
		Input2: dep2,
	})

	// Add output function to return some result
	AddFn(swf, func(s string) (string, error) {
		return fmt.Sprintf("processed: %s", s), nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "hello"
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	// Check that biconsumer was called
	expectedConsumed := "consumed: length=5, long"
	if consumed != expectedConsumed {
		t.Errorf("Expected consumed %q, got %q", expectedConsumed, consumed)
	}

	// Check output
	expectedOutput := "processed: hello"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddBiConsumerMixedInputs(t *testing.T) {
	// Test BiConsumer with one workflow input and one function dependency
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed string

	// Add function: string -> int (length)
	dep := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add BiConsumer: workflow input + function dependency
	AddBiConsumer(swf, func(original string, length int) error {
		consumed = fmt.Sprintf("consumed: '%s' has %d chars", original, length)
		return nil
	}, &BiConsmrOpts[string, int]{
		Input1: nil, // workflow input
		Input2: dep,
	})

	// Add output function
	AddFn(swf, func(s string) (string, error) {
		return fmt.Sprintf("result: %s", s), nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "test"
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	// Check that biconsumer was called
	expectedConsumed := "consumed: 'test' has 4 chars"
	if consumed != expectedConsumed {
		t.Errorf("Expected consumed %q, got %q", expectedConsumed, consumed)
	}

	expectedOutput := "result: test"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddBiConsumerMultiple(t *testing.T) {
	// Test multiple BiConsumers
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed1, consumed2 string

	// Add functions
	dep1 := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	dep2 := AddFn(swf, func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}, nil)

	// Add first BiConsumer: workflow input + length
	AddBiConsumer(swf, func(original string, length int) error {
		consumed1 = fmt.Sprintf("biconsumer1: %s (%d)", original, length)
		return nil
	}, &BiConsmrOpts[string, int]{
		Input1: nil, // workflow input
		Input2: dep1,
	})

	// Add second BiConsumer: length + uppercase
	AddBiConsumer(swf, func(length int, upper string) error {
		consumed2 = fmt.Sprintf("biconsumer2: %d chars in %s", length, upper)
		return nil
	}, &BiConsmrOpts[int, string]{
		Input1: dep1,
		Input2: dep2,
	})

	// Add output function
	AddFn(swf, func(s string) (string, error) {
		return fmt.Sprintf("final: %s", s), nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "hello"
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	// Check that both biconsumers were called
	expectedConsumed1 := "biconsumer1: hello (5)"
	if consumed1 != expectedConsumed1 {
		t.Errorf("Expected consumed1 %q, got %q", expectedConsumed1, consumed1)
	}

	expectedConsumed2 := "biconsumer2: 5 chars in HELLO"
	if consumed2 != expectedConsumed2 {
		t.Errorf("Expected consumed2 %q, got %q", expectedConsumed2, consumed2)
	}

	expectedOutput := "final: hello"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddBiConsumerWithError(t *testing.T) {
	// Test BiConsumer that returns an error
	swf := composer.NewSimpleWorkflow[int, string]()

	// Add BiConsumer that fails under certain conditions
	AddBiConsumer(swf, func(num1, num2 int) error {
		if num1+num2 > 10 {
			return fmt.Errorf("biconsumer error: sum %d too large", num1+num2)
		}
		return nil
	}, nil)

	// Add output function
	AddFn(swf, func(i int) (string, error) {
		return fmt.Sprintf("result: %d", i), nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with successful input (5 + 5 = 10, not > 10)
	output, err := composedFn(5)
	if err != nil {
		t.Fatalf("Composed function returned unexpected error: %v", err)
	}
	if output != "result: 5" {
		t.Errorf("Expected output 'result: 5', got %q", output)
	}

	// Test with failing input (6 + 6 = 12, > 10)
	_, err = composedFn(6)
	if err == nil {
		t.Fatal("Expected error from biconsumer, got nil")
	}
	if !strings.Contains(err.Error(), "biconsumer error: sum 12 too large") {
		t.Errorf("Expected error to contain 'biconsumer error: sum 12 too large', got %v", err)
	}
}
