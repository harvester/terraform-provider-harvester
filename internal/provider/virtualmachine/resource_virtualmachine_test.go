package virtualmachine

import (
	"slices"
	"testing"

	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestRemovedAutoDeletePVCs(t *testing.T) {
	const (
		rootPVC = "vm-rootdisk-abcde"
		dataPVC = "vm-data-fghij"
		isoPVC  = "vm-cd1-klmno"
	)
	disk := func(pvcName string, autoDelete bool) any {
		return map[string]any{
			constants.FieldDiskVolumeName: pvcName,
			constants.FieldDiskAutoDelete: autoDelete,
		}
	}
	vm := func(pvcNames ...string) *kubevirtv1.VirtualMachine {
		volumes := make([]kubevirtv1.Volume, 0, len(pvcNames))
		for _, pvcName := range pvcNames {
			volumes = append(volumes, kubevirtv1.Volume{
				VolumeSource: kubevirtv1.VolumeSource{
					PersistentVolumeClaim: &kubevirtv1.PersistentVolumeClaimVolumeSource{},
				},
			})
			volumes[len(volumes)-1].PersistentVolumeClaim.ClaimName = pvcName
		}
		return &kubevirtv1.VirtualMachine{Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{Volumes: volumes},
			},
		}}
	}

	testcases := []struct {
		name     string
		oldDisks []any
		vm       *kubevirtv1.VirtualMachine
		expected []string
	}{
		{
			name:     "disks still in use",
			oldDisks: []any{disk(rootPVC, true), disk(dataPVC, true)},
			vm:       vm(rootPVC, dataPVC),
		},
		{
			name:     "removed disk with auto_delete",
			oldDisks: []any{disk(rootPVC, true), disk(dataPVC, true)},
			vm:       vm(rootPVC),
			expected: []string{dataPVC},
		},
		{
			name:     "removed disk without auto_delete",
			oldDisks: []any{disk(rootPVC, true), disk(dataPVC, false)},
			vm:       vm(rootPVC),
		},
		{
			name:     "image ejected from a cd-rom",
			oldDisks: []any{disk(rootPVC, false), disk(isoPVC, true)},
			vm:       vm(rootPVC),
			expected: []string{isoPVC},
		},
		{
			name:     "several removed disks",
			oldDisks: []any{disk(rootPVC, false), disk(dataPVC, true), disk(isoPVC, true)},
			vm:       vm(rootPVC),
			expected: []string{dataPVC, isoPVC},
		},
		{
			name:     "disk without volume",
			oldDisks: []any{disk("", true)},
			vm:       vm(),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := removedAutoDeletePVCs(tc.oldDisks, tc.vm); !slices.Equal(got, tc.expected) {
				t.Errorf("removedAutoDeletePVCs() = %v, want %v", got, tc.expected)
			}
		})
	}
}
