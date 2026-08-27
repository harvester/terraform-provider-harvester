package importer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestGetNamespaceLabels(t *testing.T) {
	testcases := []struct {
		name     string
		labels   map[string]string
		expected map[string]string
	}{
		{
			name:     "nil labels",
			labels:   nil,
			expected: map[string]string{},
		},
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "filters kubernetes.io/metadata.name",
			labels: map[string]string{
				"kubernetes.io/metadata.name": "my-ns",
				"custom-label":                "value",
			},
			expected: map[string]string{
				"custom-label": "value",
			},
		},
		{
			name: "filters cattle.io and lifecycle.cattle.io prefixes",
			labels: map[string]string{
				"cattle.io/status":                   "something",
				"lifecycle.cattle.io/create.ns-auth": "true",
				"kubernetes.io/metadata.name":        "my-ns",
				"tag.harvesterhci.io/env":            "test",
				"harvesterhci.io/some-internal":      "internal",
				"custom-label":                       "value",
			},
			expected: map[string]string{
				"custom-label": "value",
			},
		},
		{
			name: "preserves user labels only",
			labels: map[string]string{
				"kubernetes.io/metadata.name": "my-ns",
				"app":                         "myapp",
				"tier":                        "frontend",
			},
			expected: map[string]string{
				"app":  "myapp",
				"tier": "frontend",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := getNamespaceLabels(tc.labels)
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d labels, got %d: %v", len(tc.expected), len(result), result)
				return
			}
			for key, val := range tc.expected {
				if result[key] != val {
					t.Errorf("label %q: expected %q, got %q", key, val, result[key])
				}
			}
		})
	}
}

func TestResourceNamespaceStateGetter(t *testing.T) {
	testcases := []struct {
		name          string
		namespace     *corev1.Namespace
		expectedID    string
		expectedState string
		expectedDesc  string
		expectedTags  map[string]string
	}{
		{
			name: "active namespace with tags and description",
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ns",
					Labels: map[string]string{
						"tag.harvesterhci.io/env":     "test",
						"kubernetes.io/metadata.name": "test-ns",
					},
					Annotations: map[string]string{
						"field.cattle.io/description": "A test namespace",
					},
				},
				Status: corev1.NamespaceStatus{
					Phase: corev1.NamespaceActive,
				},
			},
			expectedID:    helper.BuildID("", "test-ns"),
			expectedState: "Active",
			expectedDesc:  "A test namespace",
			expectedTags:  map[string]string{"env": "test"},
		},
		{
			name: "terminating namespace with no tags",
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deleting-ns",
					Labels: map[string]string{
						"kubernetes.io/metadata.name": "deleting-ns",
					},
				},
				Status: corev1.NamespaceStatus{
					Phase: corev1.NamespaceTerminating,
				},
			},
			expectedID:    helper.BuildID("", "deleting-ns"),
			expectedState: "Terminating",
			expectedDesc:  "",
			expectedTags:  map[string]string{},
		},
		{
			name: "namespace with nil labels and annotations",
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bare-ns",
				},
				Status: corev1.NamespaceStatus{
					Phase: corev1.NamespaceActive,
				},
			},
			expectedID:    helper.BuildID("", "bare-ns"),
			expectedState: "Active",
			expectedDesc:  "",
			expectedTags:  map[string]string{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceNamespaceStateGetter(tc.namespace)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.Name != tc.namespace.Name {
				t.Errorf("Name: expected %q, got %q", tc.namespace.Name, getter.Name)
			}
			if getter.ResourceType != constants.ResourceTypeNamespace {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeNamespace, getter.ResourceType)
			}

			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}

			desc := getter.States[constants.FieldCommonDescription].(string)
			if desc != tc.expectedDesc {
				t.Errorf("Description: expected %q, got %q", tc.expectedDesc, desc)
			}

			tags := getter.States[constants.FieldCommonTags].(map[string]string)
			if len(tags) != len(tc.expectedTags) {
				t.Errorf("Tags: expected %d, got %d: %v", len(tc.expectedTags), len(tags), tags)
			}
			for key, val := range tc.expectedTags {
				if tags[key] != val {
					t.Errorf("Tag %q: expected %q, got %q", key, val, tags[key])
				}
			}

			labels := getter.States[constants.FieldCommonLabels].(map[string]string)
			for key := range labels {
				if key == "kubernetes.io/metadata.name" {
					t.Errorf("Labels should not contain kubernetes.io/metadata.name")
				}
			}
		})
	}
}
