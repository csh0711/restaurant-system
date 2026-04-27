package model

import "testing"

func TestIsValidMenuItem(t *testing.T) {
	tests := map[string]struct {
		name     string
		input    MenuItem
		expected bool
	}{
		"valid item": {
			input:    "Caesar Salad",
			expected: true,
		},
		"another valid item": {
			input:    "Beef Burger",
			expected: true,
		},
		"invalid item": {
			input:    "Pizza Hawaii",
			expected: false,
		},
		"empty item": {
			input:    "",
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsValid(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
