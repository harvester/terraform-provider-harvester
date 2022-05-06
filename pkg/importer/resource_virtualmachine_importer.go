package importer

import (
	"encoding/json"
	"fmt"

	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"
	corev1 "k8s.io/api/core/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

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

func (v *VMImporter) Description() string {
	return v.VirtualMachine.Annotations[builder.AnnotationKeyDescription]
}

func (v *VMImporter) Memory() string {
	return v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits.Memory().String()
}

func (v *VMImporter) CPU() int {
	return int(v.VirtualMachine.Spec.Template.Spec.Domain.CPU.Cores)
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

func (v *VMImporter) NetworkInterface() ([]map[string]interface{}, error) {
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
			constants.FiledNetworkInterfaceName:        networkInterface.Name,
			constants.FiledNetworkInterfaceType:        interfaceType,
			constants.FiledNetworkInterfaceModel:       networkInterface.Model,
			constants.FiledNetworkInterfaceMACAddress:  networkInterface.MacAddress,
			constants.FiledNetworkInterfaceNetworkName: networkName,
		}
		if interfaceStatus, ok := interfaceStatusMap[networkInterface.Name]; ok {
			networkInterfaceState[constants.FiledNetworkInterfaceIPAddress] = interfaceStatus.IP
			networkInterfaceState[constants.FiledNetworkInterfaceInterfaceName] = interfaceStatus.InterfaceName
		}
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
		cloudInitState = make([]map[string]interface{}, 0, 1)
		diskStates     = make([]map[string]interface{}, 0, len(disks))
	)
	for _, volume := range volumes {
		diskState := make(map[string]interface{})
		for _, disk := range disks {
			if volume.Name != disk.Name {
				continue
			}
			var (
				diskType string
				diskBus  string
			)
			if disk.CDRom != nil {
				diskType = builder.DiskTypeCDRom
				diskBus = disk.CDRom.Bus
			} else if disk.Disk != nil {
				diskType = builder.DiskTypeDisk
				diskBus = disk.Disk.Bus
			} else {
				return nil, nil, fmt.Errorf("unsupported volume type found on volume %s. ", disk.Name)
			}
			diskState[constants.FieldDiskName] = disk.Name
			diskState[constants.FieldDiskBootOrder] = disk.BootOrder
			diskState[constants.FieldDiskType] = diskType
			diskState[constants.FieldDiskBus] = diskBus
		}
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
			diskStates = append(diskStates, diskState)
		}
	}
	return diskStates, cloudInitState, nil
}

func (v *VMImporter) NodeName() string {
	if v.VirtualMachineInstance == nil {
		return ""
	}
	return v.VirtualMachineInstance.Status.NodeName
}

func (v *VMImporter) State() string {
	if v.VirtualMachineInstance == nil {
		return constants.StateVirtualMachineStopped
	}
	switch v.VirtualMachineInstance.Status.Phase {
	case "Pending", "Scheduling", "Scheduled":
		return constants.StateVirtualMachineStarting
	case "Running":
		return constants.StateVirtualMachineRunning
	case "Succeeded":
		return constants.StateVirtualMachineStopping
	case "Failed":
		return constants.StateVirtualMachineError
	default:
		return constants.StateCommonNone
	}
}

func NewVMImporter(vm *kubevirtv1.VirtualMachine, vmi *kubevirtv1.VirtualMachineInstance) *VMImporter {
	return &VMImporter{
		VirtualMachine:         vm,
		VirtualMachineInstance: vmi,
	}
}

func ResourceVirtualMachineStateGetter(vm *kubevirtv1.VirtualMachine, vmi *kubevirtv1.VirtualMachineInstance) (*StateGetter, error) {
	vmImporter := NewVMImporter(vm, vmi)
	networkInterface, err := vmImporter.NetworkInterface()
	if err != nil {
		return nil, err
	}
	disk, cloudInit, err := vmImporter.Volume()
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
			constants.FieldCommonNamespace:                vm.Namespace,
			constants.FieldCommonName:                     vm.Name,
			constants.FieldCommonDescription:              GetDescriptions(vm.Annotations),
			constants.FieldCommonTags:                     GetTags(vm.Labels),
			constants.FieldCommonState:                    vmImporter.State(),
			constants.FieldVirtualMachineCPU:              vmImporter.CPU(),
			constants.FieldVirtualMachineMemory:           vmImporter.Memory(),
			constants.FieldVirtualMachineHostname:         vmImporter.HostName(),
			constants.FieldVirtualMachineMachineType:      vmImporter.MachineType(),
			constants.FieldVirtualMachineRunStrategy:      string(runStrategy),
			constants.FieldVirtualMachineNetworkInterface: networkInterface,
			constants.FieldVirtualMachineDisk:             disk,
			constants.FieldVirtualMachineCloudInit:        cloudInit,
			constants.FieldVirtualMachineSSHKeys:          sshKeys,
			constants.FieldVirtualMachineInstanceNodeName: vmImporter.NodeName(),
		},
	}, nil
}
