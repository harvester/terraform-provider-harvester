package kubeovn_subnet

import (
	"reflect"
	"testing"
)

func TestWithGatewayExcluded(t *testing.T) {
	testcases := []struct {
		name     string
		entries  []string
		gateway  string
		expected []string
	}{
		{
			name:     "no entries reserves the gateway",
			entries:  nil,
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1"},
		},
		{
			name:     "gateway missing from the entries",
			entries:  []string{"10.0.0.50"},
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1", "10.0.0.50"},
		},
		{
			name:     "gateway already listed",
			entries:  []string{"10.0.0.50", "10.0.0.1"},
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1", "10.0.0.50"},
		},
		{
			name:     "range covering the gateway",
			entries:  []string{"10.0.0.1..10.0.0.10"},
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1..10.0.0.10"},
		},
		{
			name:     "range not covering the gateway",
			entries:  []string{"10.0.0.20..10.0.0.30"},
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1", "10.0.0.20..10.0.0.30"},
		},
		{
			name:     "dual-stack gateway",
			entries:  []string{"10.0.0.50"},
			gateway:  "10.0.0.1,fd00::1",
			expected: []string{"10.0.0.1", "10.0.0.50", "fd00::1"},
		},
		{
			name:     "duplicates are dropped and the result is sorted",
			entries:  []string{"10.0.0.9", "10.0.0.50", "10.0.0.9"},
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1", "10.0.0.50", "10.0.0.9"},
		},
		{
			name:     "empty gateway leaves the entries untouched",
			entries:  []string{"10.0.0.50"},
			gateway:  "",
			expected: []string{"10.0.0.50"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := withGatewayExcluded(tc.entries, tc.gateway)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestExcludeEntryCovers(t *testing.T) {
	testcases := []struct {
		entry    string
		ip       string
		expected bool
	}{
		{"10.0.0.1", "10.0.0.1", true},
		{"10.0.0.2", "10.0.0.1", false},
		{"10.0.0.1..10.0.0.10", "10.0.0.1", true},
		{"10.0.0.1..10.0.0.10", "10.0.0.10", true},
		{"10.0.0.1..10.0.0.10", "10.0.0.11", false},
		{"10.0.0.2..10.0.0.10", "10.0.0.1", false},
		{"fd00::1..fd00::ff", "fd00::a", true},
		{"fd00::1..fd00::ff", "10.0.0.1", false},
		{"not-an-ip", "10.0.0.1", false},
		{"10.0.0.1..garbage", "10.0.0.1", false},
	}

	for _, tc := range testcases {
		t.Run(tc.entry+"/"+tc.ip, func(t *testing.T) {
			if actual := excludeEntryCovers(tc.entry, tc.ip); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
