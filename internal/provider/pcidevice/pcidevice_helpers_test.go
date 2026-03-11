package pcidevice

import (
	"testing"
)

func TestGetField(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{"nil map", nil, "key", ""},
		{"missing key", map[string]interface{}{"other": "val"}, "key", ""},
		{"present key", map[string]interface{}{"key": "val"}, "key", "val"},
		{"non-string value", map[string]interface{}{"key": 42}, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getField(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getField(%v, %q) = %q, want %q", tt.m, tt.key, result, tt.expected)
			}
		})
	}
}
