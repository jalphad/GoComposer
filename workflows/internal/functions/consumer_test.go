package functions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddConsumerWithWorkflowInput(t *testing.T) {
	// Test Consumer that takes workflow input directly
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed string

	// Add Consumer: consume workflow input
	AddConsumer(swf, func(s string) error {
		consumed = fmt.Sprintf("consumed: %s", s)
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

	// Check that consumer was called
	expectedConsumed := "consumed: test"
	if consumed != expectedConsumed {
		t.Errorf("Expected consumed %q, got %q", expectedConsumed, consumed)
	}

	// Check output
	expectedOutput := "processed: test"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddConsumerWithFunctionDependency(t *testing.T) {
	// Test Consumer that takes output from another function
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed string

	// Add function: string -> int (length)
	dep := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add Consumer: consume the length
	AddConsumer(swf, func(length int) error {
		consumed = fmt.Sprintf("consumed length: %d", length)
		return nil
	}, &ConsmrOpts[int]{
		Input: dep,
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

	// Check that consumer was called
	expectedConsumed := "consumed length: 5"
	if consumed != expectedConsumed {
		t.Errorf("Expected consumed %q, got %q", expectedConsumed, consumed)
	}

	// Check output
	expectedOutput := "processed: hello"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddConsumerMultiple(t *testing.T) {
	// Test multiple consumers
	swf := composer.NewSimpleWorkflow[string, string]()

	// Track consumed values
	var consumed1, consumed2 string

	// Add function: string -> int (length)
	dep := AddFn(swf, func(s string) (int, error) {
		return len(s), nil
	}, nil)

	// Add first Consumer: consume workflow input
	AddConsumer(swf, func(s string) error {
		consumed1 = fmt.Sprintf("consumer1: %s", strings.ToUpper(s))
		return nil
	}, nil)

	// Add second Consumer: consume the length
	AddConsumer(swf, func(length int) error {
		consumed2 = fmt.Sprintf("consumer2: %d chars", length)
		return nil
	}, &ConsmrOpts[int]{
		Input: dep,
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

	// Check that both consumers were called
	expectedConsumed1 := "consumer1: TEST"
	if consumed1 != expectedConsumed1 {
		t.Errorf("Expected consumed1 %q, got %q", expectedConsumed1, consumed1)
	}

	expectedConsumed2 := "consumer2: 4 chars"
	if consumed2 != expectedConsumed2 {
		t.Errorf("Expected consumed2 %q, got %q", expectedConsumed2, consumed2)
	}

	expectedOutput := "result: test"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddConsumerWithError(t *testing.T) {
	// Test Consumer that returns an error
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add Consumer that fails
	AddConsumer(swf, func(s string) error {
		if s == "fail" {
			return fmt.Errorf("consumer error: %s", s)
		}
		return nil
	}, nil)

	// Add output function
	AddFn(swf, func(s string) (string, error) {
		return fmt.Sprintf("result: %s", s), nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with successful input
	output, err := composedFn("success")
	if err != nil {
		t.Fatalf("Composed function returned unexpected error: %v", err)
	}
	if output != "result: success" {
		t.Errorf("Expected output 'result: success', got %q", output)
	}

	// Test with failing input
	_, err = composedFn("fail")
	if err == nil {
		t.Fatal("Expected error from consumer, got nil")
	}
	if !strings.Contains(err.Error(), "consumer error: fail") {
		t.Errorf("Expected error to contain 'consumer error: fail', got %v", err)
	}
}
