package kubeovn_subnet

import (
	"testing"
)

func TestExcludeIPsCover(t *testing.T) {
	testcases := []struct {
		name     string
		entries  []string
		expected bool
	}{
		{
			name:     "no entries",
			entries:  nil,
			expected: false,
		},
		{
			name:     "single address equal to the gateway",
			entries:  []string{gatewayIP},
			expected: true,
		},
		{
			name:     "several entries, the last one covers the gateway",
			entries:  []string{excludedIP, excludedRange, gatewayRange},
			expected: true,
		},
		{
			name:     "several entries, none covers the gateway",
			entries:  []string{excludedIP, excludedRange, "10.0.0.2"},
			expected: false,
		},
		{
			name:     "unparsable entries cover nothing",
			entries:  []string{invalidEntry, "10.0.0.1..garbage", "garbage..10.0.0.1"},
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := excludeIPsCover(tc.entries, gatewayIP); actual != tc.expected {
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
		{gatewayIP, gatewayIP, true},
		{"10.0.0.2", gatewayIP, false},
		{gatewayRange, gatewayIP, true},
		{gatewayRange, "10.0.0.10", true},
		{gatewayRange, "10.0.0.11", false},
		{"10.0.0.2..10.0.0.10", gatewayIP, false},
		{"fd00::1..fd00::ff", "fd00::a", true},
		{"fd00::1..fd00::ff", gatewayIP, false},
		{"10.0.0.10..10.0.0.1", "10.0.0.5", false},
		{gatewayIP, invalidEntry, false},
		{"", gatewayIP, false},
		{invalidEntry, gatewayIP, false},
		{"10.0.0.1..garbage", gatewayIP, false},
		{"10.0.0.1..", gatewayIP, false},
	}

	for _, tc := range testcases {
		t.Run(tc.entry+"/"+tc.ip, func(t *testing.T) {
			if actual := excludeEntryCovers(tc.entry, tc.ip); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
