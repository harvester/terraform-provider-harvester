package helper

import "testing"

func TestIsIPv4LinkLocal(t *testing.T) {
	type testcase struct {
		address     string
		expectation bool
	}

	testcases := []testcase{
		{
			address:     "",
			expectation: false,
		},
		{
			address:     "192.168.178.64",
			expectation: false,
		},
		{
			address:     "127.0.0.1",
			expectation: false,
		},
		{
			address:     "",
			expectation: false,
		},
		{
			address:     "2001:9e8:279:b900:21f:bcff:fe13:405/64",
			expectation: false,
		},
		{
			address:     "fe80::21f:bcff:fe13:405/64",
			expectation: false,
		},
		{
			address:     "169.254.178.34",
			expectation: true,
		},
		{
			address:     "169.254.34.98/24",
			expectation: true,
		},
		{
			address:     "169.254.0.34",
			expectation: false,
		},
		{
			address:     "169.254.255.98/24",
			expectation: false,
		},
	}

	for _, tc := range testcases {
		outcome := IsIPv4LinkLocal(tc.address)
		if outcome != tc.expectation {
			t.Errorf("unexpected outcome for address %v: %v, expected: %v", tc.address, outcome, tc.expectation)
		}
	}
}

func TestIsIPv6LinkLocal(t *testing.T) {
	type testcase struct {
		address     string
		expectation bool
	}

	testcases := []testcase{
		{
			address:     "",
			expectation: false,
		},
		{
			address:     "192.168.178.64",
			expectation: false,
		},
		{
			address:     "127.0.0.1",
			expectation: false,
		},
		{
			address:     "169.254.178.34",
			expectation: false,
		},
		{
			address:     "169.254.34.98/64",
			expectation: false,
		},
		{
			address:     "",
			expectation: false,
		},
		{
			address:     "2001:9e8:279:b900:21f:bcff:fe13:405/64",
			expectation: false,
		},
		{
			address:     "fe80::21f:bcff:fe13:405/64",
			expectation: true,
		},
		{
			address:     "fe80::21f:bcff:fe13:405",
			expectation: true,
		},
		{
			address:     "fe80::",
			expectation: true,
		},
		{
			address:     "FE80::21F:BCFF:FE13:405/64",
			expectation: true,
		},
		{
			address:     "FE80::21F:BCFF:FE13:405",
			expectation: true,
		},
		{
			address:     "FE80::",
			expectation: true,
		},
	}

	for _, tc := range testcases {
		outcome := IsIPv6LinkLocal(tc.address)
		if outcome != tc.expectation {
			t.Errorf("unexpected outcome for address %v: %v, expected: %v", tc.address, outcome, tc.expectation)
		}
	}
}
