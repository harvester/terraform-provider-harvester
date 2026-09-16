package image

import (
	"strings"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func imageWithDisplayName(name, displayName string) harvsterv1.VirtualMachineImage {
	return harvsterv1.VirtualMachineImage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: harvsterv1.VirtualMachineImageSpec{
			DisplayName: displayName,
		},
	}
}

// TestImagesMatchingDisplayName verifies the matching is driven by
// spec.displayName only: images are matched regardless of the presence of the
// imageDisplayName label, and non-matching images are excluded.
func TestImagesMatchingDisplayName(t *testing.T) {
	items := []harvsterv1.VirtualMachineImage{
		imageWithDisplayName("img-a", "openSUSE Tumbleweed Cloud"),
		imageWithDisplayName("img-b", "sles15-sp7.qcow2"),
		imageWithDisplayName("img-c", "openSUSE Tumbleweed Cloud"),
	}

	testcases := []struct {
		name          string
		displayName   string
		expectedNames []string
	}{
		{
			name:          "no match",
			displayName:   "does-not-exist",
			expectedNames: nil,
		},
		{
			name:          "single match",
			displayName:   "sles15-sp7.qcow2",
			expectedNames: []string{"img-b"},
		},
		{
			name:          "duplicate display names with spaces are all reported",
			displayName:   "openSUSE Tumbleweed Cloud",
			expectedNames: []string{"img-a", "img-c"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			matches := imagesMatchingDisplayName(items, tc.displayName)
			if len(matches) != len(tc.expectedNames) {
				t.Fatalf("got %d matches, want %d", len(matches), len(tc.expectedNames))
			}
			for i, expected := range tc.expectedNames {
				if matches[i].Name != expected {
					t.Errorf("match %d = %q, want %q", i, matches[i].Name, expected)
				}
			}
		})
	}
}

// TestIsValidLabelValue pins the classification used to decide between the
// server-side label fast path and the full list fallback.
func TestIsValidLabelValue(t *testing.T) {
	testcases := []struct {
		value string
		valid bool
	}{
		{"sles15-sp7-minimal-vm.x86_64-cloud-qu2.qcow2", true},
		{"win2025-core-autounattend", true},
		{"openSUSE Tumbleweed Cloud", false},
		{"openSUSE Tumbleweed Cloud (AISCALEIMAGE)", false},
		{"Rocky Linux 9 GenericCloud", false},
		{strings.Repeat("a", 64), false},
		{"", true},
	}
	for _, tc := range testcases {
		if got := isValidLabelValue(tc.value); got != tc.valid {
			t.Errorf("isValidLabelValue(%q) = %v, want %v", tc.value, got, tc.valid)
		}
	}
}
