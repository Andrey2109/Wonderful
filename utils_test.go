package main

import (
	"testing"
)

func TestLoadEnvVariables(t *testing.T) {
	c := loadEnvVariables()
	if c.APIKey == "" {
		t.Error("Expected API key to be loaded")
	}
}

func TestExecuteLocalFunction(t *testing.T) {
	// Test case 1: Valid multiplication
	result, err := executeLocalFunction("multiply", `{"a": 5, "b": 3}`)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any result, got %T", result)
	}

	if resultMap["result"] != float64(15) {
		t.Errorf("Expected result 15, got %v", resultMap["result"])
	}

	// Test case 2: Unknown function
	result, err = executeLocalFunction("divide", `{"a": 6, "b": 2}`)
	if err != nil {
		t.Errorf("Expected no error for unknown function, got %v", err)
	}

	resultMap, ok = result.(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any result, got %T", result)
	}

	if resultMap["error"] != "unknown function" || resultMap["name"] != "divide" {
		t.Errorf("Expected error for unknown function, got %v", resultMap)
	}

	// Test case 3: Invalid arguments
	_, err = executeLocalFunction("multiply", `{"a": "not a number", "b": 3}`)
	if err == nil {
		t.Error("Expected error for invalid arguments, got nil")
	}
}
