package deduplicator

import (
	"testing"
)

func TestShouldSend(t *testing.T) {
	// Reset sentHashes map before each test to avoid conflicts between tests
	sentHashes = make(map[string]bool)

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{"First time message", "Hello, world!", true},
		{"Duplicate message", "Hello, world!", false},
		{"Different message", "Goodbye, world!", true},
		{"Same message again", "Hello, world!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Deduplicator{}
			result := d.ShouldSend(tt.message)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v for message: %s", tt.expected, result, tt.message)
			}
		})
	}
}