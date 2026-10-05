package virtualmachine

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	networkapi "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io"
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

const (
	affinityTestLabel           = "app"
	affinityTestValue           = "web"
	affinityTestNode            = "node1"
	affinityPreferenceTestKey   = "disktype"
	affinityPreferenceTestOp    = "In"
	affinityPreferenceTestValue = "ssd"
)

func affinityExpression(key, operator string, values ...string) map[string]interface{} {
	vals := make([]interface{}, 0, len(values))
	for _, v := range values {
		vals = append(vals, v)
	}
	return map[string]interface{}{
		constants.FieldExpressionKey:      key,
		constants.FieldExpressionOperator: operator,
		constants.FieldExpressionValues:   vals,
	}
}

// nodeAffinityBlock returns a node_affinity block with one required term.
func nodeAffinityBlock(term interface{}) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldNodeAffinityRequired: []interface{}{map[string]interface{}{
			constants.FieldNodeSelectorTerm: []interface{}{term},
		}},
	}
}

func TestParseNodeAffinity(t *testing.T) {
	testcases := []struct {
		name              string
		block             map[string]any
		expected_affinity *corev1.NodeAffinity
		expected_error    error
	}{
		{
			name:              "empty block", //nolint:goconst
			block:             util.BlockMap(nil),
			expected_affinity: nil,
			expected_error:    ErrMissingBlock,
		},
		{
			name:              "empty term",
			block:             nodeAffinityBlock(nil),
			expected_affinity: nil,
			expected_error:    ErrMissingMatchExpression,
		},
		{
			name: "match_fields only",
			block: nodeAffinityBlock(
				map[string]any{
					constants.FieldMatchFields: []any{
						affinityExpression("metadata.name", "In", affinityTestNode),
					},
				},
			),
			expected_affinity: nil,
			expected_error:    ErrMissingMatchExpression,
		},
		{
			name: "empty preference",
			block: map[string]any{
				constants.FieldNodeAffinityPreferred: []any{
					map[string]any{
						constants.FieldPreferredWeight:     10,
						constants.FieldPreferredPreference: []any{},
					},
				},
			},
			expected_affinity: nil,
			expected_error:    ErrMissingTerms,
		},
		{
			name: "parse valid node affinity block",
			block: map[string]any{
				constants.FieldNodeAffinityRequired: []any{map[string]any{
					constants.FieldNodeSelectorTerm: []any{
						map[string]any{
							constants.FieldMatchExpressions: []any{
								affinityExpression(corev1.LabelHostname, "In", affinityTestNode),
							},
							constants.FieldMatchFields: []any{
								affinityExpression("metadata.name", "In", affinityTestNode),
							},
						},
					},
				}},
				constants.FieldNodeAffinityPreferred: []any{map[string]any{
					constants.FieldPreferredWeight: 10,
					constants.FieldPreferredPreference: []any{
						map[string]any{constants.FieldMatchExpressions: []any{
							affinityExpression(affinityPreferenceTestKey, affinityPreferenceTestOp, affinityPreferenceTestValue),
						}},
					},
				}},
			},
			expected_affinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      corev1.LabelHostname,
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{affinityTestNode},
								},
							},
							MatchFields: []corev1.NodeSelectorRequirement{
								{
									Key:      "metadata.name",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{affinityTestNode},
								},
							},
						},
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
					{
						Weight: 10,
						Preference: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      affinityPreferenceTestKey,
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{affinityPreferenceTestValue},
								},
							},
						},
					},
				},
			},
			expected_error: nil,
		},
		{
			name: "parse node affinity block without node selector term",
			block: map[string]any{
				constants.FieldNodeAffinityRequired: []any{map[string]any{}},
			},
			expected_affinity: nil,
			expected_error:    ErrMissingTerms,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseNodeAffinity(tc.block)
			if err != nil || tc.expected_error != nil {
				if !errors.Is(err, tc.expected_error) {
					t.Errorf("unexpected error: %s", err)
				}
			}
			if !reflect.DeepEqual(got, tc.expected_affinity) {
				t.Errorf("parseNodeAffinity() = %+v, want %+v", got, tc.expected_affinity)
			}
		})
	}
}

func podAffinityTermBlock(labelSelector, namespaceSelector interface{}) map[string]interface{} {
	term := map[string]interface{}{constants.FieldTopologyKey: corev1.LabelHostname}
	if labelSelector != nil {
		term[constants.FieldLabelSelector] = []interface{}{labelSelector}
	}
	if namespaceSelector != nil {
		term[constants.FieldNamespaceSelector] = []interface{}{namespaceSelector}
	}
	return term
}

// TestParsePodAffinityNamespaceSelector checks that an empty namespace
// selector selects every namespace, like the "all namespaces" option of the
// Harvester UI, instead of being dropped.
func TestParsePodAffinityNamespaceSelector(t *testing.T) {
	block := map[string]interface{}{
		constants.FieldPodAffinityRequired: []interface{}{
			podAffinityTermBlock(map[string]interface{}{constants.FieldMatchLabels: map[string]interface{}{affinityTestLabel: affinityTestValue}}, nil),
		},
	}
	block[constants.FieldPodAffinityRequired].([]interface{})[0].(map[string]interface{})[constants.FieldNamespaceSelector] = []interface{}{nil}
	required, _, err := parsePodAffinityAndAntiAffinityTerms(constants.FieldVirtualMachinePodAntiAffinity, block)
	if err != nil {
		t.Fatal(err)
	}
	if len(required) != 1 || !reflect.DeepEqual(required[0].NamespaceSelector, &metav1.LabelSelector{}) {
		t.Errorf("namespace selector = %+v, want an empty selector", required)
	}
}

func TestParsePodAffinityRulesRejects(t *testing.T) {
	testcases := []struct {
		name              string
		block             map[string]any
		expected_affinity *corev1.PodAffinity
		expected_error    error
	}{
		{
			name:              "empty block (nil from Terraform)",
			block:             util.BlockMap(nil),
			expected_affinity: nil,
			expected_error:    ErrMissingBlock,
		},
		{
			name: "empty label selector",
			block: map[string]any{
				constants.FieldPodAffinityRequired: []any{podAffinityTermBlock(map[string]any{}, nil)},
			},
			expected_affinity: nil,
			expected_error:    ErrMissingLabelMatcher,
		},
		{
			name: "empty label selector in a preferred term",
			block: map[string]any{
				constants.FieldPodAffinityPreferred: []any{map[string]any{
					constants.FieldPreferredWeight: 20,
					constants.FieldPodAffinityTerm: []any{podAffinityTermBlock(map[string]any{}, nil)},
				}},
			},
			expected_affinity: nil,
			expected_error:    ErrMissingLabelMatcher,
		},
		{
			name: "pod affinity with required and preferred",
			block: map[string]any{
				constants.FieldPodAffinityRequired: []any{
					podAffinityTermBlock(map[string]any{constants.FieldMatchExpressions: []any{affinityExpression(affinityTestLabel, "In", affinityTestValue)}}, nil),
				},
				constants.FieldPodAffinityPreferred: []any{map[string]any{
					constants.FieldPreferredWeight: 20,
					constants.FieldPodAffinityTerm: []any{podAffinityTermBlock(map[string]any{constants.FieldMatchLabels: map[string]any{affinityTestLabel: affinityTestValue}}, nil)},
				}},
			},
			expected_affinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
					TopologyKey:   corev1.LabelHostname,
					LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: affinityTestLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{affinityTestValue}}}},
				}},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
					Weight: 20,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey:   corev1.LabelHostname,
						LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{affinityTestLabel: affinityTestValue}},
					},
				}},
			},
			expected_error: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePodAffinity(tc.block)
			if err != nil || tc.expected_error != nil {
				if !errors.Is(err, tc.expected_error) {
					t.Errorf("unexpected error: %s", err)
				}
			}
			if !reflect.DeepEqual(got, tc.expected_affinity) {
				t.Errorf("parsePodAffinity() = %+v, want %+v", got, tc.expected_affinity)
			}
		})
	}
}

func TestUpdaterResetsAffinity(t *testing.T) {
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
		Spec: kubevirtv1.VirtualMachineSpec{Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
			Spec: kubevirtv1.VirtualMachineInstanceSpec{Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: corev1.LabelHostname, Operator: corev1.NodeSelectorOpIn, Values: []string{affinityTestNode}}}}},
				}},
			}},
		}},
	}
	updater := Updater(nil, context.Background(), vm).(*Constructor)
	expected := builder.NewVMBuilder(vmCreator).DefaultPodAntiAffinity().VirtualMachine.Spec.Template.Spec.Affinity
	if got := updater.Builder.VirtualMachine.Spec.Template.Spec.Affinity; !reflect.DeepEqual(got, expected) {
		t.Errorf("affinity after Updater() = %+v, want the default of Creator() %+v", got, expected)
	}
}

// TestPodAntiAffinityParser checks that the user rules are added to the
// default anti-affinity, and that a copy of the default term is refused since
// it would be filtered on read and drift.
func TestPodAntiAffinityParser(t *testing.T) {
	vmBuilder := builder.NewVMBuilder(vmCreator).DefaultPodAntiAffinity()
	defaultTerm := vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
	userBlock := map[string]interface{}{
		constants.FieldPodAffinityRequired: []interface{}{
			podAffinityTermBlock(map[string]interface{}{constants.FieldMatchLabels: map[string]interface{}{affinityTestLabel: affinityTestValue}}, nil),
		},
	}
	if err := podAntiAffinityParser(vmBuilder)(userBlock); err != nil {
		t.Fatal(err)
	}
	podAntiAffinity := vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity.PodAntiAffinity
	if len(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		!reflect.DeepEqual(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, []corev1.WeightedPodAffinityTerm{defaultTerm}) {
		t.Errorf("pod anti-affinity = %+v, want the user rule next to the default term", podAntiAffinity)
	}

	defaultCopy := map[string]interface{}{
		constants.FieldPodAffinityPreferred: []interface{}{map[string]interface{}{
			constants.FieldPreferredWeight: int(defaultTerm.Weight),
			constants.FieldPodAffinityTerm: []interface{}{podAffinityTermBlock(map[string]interface{}{
				constants.FieldMatchExpressions: []interface{}{affinityExpression(builder.LabelKeyVirtualMachineCreator, string(metav1.LabelSelectorOpExists))},
			}, nil)},
		}},
	}
	if err := podAntiAffinityParser(builder.NewVMBuilder(vmCreator).DefaultPodAntiAffinity())(defaultCopy); err == nil {
		t.Error("a copy of the default term must be refused")
	}
}

// TestAffinityRoundTrip builds a VM from a configuration that sets every
// affinity field, reads it back with the importer and checks that the state
// equals the configuration, so that no field is lost or added on the way.
func TestAffinityRoundTrip(t *testing.T) {
	const otherNamespace = "other"
	labelSelector := []interface{}{map[string]interface{}{constants.FieldMatchLabels: map[string]interface{}{affinityTestLabel: affinityTestValue}}}
	raw := map[string]interface{}{
		constants.FieldCommonName:      "round-trip",
		constants.FieldCommonNamespace: "default",
		constants.FieldVirtualMachineNodeAffinity: []interface{}{map[string]interface{}{
			constants.FieldNodeAffinityRequired: []interface{}{map[string]interface{}{
				constants.FieldNodeSelectorTerm: []interface{}{map[string]interface{}{
					constants.FieldMatchExpressions: []interface{}{affinityExpression(corev1.LabelHostname, "In", affinityTestNode, "node2")},
					constants.FieldMatchFields:      []interface{}{affinityExpression("metadata.name", "NotIn", "node3")},
				}},
			}},
			constants.FieldNodeAffinityPreferred: []interface{}{map[string]interface{}{
				constants.FieldPreferredWeight: 10,
				constants.FieldPreferredPreference: []interface{}{map[string]interface{}{
					constants.FieldMatchExpressions: []interface{}{affinityExpression(affinityPreferenceTestKey, "Exists")},
				}},
			}},
		}},
		constants.FieldVirtualMachinePodAffinity: []interface{}{map[string]interface{}{
			constants.FieldPodAffinityRequired: []interface{}{map[string]interface{}{
				constants.FieldTopologyKey:   corev1.LabelHostname,
				constants.FieldLabelSelector: labelSelector,
				constants.FieldNamespaces:    []interface{}{"default", otherNamespace},
			}},
		}},
		constants.FieldVirtualMachinePodAntiAffinity: []interface{}{map[string]interface{}{
			constants.FieldPodAffinityPreferred: []interface{}{map[string]interface{}{
				constants.FieldPreferredWeight: 40,
				constants.FieldPodAffinityTerm: []interface{}{map[string]interface{}{
					constants.FieldTopologyKey:       corev1.LabelHostname,
					constants.FieldLabelSelector:     []interface{}{map[string]interface{}{constants.FieldMatchExpressions: []interface{}{affinityExpression(affinityTestLabel, "In", affinityTestValue)}}},
					constants.FieldNamespaceSelector: []interface{}{map[string]interface{}{}},
					constants.FieldNamespaces:        []interface{}{otherNamespace},
				}},
			}},
		}},
	}
	ctx := context.Background()
	config := schema.TestResourceDataRaw(t, Schema(), raw)
	obj, err := util.ResourceConstruct(ctx, config, Creator(nil, ctx, "default", "round-trip"))
	if err != nil {
		t.Fatal(err)
	}
	vmImporter := importer.NewVMImporter(obj.(*kubevirtv1.VirtualMachine), nil)
	state := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{})
	exported := map[string][]map[string]interface{}{
		constants.FieldVirtualMachineNodeAffinity:    vmImporter.NodeAffinity(),
		constants.FieldVirtualMachinePodAffinity:     vmImporter.PodAffinity(),
		constants.FieldVirtualMachinePodAntiAffinity: vmImporter.PodAntiAffinity(),
	}
	for field, value := range exported {
		if err := state.Set(field, value); err != nil {
			t.Fatalf("set %s: %v", field, err)
		}
		if !reflect.DeepEqual(state.Get(field), config.Get(field)) {
			t.Errorf("%s read back as %v, want %v", field, state.Get(field), config.Get(field))
		}
	}
}

// TestUpdaterKeepsHarvesterManagedRequirements checks that an update carries
// the requirements Harvester manages (CPU manager, cluster networks): Harvester
// v1.9 refuses an update that changes the CPU manager ones.
func TestUpdaterKeepsHarvesterManagedRequirements(t *testing.T) {
	cpuManager := corev1.NodeSelectorRequirement{Key: helper.CPUManagerLabel, Operator: corev1.NodeSelectorOpIn, Values: []string{strconv.FormatBool(true)}}
	network := corev1.NodeSelectorRequirement{Key: networkapi.GroupName + "/mgmt", Operator: corev1.NodeSelectorOpIn, Values: []string{strconv.FormatBool(true)}}
	user := corev1.NodeSelectorRequirement{Key: corev1.LabelHostname, Operator: corev1.NodeSelectorOpIn, Values: []string{affinityTestNode}}
	newVM := func() *kubevirtv1.VirtualMachine {
		return &kubevirtv1.VirtualMachine{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			Spec: kubevirtv1.VirtualMachineSpec{Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{MatchExpressions: []corev1.NodeSelectorRequirement{user, cpuManager, network}},
							{MatchExpressions: []corev1.NodeSelectorRequirement{cpuManager}},
						},
					}},
				}},
			}},
		}
	}
	required := func(c *Constructor) []corev1.NodeSelectorTerm {
		return c.Builder.VirtualMachine.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	}

	updater := Updater(nil, context.Background(), newVM()).(*Constructor)
	expected := []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{cpuManager, network, cpuManager}}}
	if got := required(updater); !reflect.DeepEqual(got, expected) {
		t.Errorf("required terms after Updater() = %+v, want only the managed requirements %+v", got, expected)
	}

	// A node_affinity block keeps them next to the user rules.
	updater = Updater(nil, context.Background(), newVM()).(*Constructor)
	if err := nodeAffinityParser(updater.Builder)(nodeAffinityBlock(map[string]interface{}{
		constants.FieldMatchExpressions: []interface{}{affinityExpression(corev1.LabelHostname, "In", affinityTestNode)},
	})); err != nil {
		t.Fatal(err)
	}
	expected = []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{user, cpuManager, network, cpuManager}}}
	if got := required(updater); !reflect.DeepEqual(got, expected) {
		t.Errorf("required terms with node_affinity = %+v, want %+v", got, expected)
	}

	// With preferred rules only, they stay in a required term of their own.
	updater = Updater(nil, context.Background(), newVM()).(*Constructor)
	if err := nodeAffinityParser(updater.Builder)(map[string]interface{}{
		constants.FieldNodeAffinityPreferred: []interface{}{map[string]interface{}{
			constants.FieldPreferredWeight: 10,
			constants.FieldPreferredPreference: []interface{}{
				map[string]interface{}{constants.FieldMatchExpressions: []interface{}{affinityExpression(corev1.LabelHostname, "In", affinityTestNode)}},
			},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	expected = []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{cpuManager, network, cpuManager}}}
	if got := required(updater); !reflect.DeepEqual(got, expected) {
		t.Errorf("required terms with preferred rules only = %+v, want %+v", got, expected)
	}
}
