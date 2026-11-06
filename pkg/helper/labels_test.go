package helper

import "testing"

func TestIsReservedLabelKey(t *testing.T) {
	type testcase struct {
		key         string
		expectation bool
	}

	testcases := []testcase{
		{
			key:         "",
			expectation: false,
		},
		{
			key:         "environment",
			expectation: false,
		},
		{
			key:         "example.com/team",
			expectation: false,
		},
		{
			key:         "harvesterhci.io",
			expectation: false,
		},
		{
			key:         "notharvesterhci.io/foo",
			expectation: false,
		},
		{
			key:         "harvesterhci.io.example.com/foo",
			expectation: false,
		},
		{
			key:         "harvesterhci.io/creator",
			expectation: true,
		},
		{
			key:         "tag.harvesterhci.io/environment",
			expectation: true,
		},
		{
			key:         "network.harvesterhci.io/clusternetwork",
			expectation: true,
		},
		{
			key:         "network.harvesterhci.io/ready",
			expectation: true,
		},
		{
			key:         "network.harvesterhci.io/type",
			expectation: true,
		},
		{
			key:         "network.harvesterhci.io/vlan-id",
			expectation: true,
		},
	}

	for _, tc := range testcases {
		outcome := IsReservedLabelKey(tc.key)
		if outcome != tc.expectation {
			t.Errorf("unexpected outcome for key %v: %v, expected: %v", tc.key, outcome, tc.expectation)
		}
	}
}
