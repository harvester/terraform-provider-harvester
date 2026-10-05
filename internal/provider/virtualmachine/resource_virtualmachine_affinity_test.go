package virtualmachine

import (
	"context"
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
	affinityTestLabel = "app"
	affinityTestValue = "web"
	affinityTestNode  = "node1"
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
	block := map[string]interface{}{
		constants.FieldNodeAffinityRequired: []interface{}{map[string]interface{}{
			constants.FieldNodeSelectorTerm: []interface{}{
				map[string]interface{}{
					constants.FieldMatchExpressions: []interface{}{affinityExpression(corev1.LabelHostname, "In", affinityTestNode)},
					constants.FieldMatchFields:      []interface{}{affinityExpression("metadata.name", "In", affinityTestNode)},
				},
			},
		}},
		constants.FieldNodeAffinityPreferred: []interface{}{map[string]interface{}{
			constants.FieldPreferredWeight: 10,
			constants.FieldPreferredPreference: []interface{}{
				map[string]interface{}{constants.FieldMatchExpressions: []interface{}{affinityExpression("disktype", "In", "ssd")}},
			},
		}},
	}
	expected := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{
			MatchExpressions: []corev1.NodeSelectorRequirement{{Key: corev1.LabelHostname, Operator: corev1.NodeSelectorOpIn, Values: []string{affinityTestNode}}},
			MatchFields:      []corev1.NodeSelectorRequirement{{Key: "metadata.name", Operator: corev1.NodeSelectorOpIn, Values: []string{affinityTestNode}}},
		}}},
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
			Weight:     10,
			Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
		}},
	}
	got, err := parseNodeAffinity(block)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("parseNodeAffinity() = %+v, want %+v", got, expected)
	}
}

// TestParseNodeAffinityRejects covers the blocks that would crash, match every
// node, or be dropped by the Harvester webhook.
func TestParseNodeAffinityRejects(t *testing.T) {
	testcases := map[string]map[string]interface{}{
		"empty block (nil from Terraform)": blockMap(nil),
		"empty term":                       nodeAffinityBlock(nil),
		"match_fields only": nodeAffinityBlock(map[string]interface{}{
			constants.FieldMatchFields: []interface{}{affinityExpression("metadata.name", "In", affinityTestNode)},
		}),
		"empty preference": {
			constants.FieldNodeAffinityPreferred: []interface{}{map[string]interface{}{
				constants.FieldPreferredWeight:     10,
				constants.FieldPreferredPreference: []interface{}{nil},
			}},
		},
	}
	for name, block := range testcases {
		t.Run(name, func(t *testing.T) {
			if got, err := parseNodeAffinity(block); err == nil {
				t.Errorf("parseNodeAffinity() = %+v, want an error", got)
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

func TestParsePodAffinityRules(t *testing.T) {
	block := map[string]interface{}{
		constants.FieldPodAffinityRequired: []interface{}{
			podAffinityTermBlock(map[string]interface{}{constants.FieldMatchExpressions: []interface{}{affinityExpression(affinityTestLabel, "In", affinityTestValue)}}, nil),
		},
		constants.FieldPodAffinityPreferred: []interface{}{map[string]interface{}{
			constants.FieldPreferredWeight: 20,
			constants.FieldPodAffinityTerm: []interface{}{podAffinityTermBlock(map[string]interface{}{constants.FieldMatchLabels: map[string]interface{}{affinityTestLabel: "cache"}}, nil)},
		}},
	}
	required, preferred, err := parsePodAffinityRules(constants.FieldVirtualMachinePodAffinity, block)
	if err != nil {
		t.Fatal(err)
	}
	expectedRequired := []corev1.PodAffinityTerm{{
		TopologyKey:   corev1.LabelHostname,
		LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: affinityTestLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{affinityTestValue}}}},
	}}
	expectedPreferred := []corev1.WeightedPodAffinityTerm{{
		Weight: 20,
		PodAffinityTerm: corev1.PodAffinityTerm{
			TopologyKey:   corev1.LabelHostname,
			LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{affinityTestLabel: "cache"}},
		},
	}}
	if !reflect.DeepEqual(required, expectedRequired) {
		t.Errorf("required = %+v, want %+v", required, expectedRequired)
	}
	if !reflect.DeepEqual(preferred, expectedPreferred) {
		t.Errorf("preferred = %+v, want %+v", preferred, expectedPreferred)
	}
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
	required, _, err := parsePodAffinityRules(constants.FieldVirtualMachinePodAntiAffinity, block)
	if err != nil {
		t.Fatal(err)
	}
	if len(required) != 1 || !reflect.DeepEqual(required[0].NamespaceSelector, &metav1.LabelSelector{}) {
		t.Errorf("namespace selector = %+v, want an empty selector", required)
	}
}

func TestParsePodAffinityRulesRejects(t *testing.T) {
	testcases := map[string]map[string]interface{}{
		"empty block (nil from Terraform)": blockMap(nil),
		"empty label selector": {
			constants.FieldPodAffinityRequired: []interface{}{podAffinityTermBlock(map[string]interface{}{}, nil)},
		},
		"empty label selector in a preferred term": {
			constants.FieldPodAffinityPreferred: []interface{}{map[string]interface{}{
				constants.FieldPreferredWeight: 20,
				constants.FieldPodAffinityTerm: []interface{}{podAffinityTermBlock(map[string]interface{}{}, nil)},
			}},
		},
	}
	for name, block := range testcases {
		t.Run(name, func(t *testing.T) {
			if required, preferred, err := parsePodAffinityRules(constants.FieldVirtualMachinePodAffinity, block); err == nil {
				t.Errorf("parsePodAffinityRules() = %+v, %+v, want an error", required, preferred)
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
					constants.FieldMatchExpressions: []interface{}{affinityExpression("disktype", "Exists")},
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
