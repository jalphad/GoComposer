package functions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jalphad/gocomposer/workflows/internal/composer"
)

func TestAddProducerAsIntermediateFunction(t *testing.T) {
	// Test Producer that generates data for other functions to consume
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add Producer: generate a constant value
	prod := AddProducer(swf, func() (int, error) {
		return 42, nil
	}, nil)

	// Add function that uses the producer output
	dep := AddFn(swf, func(num int) (string, error) {
		return fmt.Sprintf("number: %d", num), nil
	}, &FnOpts[int]{
		Input: prod,
	})

	// Add output function that uses both workflow input and producer-generated data
	AddBiFn(swf, func(input, generated string) (string, error) {
		return fmt.Sprintf("input=%s, generated=%s", input, generated), nil
	}, &BiFnOpts[string, string]{
		Input1: nil, // workflow input
		Input2: dep,
	})

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

	expectedOutput := "input=hello, generated=number: 42"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddProducerAsOutputFunction(t *testing.T) {
	// Test Producer that directly provides the workflow output
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add Producer that generates the output directly
	AddProducer(swf, func() (string, error) {
		return "produced output", nil
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "ignored"
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	expectedOutput := "produced output"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddProducerMultiple(t *testing.T) {
	// Test multiple producers feeding into a combining function
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add first producer
	prod1 := AddProducer(swf, func() (int, error) {
		return 10, nil
	}, &ProducerOpts{Name: "Producer1"})

	// Add second producer
	prod2 := AddProducer(swf, func() (string, error) {
		return "data", nil
	}, &ProducerOpts{Name: "Producer2"})

	// Add third producer
	prod3 := AddProducer(swf, func() (bool, error) {
		return true, nil
	}, &ProducerOpts{Name: "Producer3"})

	// Combine all producer outputs with workflow input
	AddQuadFn(swf, func(input string, num int, data string, flag bool) (string, error) {
		flagStr := "false"
		if flag {
			flagStr = "true"
		}
		return fmt.Sprintf("input=%s, num=%d, data=%s, flag=%s", input, num, data, flagStr), nil
	}, &QuadFnOpts[string, int, string, bool]{
		Input1: nil, // workflow input
		Input2: prod1,
		Input3: prod2,
		Input4: prod3,
	})

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

	expectedOutput := "input=test, num=10, data=data, flag=true"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddProducerWithError(t *testing.T) {
	// Test Producer that returns an error
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add Producer that fails under certain conditions
	AddProducer(swf, func() (string, error) {
		return "", fmt.Errorf("producer error: something went wrong")
	}, nil)

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test with input that should trigger error
	_, err = composedFn("test")
	if err == nil {
		t.Fatal("Expected error from producer, got nil")
	}
	if !strings.Contains(err.Error(), "producer error: something went wrong") {
		t.Errorf("Expected error to contain 'producer error: something went wrong', got %v", err)
	}
}

func TestAddProducerChained(t *testing.T) {
	// Test Producer output feeding into another function which then feeds into final output
	swf := composer.NewSimpleWorkflow[int, string]()

	// Add Producer that generates base data
	prod := AddProducer(swf, func() (string, error) {
		return "base", nil
	}, nil)

	// Add function that processes producer output with workflow input
	processed := AddBiFn(swf, func(num int, base string) (string, error) {
		return fmt.Sprintf("%s-%d", base, num), nil
	}, &BiFnOpts[int, string]{
		Input1: nil, // workflow input
		Input2: prod,
	})

	// Add final transformation
	AddFn(swf, func(combined string) (string, error) {
		return fmt.Sprintf("final: %s", combined), nil
	}, &FnOpts[string]{
		Input: processed,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := 123
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	expectedOutput := "final: base-123"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddProducerWithConsumer(t *testing.T) {
	// Test Producer output being consumed by a consumer
	swf := composer.NewSimpleWorkflow[string, string]()

	var consumed string

	// Add Producer
	prod := AddProducer(swf, func() (int, error) {
		return 999, nil
	}, nil)

	// Add Consumer that consumes producer output
	AddConsumer(swf, func(num int) error {
		consumed = fmt.Sprintf("consumed: %d", num)
		return nil
	}, &ConsmrOpts[int]{
		Input: prod,
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

	// Check that consumer was called with producer output
	expectedConsumed := "consumed: 999"
	if consumed != expectedConsumed {
		t.Errorf("Expected consumed %q, got %q", expectedConsumed, consumed)
	}

	expectedOutput := "result: test"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestAddProducerNaming(t *testing.T) {
	// Test that producer naming works correctly
	swf := composer.NewSimpleWorkflow[string, string]()

	// Add producer with custom name
	prod := AddProducer(swf, func() (string, error) {
		return "custom", nil
	}, &ProducerOpts{Name: "CustomProducer"})

	// Add function that uses the named producer
	AddFn(swf, func(data string) (string, error) {
		return fmt.Sprintf("from custom producer: %s", data), nil
	}, &FnOpts[string]{
		Input: prod,
	})

	// Compose the workflow
	composedFn, err := swf.Compose()
	if err != nil {
		t.Fatalf("Failed to compose workflow: %v", err)
	}

	// Test the composed function
	input := "ignored"
	output, err := composedFn(input)
	if err != nil {
		t.Fatalf("Composed function returned error: %v", err)
	}

	expectedOutput := "from custom producer: custom"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}