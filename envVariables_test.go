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
