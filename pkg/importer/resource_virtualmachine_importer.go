package importer

import (
	"encoding/json"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

type VMImporter struct {
	VirtualMachine         *kubevirtv1.VirtualMachine
	VirtualMachineInstance *kubevirtv1.VirtualMachineInstance
}

func (v *VMImporter) Name() string {
	return v.VirtualMachine.Name
}

func (v *VMImporter) Namespace() string {
	return v.VirtualMachine.Namespace
}

func (v *VMImporter) MachineType() string {
	return v.VirtualMachine.Spec.Template.Spec.Domain.Machine.Type
}

func (v *VMImporter) HostName() string {
	return v.VirtualMachine.Spec.Template.Spec.Hostname
}

func (v *VMImporter) ReservedMemory() string {
	return v.VirtualMachine.Annotations[harvesterutil.AnnotationReservedMemory]
}

func (v *VMImporter) Description() string {
	return v.VirtualMachine.Annotations[builder.AnnotationKeyDescription]
}

func (v *VMImporter) Memory() string {
	return v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits.Memory().String()
}

func (v *VMImporter) CPU() int {
	return int(v.VirtualMachine.Spec.Template.Spec.Domain.CPU.Cores)
}

func (v *VMImporter) CPUModel() string {
	return v.VirtualMachine.Spec.Template.Spec.Domain.CPU.Model
}

func (v *VMImporter) DedicatedCPUPlacement() bool {
	return bool(v.VirtualMachine.Spec.Template.Spec.Domain.CPU.DedicatedCPUPlacement)
}

func (v *VMImporter) IsolateEmulatorThread() bool {
	return bool(v.VirtualMachine.Spec.Template.Spec.Domain.CPU.IsolateEmulatorThread)
}

func (v *VMImporter) EFI() bool {
	firmware := v.VirtualMachine.Spec.Template.Spec.Domain.Firmware
	return firmware != nil && firmware.Bootloader != nil && firmware.Bootloader.EFI != nil
}

func (v *VMImporter) SecureBoot() bool {
	return v.EFI() && *v.VirtualMachine.Spec.Template.Spec.Domain.Firmware.Bootloader.EFI.SecureBoot
}

func (v *VMImporter) EvictionStrategy() bool {
	return *v.VirtualMachine.Spec.Template.Spec.EvictionStrategy == kubevirtv1.EvictionStrategyLiveMigrate
}

func (v *VMImporter) SSHKeys() ([]string, error) {
	var sshKeys []string
	sshNames := v.VirtualMachine.Spec.Template.ObjectMeta.Annotations[builder.AnnotationKeyVirtualMachineSSHNames]
	if err := json.Unmarshal([]byte(sshNames), &sshKeys); err != nil {
		return nil, err
	}
	for i, sshKey := range sshKeys {
		sshKeyNamespacedName, err := helper.RebuildNamespacedName(sshKey, v.Namespace())
		if err != nil {
			return nil, err
		}
		sshKeys[i] = sshKeyNamespacedName
	}
	return sshKeys, nil
}

func (v *VMImporter) Input() ([]map[string]interface{}, error) {
	inputs := v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Inputs
	var inputStates = make([]map[string]interface{}, 0, len(inputs))
	for _, input := range inputs {
		inputState := map[string]interface{}{
			constants.FieldInputName: input.Name,
			constants.FieldInputType: input.Type,
			constants.FieldInputBus:  input.Bus,
		}
		inputStates = append(inputStates, inputState)
	}
	return inputStates, nil
}

func (v *VMImporter) TPM() []map[string]interface{} {
	tpm := v.VirtualMachine.Spec.Template.Spec.Domain.Devices.TPM
	tpmStates := make([]map[string]interface{}, 0, 1)
	if tpm != nil {
		tpmState := map[string]interface{}{}
		tpmStates = append(tpmStates, tpmState)
	}
	return tpmStates
}

func (v *VMImporter) NetworkInterface() ([]map[string]interface{}, error) {
	var (
		waitForLeaseInterfaces   []string
		waitForLeaseInterfaceMap = map[string]struct{}{}
	)

	waitForLeaseInterfaceNames := v.VirtualMachine.Spec.Template.ObjectMeta.Annotations[builder.AnnotationKeyVirtualMachineWaitForLeaseInterfaceNames]
	if waitForLeaseInterfaceNames != "" {
		if err := json.Unmarshal([]byte(waitForLeaseInterfaceNames), &waitForLeaseInterfaces); err != nil {
			return nil, err
		}
		for _, waitForLeaseInterface := range waitForLeaseInterfaces {
			waitForLeaseInterfaceMap[waitForLeaseInterface] = struct{}{}
		}
	}

	interfaceStatusMap := map[string]kubevirtv1.VirtualMachineInstanceNetworkInterface{}
	if v.VirtualMachineInstance != nil {
		interfaceStatuses := v.VirtualMachineInstance.Status.Interfaces
		for _, interfaceStatus := range interfaceStatuses {
			interfaceStatusMap[interfaceStatus.Name] = interfaceStatus
		}
	}

	interfaces := v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Interfaces
	var networkInterfaceStates = make([]map[string]interface{}, 0, len(interfaces))
	for _, networkInterface := range interfaces {
		var interfaceType string
		if networkInterface.Bridge != nil {
			interfaceType = builder.NetworkInterfaceTypeBridge
		} else if networkInterface.Masquerade != nil {
			interfaceType = builder.NetworkInterfaceTypeMasquerade
		} else {
			return nil, fmt.Errorf("unsupported type found on network %s. ", networkInterface.Name)
		}
		var networkName string
		for _, network := range v.VirtualMachine.Spec.Template.Spec.Networks {
			if network.Name == networkInterface.Name {
				if network.Multus != nil {
					networkName = network.Multus.NetworkName
				}
				break
			}
		}

		networkInterfaceState := map[string]interface{}{
			constants.FieldNetworkInterfaceName:        networkInterface.Name,
			constants.FieldNetworkInterfaceType:        interfaceType,
			constants.FieldNetworkInterfaceModel:       networkInterface.Model,
			constants.FieldNetworkInterfaceMACAddress:  networkInterface.MacAddress,
			constants.FieldNetworkInterfaceNetworkName: networkName,
			constants.FieldNetworkInterfaceBootOrder:   networkInterface.BootOrder,
		}
		if interfaceStatus, ok := interfaceStatusMap[networkInterface.Name]; ok {
			// disregard any link-local addresses
			ips := slices.DeleteFunc(
				slices.DeleteFunc(interfaceStatus.IPs, helper.IsIPv6LinkLocal),
				helper.IsIPv4LinkLocal)
			slices.Sort(ips)
			if len(ips) > 0 {
				networkInterfaceState[constants.FieldNetworkInterfaceIPAddress] = slices.Min(ips)
				networkInterfaceState[constants.FieldNetworkInterfaceInterfaceName] = interfaceStatus.InterfaceName
			}
		}
		_, ok := waitForLeaseInterfaceMap[networkInterface.Name]
		networkInterfaceState[constants.FieldNetworkInterfaceWaitForLease] = ok
		networkInterfaceStates = append(networkInterfaceStates, networkInterfaceState)
	}
	return networkInterfaceStates, nil
}

func (v *VMImporter) pvcVolume(volume kubevirtv1.Volume, state map[string]interface{}) error {
	pvc := volume.PersistentVolumeClaim
	pvcName := pvc.ClaimName
	state[constants.FieldDiskVolumeName] = pvcName
	state[constants.FieldDiskHotPlug] = pvc.Hotpluggable
	var (
		isInPVCTemplates bool
		pvcTemplates     []*corev1.PersistentVolumeClaim
	)
	volumeClaimTemplates := v.VirtualMachine.Annotations[harvesterutil.AnnotationVolumeClaimTemplates]
	if volumeClaimTemplates != "" {
		if err := json.Unmarshal([]byte(volumeClaimTemplates), &pvcTemplates); err != nil {
			return err
		}
		for _, pvcTemplate := range pvcTemplates {
			if pvcTemplate.Name == pvcName {
				state[constants.FieldDiskSize] = pvcTemplate.Spec.Resources.Requests.Storage().String()
				if imageID := pvcTemplate.Annotations[builder.AnnotationKeyImageID]; imageID != "" {
					imageNamespacedName, err := helper.RebuildNamespacedName(imageID, v.Namespace())
					if err != nil {
						return err
					}
					state[constants.FieldVolumeImage] = imageNamespacedName
				}
				if pvcTemplate.Spec.VolumeMode != nil {
					state[constants.FieldVolumeMode] = string(*pvcTemplate.Spec.VolumeMode)
				}
				if accessModes := pvcTemplate.Spec.AccessModes; len(accessModes) > 0 {
					state[constants.FieldVolumeAccessMode] = string(pvcTemplate.Spec.AccessModes[0])
				}
				if pvcTemplate.Spec.StorageClassName != nil {
					state[constants.FieldVolumeStorageClassName] = *pvcTemplate.Spec.StorageClassName
				}
				state[constants.FieldDiskAutoDelete] = pvcTemplate.Annotations[constants.AnnotationDiskAutoDelete] == "true"
				isInPVCTemplates = true
				break
			}
		}
	}
	if !isInPVCTemplates {
		state[constants.FieldDiskExistingVolumeName] = pvcName
	}
	return nil
}

func (v *VMImporter) cloudInit(volume kubevirtv1.Volume) []map[string]interface{} {
	var cloudInitState = make([]map[string]interface{}, 0, 1)
	if volume.CloudInitNoCloud != nil {
		cloudInitState = append(cloudInitState, map[string]interface{}{
			constants.FieldCloudInitType:              builder.CloudInitTypeNoCloud,
			constants.FieldCloudInitUserData:          volume.CloudInitNoCloud.UserData,
			constants.FieldCloudInitUserDataBase64:    volume.CloudInitNoCloud.UserDataBase64,
			constants.FieldCloudInitNetworkData:       volume.CloudInitNoCloud.NetworkData,
			constants.FieldCloudInitNetworkDataBase64: volume.CloudInitNoCloud.NetworkDataBase64,
		})
		if volume.CloudInitNoCloud.UserDataSecretRef != nil {
			cloudInitState[0][constants.FieldCloudInitUserDataSecretName] = volume.CloudInitNoCloud.UserDataSecretRef.Name
		}
		if volume.CloudInitNoCloud.NetworkDataSecretRef != nil {
			cloudInitState[0][constants.FieldCloudInitNetworkDataSecretName] = volume.CloudInitNoCloud.NetworkDataSecretRef.Name
		}
	} else if volume.CloudInitConfigDrive != nil {
		cloudInitState = append(cloudInitState, map[string]interface{}{
			constants.FieldCloudInitType:              builder.CloudInitTypeConfigDrive,
			constants.FieldCloudInitUserData:          volume.CloudInitConfigDrive.UserData,
			constants.FieldCloudInitUserDataBase64:    volume.CloudInitConfigDrive.UserDataBase64,
			constants.FieldCloudInitNetworkData:       volume.CloudInitConfigDrive.NetworkData,
			constants.FieldCloudInitNetworkDataBase64: volume.CloudInitConfigDrive.NetworkDataBase64,
		})
		if volume.CloudInitConfigDrive.UserDataSecretRef != nil {
			cloudInitState[0][constants.FieldCloudInitUserDataSecretName] = volume.CloudInitConfigDrive.UserDataSecretRef.Name
		}
		if volume.CloudInitConfigDrive.NetworkDataSecretRef != nil {
			cloudInitState[0][constants.FieldCloudInitNetworkDataSecretName] = volume.CloudInitConfigDrive.NetworkDataSecretRef.Name
		}
	}
	return cloudInitState
}

func (v *VMImporter) Volume() ([]map[string]interface{}, []map[string]interface{}, error) {
	var (
		disks          = v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Disks
		volumes        = v.VirtualMachine.Spec.Template.Spec.Volumes
		volumesMap     = make(map[string]kubevirtv1.Volume, len(volumes))
		cloudInitState = make([]map[string]interface{}, 0, 1)
		diskStates     = make([]map[string]interface{}, 0, len(disks))
	)

	for _, volume := range volumes {
		volumesMap[volume.Name] = volume
	}

	for _, disk := range disks {
		diskState := make(map[string]interface{})
		var (
			diskType string
			diskBus  string
		)
		if disk.CDRom != nil {
			diskType = builder.DiskTypeCDRom
			diskBus = string(disk.CDRom.Bus)
		} else if disk.Disk != nil {
			diskType = builder.DiskTypeDisk
			diskBus = string(disk.Disk.Bus)
		} else {
			return nil, nil, fmt.Errorf("unsupported volume type found on volume %s. ", disk.Name)
		}
		diskState[constants.FieldDiskName] = disk.Name
		diskState[constants.FieldDiskBootOrder] = disk.BootOrder
		diskState[constants.FieldDiskType] = diskType
		diskState[constants.FieldDiskBus] = diskBus

		if volume, hasVolume := volumesMap[disk.Name]; hasVolume {
			if volume.CloudInitNoCloud != nil || volume.CloudInitConfigDrive != nil {
				cloudInitState = v.cloudInit(volume)
			} else {
				if volume.PersistentVolumeClaim != nil {
					if err := v.pvcVolume(volume, diskState); err != nil {
						return nil, nil, err
					}
				} else if volume.ContainerDisk != nil {
					diskState[constants.FieldDiskContainerImageName] = volume.ContainerDisk.Image
				} else {
					return nil, nil, fmt.Errorf("unsupported volume type found on volume %s. ", volume.Name)
				}
			}
		}
		diskStates = append(diskStates, diskState)
	}
	return diskStates, cloudInitState, nil
}

func (v *VMImporter) NodeName() string {
	if v.VirtualMachineInstance == nil {
		return ""
	}
	return v.VirtualMachineInstance.Status.NodeName
}

// exportLabelSelectorRequirements exports label selector requirements to Terraform state format
func exportLabelSelectorRequirements(requirements []metav1.LabelSelectorRequirement) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(requirements))
	for _, req := range requirements {
		values := make([]interface{}, 0, len(req.Values))
		for _, v := range req.Values {
			values = append(values, v)
		}
		result = append(result, map[string]interface{}{
			constants.FieldExpressionKey:      req.Key,
			constants.FieldExpressionOperator: string(req.Operator),
			constants.FieldExpressionValues:   values,
		})
	}
	return result
}

// exportLabelSelector exports a label selector to Terraform state format
func exportLabelSelector(selector *metav1.LabelSelector) []map[string]interface{} {
	if selector == nil {
		return nil
	}
	result := map[string]interface{}{}

	if len(selector.MatchLabels) > 0 {
		matchLabels := make(map[string]interface{})
		for k, v := range selector.MatchLabels {
			matchLabels[k] = v
		}
		result[constants.FieldMatchLabels] = matchLabels
	}

	if len(selector.MatchExpressions) > 0 {
		result[constants.FieldMatchExpressions] = exportLabelSelectorRequirements(selector.MatchExpressions)
	}

	if len(result) == 0 {
		return nil
	}
	return []map[string]interface{}{result}
}

// exportNodeSelectorRequirements exports node selector requirements to Terraform state format
func exportNodeSelectorRequirements(requirements []corev1.NodeSelectorRequirement) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(requirements))
	for _, req := range requirements {
		values := make([]interface{}, 0, len(req.Values))
		for _, v := range req.Values {
			values = append(values, v)
		}
		result = append(result, map[string]interface{}{
			constants.FieldExpressionKey:      req.Key,
			constants.FieldExpressionOperator: string(req.Operator),
			constants.FieldExpressionValues:   values,
		})
	}
	return result
}

// exportNodeSelectorTerms exports node selector terms to Terraform state format
func exportNodeSelectorTerms(terms []corev1.NodeSelectorTerm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(terms))
	for _, term := range terms {
		termMap := map[string]interface{}{}
		if len(term.MatchExpressions) > 0 {
			termMap[constants.FieldMatchExpressions] = exportNodeSelectorRequirements(term.MatchExpressions)
		}
		if len(term.MatchFields) > 0 {
			termMap[constants.FieldMatchFields] = exportNodeSelectorRequirements(term.MatchFields)
		}
		result = append(result, termMap)
	}
	return result
}

// exportPodAffinityTerms exports pod affinity terms to Terraform state format
func exportPodAffinityTerms(terms []corev1.PodAffinityTerm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(terms))
	for _, term := range terms {
		termMap := map[string]interface{}{
			constants.FieldTopologyKey: term.TopologyKey,
		}
		if term.LabelSelector != nil {
			termMap[constants.FieldLabelSelector] = exportLabelSelector(term.LabelSelector)
		}
		if len(term.Namespaces) > 0 {
			namespaces := make([]interface{}, 0, len(term.Namespaces))
			for _, ns := range term.Namespaces {
				namespaces = append(namespaces, ns)
			}
			termMap[constants.FieldNamespaces] = namespaces
		}
		if term.NamespaceSelector != nil {
			termMap[constants.FieldNamespaceSelector] = exportLabelSelector(term.NamespaceSelector)
		}
		result = append(result, termMap)
	}
	return result
}

// exportWeightedPodAffinityTerms exports weighted pod affinity terms to Terraform state format
func exportWeightedPodAffinityTerms(terms []corev1.WeightedPodAffinityTerm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(terms))
	for _, term := range terms {
		termMap := map[string]interface{}{
			constants.FieldPreferredWeight: int(term.Weight),
			constants.FieldPodAffinityTerm: exportPodAffinityTerms([]corev1.PodAffinityTerm{term.PodAffinityTerm}),
		}
		result = append(result, termMap)
	}
	return result
}

// NodeAffinity exports node affinity to Terraform state format
func (v *VMImporter) NodeAffinity() []map[string]interface{} {
	affinity := v.VirtualMachine.Spec.Template.Spec.Affinity
	if affinity == nil || affinity.NodeAffinity == nil {
		return nil
	}
	nodeAffinity := affinity.NodeAffinity
	result := map[string]interface{}{}

	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		terms := exportNodeSelectorTerms(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
		if len(terms) > 0 {
			result[constants.FieldNodeAffinityRequired] = []map[string]interface{}{
				{constants.FieldNodeSelectorTerm: terms},
			}
		}
	}

	if len(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		preferred := make([]map[string]interface{}, 0, len(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
		for _, pref := range nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			prefMap := map[string]interface{}{
				constants.FieldPreferredWeight:     int(pref.Weight),
				constants.FieldPreferredPreference: exportNodeSelectorTerms([]corev1.NodeSelectorTerm{pref.Preference}),
			}
			preferred = append(preferred, prefMap)
		}
		result[constants.FieldNodeAffinityPreferred] = preferred
	}

	if len(result) == 0 {
		return nil
	}
	return []map[string]interface{}{result}
}

// PodAffinity exports pod affinity to Terraform state format
func (v *VMImporter) PodAffinity() []map[string]interface{} {
	affinity := v.VirtualMachine.Spec.Template.Spec.Affinity
	if affinity == nil || affinity.PodAffinity == nil {
		return nil
	}
	podAffinity := affinity.PodAffinity
	result := map[string]interface{}{}

	if len(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		result[constants.FieldPodAffinityRequired] = exportPodAffinityTerms(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	}

	if len(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		result[constants.FieldPodAffinityPreferred] = exportWeightedPodAffinityTerms(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
	}

	if len(result) == 0 {
		return nil
	}
	return []map[string]interface{}{result}
}

// PodAntiAffinity exports pod anti-affinity to Terraform state format
func (v *VMImporter) PodAntiAffinity() []map[string]interface{} {
	affinity := v.VirtualMachine.Spec.Template.Spec.Affinity
	if affinity == nil || affinity.PodAntiAffinity == nil {
		return nil
	}
	podAntiAffinity := affinity.PodAntiAffinity
	result := map[string]interface{}{}

	if len(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		result[constants.FieldPodAffinityRequired] = exportPodAffinityTerms(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	}

	if len(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		result[constants.FieldPodAffinityPreferred] = exportWeightedPodAffinityTerms(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
	}

	if len(result) == 0 {
		return nil
	}
	return []map[string]interface{}{result}
}

func (v *VMImporter) State(networkInterfaces []map[string]interface{}, oldInstanceUID string) string {
	if v.VirtualMachineInstance == nil {
		return constants.StateVirtualMachineStopped
	}
	switch v.VirtualMachineInstance.Status.Phase {
	case "Pending", "Scheduling", "Scheduled":
		return constants.StateVirtualMachineStarting
	case "Running":
		if string(v.VirtualMachineInstance.UID) == oldInstanceUID {
			return constants.StateVirtualMachineRunning
		}
		for _, networkInterface := range networkInterfaces {
			if networkInterface[constants.FieldNetworkInterfaceWaitForLease].(bool) && networkInterface[constants.FieldNetworkInterfaceIPAddress] == "" {
				return constants.StateVirtualMachineRunning
			}
		}
		return constants.StateCommonReady
	case "Succeeded":
		return constants.StateVirtualMachineStopping
	case "Failed":
		return constants.StateCommonFailed
	default:
		return constants.StateCommonUnknown
	}
}

func NewVMImporter(vm *kubevirtv1.VirtualMachine, vmi *kubevirtv1.VirtualMachineInstance) *VMImporter {
	return &VMImporter{
		VirtualMachine:         vm,
		VirtualMachineInstance: vmi,
	}
}

func ResourceVirtualMachineStateGetter(vm *kubevirtv1.VirtualMachine, vmi *kubevirtv1.VirtualMachineInstance, oldInstanceUID string) (*StateGetter, error) {
	vmImporter := NewVMImporter(vm, vmi)
	networkInterface, err := vmImporter.NetworkInterface()
	if err != nil {
		return nil, err
	}
	disk, cloudInit, err := vmImporter.Volume()
	if err != nil {
		return nil, err
	}
	input, err := vmImporter.Input()
	if err != nil {
		return nil, err
	}
	sshKeys, err := vmImporter.SSHKeys()
	if err != nil {
		return nil, err
	}
	runStrategy, err := vm.RunStrategy()
	if err != nil {
		return nil, err
	}
	return &StateGetter{
		ID:           helper.BuildID(vm.Namespace, vm.Name),
		Name:         vm.Name,
		ResourceType: constants.ResourceTypeVirtualMachine,
		States: map[string]interface{}{
			constants.FieldCommonNamespace:                     vm.Namespace,
			constants.FieldCommonName:                          vm.Name,
			constants.FieldCommonDescription:                   GetDescriptions(vm.Annotations),
			constants.FieldCommonTags:                          GetTags(vm.Labels),
			constants.FieldCommonLabels:                        GetLabels(vm.Labels),
			constants.FieldCommonState:                         vmImporter.State(networkInterface, oldInstanceUID),
			constants.FieldVirtualMachineCPU:                   vmImporter.CPU(),
			constants.FieldVirtualMachineCPUModel:              vmImporter.CPUModel(),
			constants.FieldVirtualMachineMemory:                vmImporter.Memory(),
			constants.FieldVirtualMachineHostname:              vmImporter.HostName(),
			constants.FieldVirtualMachineReservedMemory:        vmImporter.ReservedMemory(),
			constants.FieldVirtualMachineMachineType:           vmImporter.MachineType(),
			constants.FieldVirtualMachineRunStrategy:           string(runStrategy),
			constants.FieldVirtualMachineNetworkInterface:      networkInterface,
			constants.FieldVirtualMachineDisk:                  disk,
			constants.FieldVirtualMachineInput:                 input,
			constants.FieldVirtualMachineTPM:                   vmImporter.TPM(),
			constants.FieldVirtualMachineCloudInit:             cloudInit,
			constants.FieldVirtualMachineSSHKeys:               sshKeys,
			constants.FieldVirtualMachineInstanceNodeName:      vmImporter.NodeName(),
			constants.FieldVirtualMachineEFI:                   vmImporter.EFI(),
			constants.FieldVirtualMachineSecureBoot:            vmImporter.SecureBoot(),
			constants.FieldVirtualMachineCPUPinning:            vmImporter.DedicatedCPUPlacement(),
			constants.FieldVirtualMachineIsolateEmulatorThread: vmImporter.IsolateEmulatorThread(),
			constants.FieldVirtualMachineNodeSelector:          vm.Spec.Template.Spec.NodeSelector,
			constants.FieldVirtualMachineNodeAffinity:          vmImporter.NodeAffinity(),
			constants.FieldVirtualMachinePodAffinity:           vmImporter.PodAffinity(),
			constants.FieldVirtualMachinePodAntiAffinity:       vmImporter.PodAntiAffinity(),
		},
	}, nil
}
