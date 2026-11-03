package helper

import (
	"strings"

	networkapi "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io"
	"github.com/harvester/harvester/pkg/builder"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// CPUManagerLabel is the value of kubevirtv1.CPUManager since KubeVirt v1.8.0
// (Harvester v1.9.0). The embedded kubevirt.io/api v1.7.0 still has the former
// value, "cpumanager", which Harvester v1.8 uses: both are handled.
const CPUManagerLabel = "kubevirt.io/cpumanager"

// IsHarvesterManagedNodeSelector reports whether a requirement of a required
// node selector term is managed by the Harvester VM webhook, with the keys its
// makeAffinityFromVMTemplate treats as its own: one expression per cluster
// network the VM is attached to (network.harvesterhci.io/<name>), and the CPU
// manager label when the VM has dedicated CPUs.
func IsHarvesterManagedNodeSelector(req corev1.NodeSelectorRequirement) bool {
	return strings.HasPrefix(req.Key, networkapi.GroupName+"/") || req.Key == kubevirtv1.CPUManager || req.Key == CPUManagerLabel
}

// IsDefaultPodAntiAffinityTerm reports whether term is the preferred
// anti-affinity term that builder.DefaultPodAntiAffinity() sets on every VM
// the provider creates or updates.
func IsDefaultPodAntiAffinityTerm(term corev1.WeightedPodAffinityTerm) bool {
	defaults := builder.NewVMBuilder("").DefaultPodAntiAffinity().VirtualMachine.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	for _, d := range defaults {
		if equality.Semantic.DeepEqual(d, term) {
			return true
		}
	}
	return false
}
