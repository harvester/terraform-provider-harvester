package kubeovn_subnet

import (
	"strings"
	"testing"
)

func TestNormalizeExcludeIPs(t *testing.T) {
	const (
		gatewayV6 = "fd00::1"
		dualGW    = gatewayIP + "," + gatewayV6
	)

	testcases := []struct {
		name     string
		entries  []string
		gateway  string
		expected []string
	}{
		{
			name:     "no entries reserves the gateway",
			gateway:  gatewayIP,
			expected: []string{gatewayIP},
		},
		{
			name:     "gateway already listed",
			entries:  []string{excludedIP, gatewayIP},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, excludedIP},
		},
		{
			name:     "range covering the gateway",
			entries:  []string{gatewayRange},
			gateway:  gatewayIP,
			expected: []string{gatewayRange},
		},
		{
			name:     "overlapping ranges are merged",
			entries:  []string{lowerRange, "10.0.0.24..10.0.0.30"},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, excludedRange},
		},
		{
			name:     "ranges sharing a bound are merged",
			entries:  []string{lowerRange, "10.0.0.25..10.0.0.30"},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, excludedRange},
		},
		{
			name:     "contained range and address are dropped",
			entries:  []string{excludedRange, "10.0.0.22..10.0.0.23", "10.0.0.21"},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, excludedRange},
		},
		{
			name:     "adjacent ranges are kept apart",
			entries:  []string{lowerRange, "10.0.0.26..10.0.0.30"},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, lowerRange, "10.0.0.26..10.0.0.30"},
		},
		{
			name:     "range of one address is written as the address",
			entries:  []string{excludedIP + ".." + excludedIP},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, excludedIP},
		},
		{
			name:     "duplicates are dropped",
			entries:  []string{excludedIP, excludedIP},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, excludedIP},
		},
		{
			name:     "dual-stack gateway reserves both addresses",
			entries:  []string{excludedIP, "fd00::10..fd00::20"},
			gateway:  dualGW,
			expected: []string{gatewayIP, excludedIP, gatewayV6, "fd00::10..fd00::20"},
		},
		{
			name:     "IPv4 and IPv6 ranges are never merged",
			entries:  []string{"10.0.0.1..10.0.0.255", ipv6Range},
			gateway:  dualGW,
			expected: []string{"10.0.0.1..10.0.0.255", ipv6Range},
		},
		{
			name:     "unparsable and reversed entries are kept as they are",
			entries:  []string{invalidEntry, "10.0.0.10..10.0.0.5", garbageRange},
			gateway:  gatewayIP,
			expected: []string{gatewayIP, garbageRange, "10.0.0.10..10.0.0.5", invalidEntry},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizeExcludeIPs(tc.entries, tc.gateway)
			if strings.Join(actual, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestParseExcludeEntry(t *testing.T) {
	testcases := []struct {
		entry         string
		expectedStart string
		expectedEnd   string
		expectedOK    bool
	}{
		{entry: gatewayIP, expectedStart: gatewayIP, expectedEnd: gatewayIP, expectedOK: true},
		{entry: gatewayRange, expectedStart: gatewayIP, expectedEnd: "10.0.0.10", expectedOK: true},
		{entry: " 10.0.0.1 .. 10.0.0.10 ", expectedStart: gatewayIP, expectedEnd: "10.0.0.10", expectedOK: true},
		{entry: ipv6Range, expectedStart: "fd00::1", expectedEnd: "fd00::ff", expectedOK: true},
		{entry: "10.0.0.10..10.0.0.1"},
		{entry: "10.0.0.1..fd00::1"},
		{entry: ""},
		{entry: invalidEntry},
		{entry: garbageRange},
		{entry: "10.0.0.1.."},
	}

	for _, tc := range testcases {
		t.Run(tc.entry, func(t *testing.T) {
			r, ok := parseExcludeEntry(tc.entry)
			if ok != tc.expectedOK {
				t.Fatalf("expected ok=%v, got %v", tc.expectedOK, ok)
			}
			if ok && (r.start.String() != tc.expectedStart || r.end.String() != tc.expectedEnd) {
				t.Errorf("expected %s..%s, got %s..%s", tc.expectedStart, tc.expectedEnd, r.start, r.end)
			}
		})
	}
}
