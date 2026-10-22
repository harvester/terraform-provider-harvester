package virtualmachine

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
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

func TestDeleteRemovedPVCs(t *testing.T) {
	const (
		namespace = "default"
		dataPVC   = "vm-data-fghij"
		isoPVC    = "vm-cd1-klmno"
	)
	errRefused := errors.New("refused by a webhook")
	pvc := func(name string) runtime.Object {
		return &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	}

	testcases := []struct {
		name             string
		existing         []runtime.Object
		refused          string // PVC whose deletion fails
		expectedWarnings int
		expectedLeft     []string
	}{
		{
			name:     "every PVC is deleted",
			existing: []runtime.Object{pvc(dataPVC), pvc(isoPVC)},
		},
		{
			name:     "a PVC already gone is not a warning",
			existing: []runtime.Object{pvc(isoPVC)},
		},
		{
			name:             "a refused deletion is a warning and the other PVC is still deleted",
			existing:         []runtime.Object{pvc(dataPVC), pvc(isoPVC)},
			refused:          dataPVC,
			expectedWarnings: 1,
			expectedLeft:     []string{dataPVC},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tc.existing...)
			client.PrependReactor("delete", "persistentvolumeclaims", func(action k8stesting.Action) (bool, runtime.Object, error) {
				if action.(k8stesting.DeleteAction).GetName() == tc.refused {
					return true, nil, errRefused
				}
				return false, nil, nil
			})
			pvcs := client.CoreV1().PersistentVolumeClaims(namespace)

			warnings := deleteRemovedPVCs(context.Background(), pvcs, []string{dataPVC, isoPVC})
			if len(warnings) != tc.expectedWarnings {
				t.Fatalf("warnings = %v, want %d", warnings, tc.expectedWarnings)
			}
			for _, warning := range warnings {
				if warning.Severity != diag.Warning {
					t.Errorf("severity = %v, want a warning", warning.Severity)
				}
			}
			list, err := pvcs.List(context.Background(), metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var left []string
			for _, item := range list.Items {
				left = append(left, item.Name)
			}
			if !slices.Equal(left, tc.expectedLeft) {
				t.Errorf("PVCs left = %v, want %v", left, tc.expectedLeft)
			}
		})
	}
}
