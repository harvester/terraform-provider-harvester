package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	vmCreator = "terraform-provider-harvester"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Builder *builder.VMBuilder
	Client  *client.Client
}

func (c *Constructor) Setup() util.Processors {
	vmBuilder := c.Builder
	if vmBuilder == nil {
		return nil
	}
	processors := util.NewProcessors().Tags(&c.Builder.VirtualMachine.Labels).Description(&c.Builder.VirtualMachine.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldVirtualMachineCPU,
			Parser: func(i interface{}) error {
				vmBuilder.CPU(i.(int))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineMemory,
			Parser: func(i interface{}) error {
				vmBuilder.Memory(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineRunStrategy,
			Parser: func(i interface{}) error {
				runStrategy := kubevirtv1.VirtualMachineRunStrategy(i.(string))
				vmBuilder.RunStrategy(runStrategy)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineStart,
			Parser: func(i interface{}) error {
				vmBuilder.Run(i.(bool))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineMachineType,
			Parser: func(i interface{}) error {
				vmBuilder.MachineType(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineHostname,
			Parser: func(i interface{}) error {
				vmBuilder.HostName(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineSSHKeys,
			Parser: func(i interface{}) error {
				sshKey := i.(string)
				sshKeyNamespacedName, err := helper.RebuildNamespacedName(sshKey, c.Builder.VirtualMachine.Namespace)
				if err != nil {
					return err
				}
				vmBuilder.SSHKey(sshKeyNamespacedName)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineNetworkInterface,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				interfaceName := r[constants.FiledNetworkInterfaceName].(string)
				interfaceType := r[constants.FiledNetworkInterfaceType].(string)
				interfaceModel := r[constants.FiledNetworkInterfaceModel].(string)
				interfaceMACAddress := r[constants.FiledNetworkInterfaceMACAddress].(string)
				networkName := r[constants.FiledNetworkInterfaceNetworkName].(string)
				if interfaceType == "" {
					if networkName == "" {
						interfaceType = builder.NetworkInterfaceTypeMasquerade
					} else {
						interfaceType = builder.NetworkInterfaceTypeBridge
					}
				}
				vmBuilder.NetworkInterface(interfaceName, interfaceModel, interfaceMACAddress, interfaceType, networkName)
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVirtualMachineDisk,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				diskName := r[constants.FieldDiskName].(string)
				diskSize := r[constants.FieldDiskSize].(string)
				diskBus := r[constants.FieldDiskBus].(string)
				diskType := r[constants.FieldDiskType].(string)
				bootOrder := r[constants.FieldDiskBootOrder].(int)
				imageNamespacedName := r[constants.FieldVolumeImage].(string)
				volumeName := r[constants.FieldDiskVolumeName].(string)
				existingVolumeName := r[constants.FieldDiskExistingVolumeName].(string)
				containerImageName := r[constants.FieldDiskContainerImageName].(string)
				hotPlug := r[constants.FieldDiskHotPlug].(bool)
				isCDRom := diskType == builder.DiskTypeCDRom
				if diskBus == "" {
					if isCDRom {
						diskBus = builder.DiskBusSata
					} else if hotPlug {
						diskBus = builder.DiskBusScsi
					} else {
						diskBus = builder.DiskBusVirtio
					}
				}
				if diskSize == "" {
					diskSize = util.If(existingVolumeName == "", "", builder.DefaultDiskSize).(string)
				}
				vmBuilder.Disk(diskName, diskBus, isCDRom, bootOrder)
				if existingVolumeName != "" {
					vmBuilder.ExistingPVCVolume(diskName, existingVolumeName, hotPlug)
				} else if containerImageName != "" {
					vmBuilder.ContainerDiskVolume(diskName, containerImageName, builder.DefaultImagePullPolicy)
				} else {
					pvcOption := &builder.PersistentVolumeClaimOption{
						VolumeMode: corev1.PersistentVolumeBlock,
						AccessMode: corev1.ReadWriteMany,
					}
					if storageClassName := r[constants.FieldVolumeStorageClassName].(string); storageClassName != "" {
						pvcOption.StorageClassName = pointer.StringPtr(storageClassName)
					}
					if volumeMode := r[constants.FieldVolumeMode].(string); volumeMode != "" {
						pvcOption.VolumeMode = corev1.PersistentVolumeMode(volumeMode)
					}
					if accessMode := r[constants.FieldVolumeAccessMode].(string); accessMode != "" {
						pvcOption.AccessMode = corev1.PersistentVolumeAccessMode(accessMode)
					}
					if imageNamespacedName != "" {
						imageNamespace, imageName, err := helper.NamespacedNamePartsByDefault(imageNamespacedName, c.Builder.VirtualMachine.Namespace)
						if err != nil {
							return err
						}
						pvcOption.ImageID = helper.BuildNamespacedName(imageNamespace, imageName)
						storageClassName := builder.BuildImageStorageClassName("", imageName)
						pvcOption.StorageClassName = pointer.StringPtr(storageClassName)
					}
					if autoDelete := r[constants.FieldDiskAutoDelete].(bool); autoDelete {
						pvcOption.Annotations = map[string]string{
							constants.AnnotationDiskAutoDelete: "true",
						}
					}
					vmBuilder.PVCVolume(diskName, diskSize, volumeName, hotPlug, pvcOption)
				}
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVirtualMachineCloudInit,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				cloudInitSource := builder.CloudInitSource{
					CloudInitType:         r[constants.FieldCloudInitType].(string),
					NetworkData:           r[constants.FieldCloudInitNetworkData].(string),
					NetworkDataBase64:     r[constants.FieldCloudInitNetworkDataBase64].(string),
					NetworkDataSecretName: r[constants.FieldCloudInitNetworkDataSecretName].(string),
					UserData:              r[constants.FieldCloudInitUserData].(string),
					UserDataBase64:        r[constants.FieldCloudInitUserDataBase64].(string),
					UserDataSecretName:    r[constants.FieldCloudInitUserDataSecretName].(string),
				}
				var diskBus string
				isCDRom := cloudInitSource.CloudInitType == builder.CloudInitTypeConfigDrive
				if isCDRom {
					diskBus = builder.DiskBusSata
				} else {
					diskBus = builder.DiskBusVirtio
				}
				diskName := builder.CloudInitDiskName
				vmBuilder.Disk(diskName, diskBus, isCDRom, 0)
				vmBuilder.CloudInit(diskName, cloudInitSource)
				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	if len(c.Builder.SSHNames) == 0 {
		return nil
	}

	keyPairs, err := c.getKeyPairs(c.Builder.SSHNames, c.Builder.VirtualMachine.Namespace)
	if err != nil {
		return err
	}
	return c.checkKeyPairsInCloudInit(keyPairs)
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Builder.VM()
}

func newVMConstructor(c *client.Client, vmBuilder *builder.VMBuilder) util.Constructor {
	return &Constructor{
		Builder: vmBuilder,
		Client:  c,
	}
}

func Creator(c *client.Client, namespace, name string) util.Constructor {
	vmBuilder := builder.NewVMBuilder(vmCreator).
		Namespace(namespace).Name(name).
		EvictionStrategy(true).
		DefaultPodAntiAffinity()
	return newVMConstructor(c, vmBuilder)
}

func Updater(c *client.Client, vm *kubevirtv1.VirtualMachine) util.Constructor {
	vm.Spec.Template.Spec.Networks = []kubevirtv1.Network{}
	vm.Spec.Template.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{}
	vm.Spec.Template.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{}
	vm.Spec.Template.Spec.Volumes = []kubevirtv1.Volume{}
	vm.Annotations[harvesterutil.AnnotationVolumeClaimTemplates] = "[]"
	return newVMConstructor(c, &builder.VMBuilder{
		VirtualMachine: vm,
	})
}
