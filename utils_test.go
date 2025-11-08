package main

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"Exact match", "hello", "hello", true},
		{"Substring match", "hello world", "world", true},
		{"Case insensitive - upper in s", "Hello World", "world", true},
		{"Case insensitive - upper in substr", "hello world", "WORLD", true},
		{"Mixed case", "TailPerf-Server", "tailperf", true},
		{"No match", "hello", "goodbye", false},
		{"Empty substring", "hello", "", true},
		{"Empty string", "", "hello", false},
		{"Both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{"Use default when not set", "NONEXISTENT_VAR_12345", "default", "default"},
		{"Empty default", "NONEXISTENT_VAR_12345", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%q, %q) = %v, want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
