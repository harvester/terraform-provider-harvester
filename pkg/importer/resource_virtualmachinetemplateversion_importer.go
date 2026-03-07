package importer

import (
	"strings"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceVirtualMachineTemplateVersionStateGetter(obj *harvsterv1.VirtualMachineTemplateVersion) (*StateGetter, error) {
	vm := buildSyntheticVM(obj)
	vmImporter := NewVMImporter(vm, nil)

	states := buildVersionMetadata(obj)
	addVMSpecStates(states, vm, vmImporter)

	if err := addVMDeviceStates(states, vmImporter); err != nil {
		return nil, err
	}

	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeVirtualMachineTemplateVersion,
		States:       states,
	}, nil
}

// buildSyntheticVM constructs a VirtualMachine from the template version's VM spec
// so we can reuse VMImporter for reading disk/network/cloudinit/etc.
func buildSyntheticVM(obj *harvsterv1.VirtualMachineTemplateVersion) *kubevirtv1.VirtualMachine {
	vm := &kubevirtv1.VirtualMachine{
		Spec: obj.Spec.VM.Spec,
	}
	vm.Namespace = obj.Namespace
	if vm.Spec.Template == nil {
		vm.Spec.Template = &kubevirtv1.VirtualMachineInstanceTemplateSpec{}
	}
	vm.Spec.Template.ObjectMeta = obj.Spec.VM.ObjectMeta

	// VolumeClaimTemplates is stored on the version's own annotations (not nested VM ObjectMeta)
	// because the K8s API strips annotations from nested metav1.ObjectMeta fields.
	if vct, ok := obj.Annotations[harvesterutil.AnnotationVolumeClaimTemplates]; ok {
		vm.Annotations = map[string]string{
			harvesterutil.AnnotationVolumeClaimTemplates: vct,
		}
	}
	return vm
}

// buildVersionMetadata populates version-specific fields (template ID, image, key pairs, labels/tags).
func buildVersionMetadata(obj *harvsterv1.VirtualMachineTemplateVersion) map[string]interface{} {
	// Filter server-side labels (template.harvesterhci.io/) from user labels
	labels := GetLabels(obj.Labels)
	for key := range labels {
		if strings.HasPrefix(key, "template."+builder.LabelAnnotationPrefixHarvester) {
			delete(labels, key)
		}
	}

	states := map[string]interface{}{
		constants.FieldCommonNamespace:                         obj.Namespace,
		constants.FieldCommonName:                              obj.Name,
		constants.FieldCommonDescription:                       GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                              GetTags(obj.Labels),
		constants.FieldCommonLabels:                            labels,
		constants.FieldVirtualMachineTemplateVersionTemplateID: obj.Spec.TemplateID,
		constants.FieldVirtualMachineTemplateVersionImageID:    obj.Spec.ImageID,
		constants.FieldVirtualMachineTemplateVersionVersion:    obj.Status.Version,
	}

	// Key pair IDs — nil when empty to avoid [] vs null drift
	var keyPairIDs []string
	if len(obj.Spec.KeyPairIDs) > 0 {
		keyPairIDs = obj.Spec.KeyPairIDs
	}
	states[constants.FieldVirtualMachineTemplateVersionKeyPairIDs] = keyPairIDs

	return states
}

// addVMSpecStates adds VM spec fields (CPU, memory, machine type, run strategy, etc.) to the state map.
func addVMSpecStates(states map[string]interface{}, vm *kubevirtv1.VirtualMachine, vmImporter *VMImporter) {
	states[constants.FieldVirtualMachineCPU] = vmImporter.CPU()
	states[constants.FieldVirtualMachineCPUModel] = vmImporter.CPUModel()
	states[constants.FieldVirtualMachineMemory] = vmImporter.Memory()
	states[constants.FieldVirtualMachineRequests] = vmImporter.Requests()
	states[constants.FieldVirtualMachineHostname] = vmImporter.HostName()
	states[constants.FieldVirtualMachineReservedMemory] = vmImporter.ReservedMemory()
	states[constants.FieldVirtualMachineEFI] = vmImporter.EFI()
	states[constants.FieldVirtualMachineSecureBoot] = vmImporter.SecureBoot()
	states[constants.FieldVirtualMachineCPUPinning] = vmImporter.DedicatedCPUPlacement()
	states[constants.FieldVirtualMachineIsolateEmulatorThread] = vmImporter.IsolateEmulatorThread()

	machineType := ""
	if vm.Spec.Template != nil && vm.Spec.Template.Spec.Domain.Machine != nil {
		machineType = vm.Spec.Template.Spec.Domain.Machine.Type
	}
	states[constants.FieldVirtualMachineMachineType] = machineType

	if vm.Spec.Template != nil && len(vm.Spec.Template.Spec.NodeSelector) > 0 {
		states[constants.FieldVirtualMachineNodeSelector] = vm.Spec.Template.Spec.NodeSelector
	}

	runStrategy, err := vm.RunStrategy()
	if err != nil {
		states[constants.FieldVirtualMachineRunStrategy] = string(kubevirtv1.RunStrategyRerunOnFailure)
	} else {
		states[constants.FieldVirtualMachineRunStrategy] = string(runStrategy)
	}

	sshKeys, err := vmImporter.SSHKeys()
	if err != nil || len(sshKeys) == 0 {
		sshKeys = nil
	}
	states[constants.FieldVirtualMachineSSHKeys] = sshKeys
}

// addVMDeviceStates adds disk, network, cloud-init, input and TPM states.
func addVMDeviceStates(states map[string]interface{}, vmImporter *VMImporter) error {
	allDisks, cloudInit, err := vmImporter.Volume()
	if err != nil {
		return err
	}
	// Filter out cloudinitdisk — it's handled by the cloudinit block
	var disks []map[string]interface{}
	for _, d := range allDisks {
		if d[constants.FieldDiskName] == builder.CloudInitDiskName {
			continue
		}
		disks = append(disks, d)
	}
	states[constants.FieldVirtualMachineDisk] = disks
	states[constants.FieldVirtualMachineCloudInit] = cloudInit

	networkInterface, err := vmImporter.NetworkInterface()
	if err != nil {
		return err
	}
	states[constants.FieldVirtualMachineNetworkInterface] = networkInterface

	input, err := vmImporter.Input()
	if err != nil {
		return err
	}
	states[constants.FieldVirtualMachineInput] = input

	states[constants.FieldVirtualMachineTPM] = vmImporter.TPM()
	return nil
}
