package importer

import (
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/builder"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func newTestVersion() *harvsterv1.VirtualMachineTemplateVersion {
	runStrategy := kubevirtv1.RunStrategyRerunOnFailure
	return &harvsterv1.VirtualMachineTemplateVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template-v1",
			Namespace: "default",
			Labels: map[string]string{
				builder.LabelPrefixHarvesterTag + "env": "prod",
			},
			Annotations: map[string]string{
				builder.AnnotationKeyDescription: "Version 1",
			},
		},
		Spec: harvsterv1.VirtualMachineTemplateVersionSpec{
			TemplateID: "default/test-template",
			ImageID:    "default/image-abc",
			KeyPairIDs: []string{"default/my-key"},
			VM: harvsterv1.VirtualMachineSourceSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						builder.AnnotationKeyVirtualMachineSSHNames: `[]`,
					},
				},
				Spec: kubevirtv1.VirtualMachineSpec{
					RunStrategy: &runStrategy,
					Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								builder.AnnotationKeyVirtualMachineSSHNames: `[]`,
							},
						},
						Spec: kubevirtv1.VirtualMachineInstanceSpec{
							Domain: kubevirtv1.DomainSpec{
								CPU: &kubevirtv1.CPU{
									Cores: 2,
									Model: "host-model",
								},
								Resources: kubevirtv1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("4Gi"),
									},
								},
								Machine: &kubevirtv1.Machine{
									Type: "q35",
								},
								Devices: kubevirtv1.Devices{
									Interfaces: []kubevirtv1.Interface{
										{
											Name: "nic-1",
											InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
												Bridge: &kubevirtv1.InterfaceBridge{},
											},
											Model: "virtio",
										},
									},
									Disks: []kubevirtv1.Disk{
										{
											Name: "rootdisk",
											DiskDevice: kubevirtv1.DiskDevice{
												Disk: &kubevirtv1.DiskTarget{
													Bus: "virtio",
												},
											},
										},
									},
								},
							},
							Networks: []kubevirtv1.Network{
								{
									Name: "nic-1",
									NetworkSource: kubevirtv1.NetworkSource{
										Multus: &kubevirtv1.MultusNetwork{
											NetworkName: "default/production",
										},
									},
								},
							},
							Volumes: []kubevirtv1.Volume{
								{
									Name: "rootdisk",
									VolumeSource: kubevirtv1.VolumeSource{
										ContainerDisk: &kubevirtv1.ContainerDiskSource{
											Image: "test-image:latest",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: harvsterv1.VirtualMachineTemplateVersionStatus{
			Version: 1,
		},
	}
}

func TestVirtualMachineTemplateVersionStateGetter(t *testing.T) {
	version := newTestVersion()
	stateGetter, err := ResourceVirtualMachineTemplateVersionStateGetter(version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateGetter.ID != "default/test-template-v1" {
		t.Errorf("ID = %q, want %q", stateGetter.ID, "default/test-template-v1")
	}
	if stateGetter.ResourceType != constants.ResourceTypeVirtualMachineTemplateVersion {
		t.Errorf("ResourceType = %q, want %q", stateGetter.ResourceType, constants.ResourceTypeVirtualMachineTemplateVersion)
	}

	t.Run("metadata", func(t *testing.T) {
		assertState(t, stateGetter, constants.FieldVirtualMachineTemplateVersionTemplateID, "default/test-template")
		assertState(t, stateGetter, constants.FieldVirtualMachineTemplateVersionImageID, "default/image-abc")
		assertState(t, stateGetter, constants.FieldVirtualMachineTemplateVersionVersion, 1)

		keyPairIDs := stateGetter.States[constants.FieldVirtualMachineTemplateVersionKeyPairIDs].([]string)
		if len(keyPairIDs) != 1 || keyPairIDs[0] != "default/my-key" {
			t.Errorf("KeyPairIDs = %v, want [default/my-key]", keyPairIDs)
		}

		tags := stateGetter.States[constants.FieldCommonTags].(map[string]string)
		if tags["env"] != "prod" {
			t.Errorf("Tags[env] = %q, want %q", tags["env"], "prod")
		}
	})

	t.Run("vm_spec", func(t *testing.T) {
		assertState(t, stateGetter, constants.FieldVirtualMachineCPU, 2)
		assertState(t, stateGetter, constants.FieldVirtualMachineCPUModel, "host-model")
		assertState(t, stateGetter, constants.FieldVirtualMachineMemory, "4Gi")
		assertState(t, stateGetter, constants.FieldVirtualMachineMachineType, "q35")
		assertState(t, stateGetter, constants.FieldVirtualMachineRunStrategy, "RerunOnFailure")
	})

	t.Run("devices", func(t *testing.T) {
		networkInterfaces := stateGetter.States[constants.FieldVirtualMachineNetworkInterface].([]map[string]interface{})
		if len(networkInterfaces) != 1 {
			t.Fatalf("NetworkInterfaces count = %d, want 1", len(networkInterfaces))
		}
		if networkInterfaces[0][constants.FieldNetworkInterfaceName] != "nic-1" {
			t.Errorf("NetworkInterface name = %q, want %q", networkInterfaces[0][constants.FieldNetworkInterfaceName], "nic-1")
		}
		if networkInterfaces[0][constants.FieldNetworkInterfaceNetworkName] != "default/production" {
			t.Errorf("NetworkInterface network = %q, want %q", networkInterfaces[0][constants.FieldNetworkInterfaceNetworkName], "default/production")
		}

		disks := stateGetter.States[constants.FieldVirtualMachineDisk].([]map[string]interface{})
		if len(disks) != 1 {
			t.Fatalf("Disks count = %d, want 1", len(disks))
		}
		if disks[0][constants.FieldDiskName] != "rootdisk" {
			t.Errorf("Disk name = %q, want %q", disks[0][constants.FieldDiskName], "rootdisk")
		}
		if disks[0][constants.FieldDiskContainerImageName] != "test-image:latest" {
			t.Errorf("Disk container image = %q, want %q", disks[0][constants.FieldDiskContainerImageName], "test-image:latest")
		}
	})
}

func TestVirtualMachineTemplateVersionStateGetterMinimal(t *testing.T) {
	version := &harvsterv1.VirtualMachineTemplateVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "minimal-version",
			Namespace:   "default",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: harvsterv1.VirtualMachineTemplateVersionSpec{
			TemplateID: "default/minimal-template",
			VM: harvsterv1.VirtualMachineSourceSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						builder.AnnotationKeyVirtualMachineSSHNames: `[]`,
					},
				},
				Spec: kubevirtv1.VirtualMachineSpec{
					Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								builder.AnnotationKeyVirtualMachineSSHNames: `[]`,
							},
						},
						Spec: kubevirtv1.VirtualMachineInstanceSpec{
							Domain: kubevirtv1.DomainSpec{
								CPU: &kubevirtv1.CPU{
									Cores: 1,
								},
								Resources: kubevirtv1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
								Devices: kubevirtv1.Devices{
									Interfaces: []kubevirtv1.Interface{},
									Disks:      []kubevirtv1.Disk{},
								},
							},
						},
					},
				},
			},
		},
		Status: harvsterv1.VirtualMachineTemplateVersionStatus{},
	}

	stateGetter, err := ResourceVirtualMachineTemplateVersionStateGetter(version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateGetter.ID != "default/minimal-version" {
		t.Errorf("ID = %q, want %q", stateGetter.ID, "default/minimal-version")
	}
	assertState(t, stateGetter, constants.FieldVirtualMachineCPU, 1)
	assertState(t, stateGetter, constants.FieldVirtualMachineMemory, "1Gi")
	assertState(t, stateGetter, constants.FieldVirtualMachineTemplateVersionVersion, 0)

	keyPairIDs := stateGetter.States[constants.FieldVirtualMachineTemplateVersionKeyPairIDs].([]string)
	if len(keyPairIDs) != 0 {
		t.Errorf("KeyPairIDs = %v, want empty", keyPairIDs)
	}
}

func assertState(t *testing.T, sg *StateGetter, key string, want interface{}) {
	t.Helper()
	if got := sg.States[key]; got != want {
		t.Errorf("%s = %v, want %v", key, got, want)
	}
}
