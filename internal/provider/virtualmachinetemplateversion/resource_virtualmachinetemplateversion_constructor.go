package virtualmachinetemplateversion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kubevirtv1 "kubevirt.io/api/core/v1"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	vmTemplateVersionCreator = "terraform-provider-harvester"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Client  *client.Client
	Context context.Context

	Version   *harvsterv1.VirtualMachineTemplateVersion
	VMBuilder *builder.VMBuilder
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Version.Labels).
		Labels(&c.Version.Labels).
		Description(&c.Version.Annotations).
		String(constants.FieldVirtualMachineTemplateVersionTemplateID, &c.Version.Spec.TemplateID, true).
		String(constants.FieldVirtualMachineTemplateVersionImageID, &c.Version.Spec.ImageID, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldVirtualMachineTemplateVersionKeyPairIDs,
			Parser: func(i interface{}) error {
				c.Version.Spec.KeyPairIDs = append(c.Version.Spec.KeyPairIDs, i.(string))
				return nil
			},
		},
	}
	customProcessors = append(customProcessors, c.vmSpecProcessors()...)
	customProcessors = append(customProcessors, c.diskProcessor(), c.networkInterfaceProcessor(), c.cloudInitProcessor())
	customProcessors = append(customProcessors, c.deviceProcessors()...)
	return append(processors, customProcessors...)
}

func (c *Constructor) vmSpecProcessors() []util.Processor {
	return []util.Processor{
		{
			Field: constants.FieldVirtualMachineCPU,
			Parser: func(i interface{}) error {
				c.VMBuilder.CPU(i.(int))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineCPUModel,
			Parser: func(i interface{}) error {
				if cpuModel := i.(string); cpuModel != "" {
					c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.CPU.Model = cpuModel
				}
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineMemory,
			Parser: func(i interface{}) error {
				c.VMBuilder.Memory(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineRequests,
			Parser: func(i interface{}) error {
				return c.parseRequests(i.(map[string]interface{}))
			},
		},
		{
			Field:    constants.FieldVirtualMachineEFI,
			Parser:   c.parseEFI,
			Required: true,
		},
		{
			Field:  constants.FieldVirtualMachineSecureBoot,
			Parser: c.parseSecureBoot,
		},
		{
			Field: constants.FieldVirtualMachineRunStrategy,
			Parser: func(i interface{}) error {
				c.VMBuilder.RunStrategy(kubevirtv1.VirtualMachineRunStrategy(i.(string)))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineMachineType,
			Parser: func(i interface{}) error {
				c.VMBuilder.MachineType(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineHostname,
			Parser: func(i interface{}) error {
				c.VMBuilder.HostName(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineReservedMemory,
			Parser: func(i interface{}) error {
				if reservedMemory := i.(string); reservedMemory != "" {
					c.VMBuilder.Annotations(map[string]string{
						harvesterutil.AnnotationReservedMemory: reservedMemory,
					})
				} else {
					delete(c.VMBuilder.VirtualMachine.Annotations, harvesterutil.AnnotationReservedMemory)
				}
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVirtualMachineSSHKeys,
			Parser: func(i interface{}) error {
				sshKeyNamespacedName, err := helper.RebuildNamespacedName(i.(string), c.VMBuilder.VirtualMachine.Namespace)
				if err != nil {
					return err
				}
				c.VMBuilder.SSHKey(sshKeyNamespacedName)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineCPUPinning,
			Parser: func(i interface{}) error {
				c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.CPU.DedicatedCPUPlacement = i.(bool)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineIsolateEmulatorThread,
			Parser: func(i interface{}) error {
				c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.CPU.IsolateEmulatorThread = i.(bool)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineNodeSelector,
			Parser: func(i interface{}) error {
				v := i.(map[string]interface{})
				c.VMBuilder.VirtualMachine.Spec.Template.Spec.NodeSelector = make(map[string]string)
				for k, val := range v {
					c.VMBuilder.VirtualMachine.Spec.Template.Spec.NodeSelector[k] = val.(string)
				}
				return nil
			},
		},
	}
}

func (c *Constructor) parseRequests(r map[string]interface{}) error {
	requests := corev1.ResourceList{}
	if cpuStr, ok := r[constants.FieldRequestsCPU].(string); ok && cpuStr != "" {
		quantity, err := resource.ParseQuantity(cpuStr)
		if err != nil {
			return fmt.Errorf("invalid requests cpu %q: %w", cpuStr, err)
		}
		requests[corev1.ResourceCPU] = quantity
	}
	if memStr, ok := r[constants.FieldRequestsMemory].(string); ok && memStr != "" {
		quantity, err := resource.ParseQuantity(memStr)
		if err != nil {
			return fmt.Errorf("invalid requests memory %q: %w", memStr, err)
		}
		requests[corev1.ResourceMemory] = quantity
	}
	if len(requests) > 0 {
		c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Resources.Requests = requests
	}
	return nil
}

func (c *Constructor) parseEFI(i interface{}) error {
	var firmware *kubevirtv1.Firmware
	if i.(bool) {
		firmware = &kubevirtv1.Firmware{
			Bootloader: &kubevirtv1.Bootloader{
				EFI: &kubevirtv1.EFI{
					SecureBoot: ptr.To(false),
				},
			},
		}
	}
	if oldFirmware := c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware; oldFirmware != nil {
		if firmware == nil {
			firmware = &kubevirtv1.Firmware{}
		}
		firmware.UUID = oldFirmware.UUID
		firmware.Serial = oldFirmware.Serial
	}
	c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware = firmware
	return nil
}

func (c *Constructor) parseSecureBoot(i interface{}) error {
	firmware := c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware
	if firmware == nil || firmware.Bootloader == nil || firmware.Bootloader.EFI == nil {
		return errors.New("EFI must be enabled to use Secure Boot. ")
	}
	firmware.Bootloader.EFI.SecureBoot = ptr.To(true)
	c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware = firmware

	features := c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features
	if features == nil {
		features = &kubevirtv1.Features{}
	}
	features.SMM = &kubevirtv1.FeatureState{
		Enabled: ptr.To(true),
	}
	c.VMBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features = features
	return nil
}

func (c *Constructor) networkInterfaceProcessor() util.Processor {
	return util.Processor{
		Field: constants.FieldVirtualMachineNetworkInterface,
		Parser: func(i interface{}) error {
			r := i.(map[string]interface{})
			interfaceName := r[constants.FieldNetworkInterfaceName].(string)
			interfaceType := r[constants.FieldNetworkInterfaceType].(string)
			interfaceModel := r[constants.FieldNetworkInterfaceModel].(string)
			interfaceMACAddress := r[constants.FieldNetworkInterfaceMACAddress].(string)
			networkName := r[constants.FieldNetworkInterfaceNetworkName].(string)
			bootOrder := r[constants.FieldNetworkInterfaceBootOrder].(int)

			if interfaceType == "" {
				if networkName == "" {
					interfaceType = builder.NetworkInterfaceTypeMasquerade
				} else {
					interfaceType = builder.NetworkInterfaceTypeBridge
				}
			}
			c.VMBuilder.NetworkInterface(interfaceName, interfaceModel, interfaceMACAddress, interfaceType, networkName)
			if bootOrder != 0 {
				c.VMBuilder.SetNetworkInterfaceBootOrder(interfaceName, uint(bootOrder)) // nolint: gosec
			}
			return nil
		},
		Required: true,
	}
}

func (c *Constructor) diskProcessor() util.Processor {
	return util.Processor{
		Field:    constants.FieldVirtualMachineDisk,
		Parser:   c.parseDisk,
		Required: true,
	}
}

func (c *Constructor) parseDisk(i interface{}) error {
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

	c.VMBuilder.Disk(diskName, diskBus, isCDRom, uint(bootOrder)) // nolint: gosec
	if existingVolumeName != "" {
		c.VMBuilder.ExistingPVCVolume(diskName, existingVolumeName, hotPlug)
	} else if containerImageName != "" {
		c.VMBuilder.ContainerDiskVolume(diskName, containerImageName, builder.DefaultImagePullPolicy)
	} else if isCDRom && imageNamespacedName == "" {
		// Empty CDRom: don't prepare volume
	} else {
		return c.preparePVCVolume(r, diskName, diskSize, volumeName, hotPlug, imageNamespacedName)
	}
	return nil
}

func (c *Constructor) preparePVCVolume(r map[string]interface{}, diskName, diskSize, volumeName string, hotPlug bool, imageNamespacedName string) error {
	pvcOption := &builder.PersistentVolumeClaimOption{
		VolumeMode: corev1.PersistentVolumeBlock,
		AccessMode: corev1.ReadWriteMany,
	}
	storageClassName := r[constants.FieldVolumeStorageClassName].(string)
	if imageNamespacedName != "" {
		imageNamespace, imageName, err := helper.NamespacedNamePartsByDefault(imageNamespacedName, c.VMBuilder.VirtualMachine.Namespace)
		if err != nil {
			return err
		}
		vmimage, err := c.Client.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(imageNamespace).Get(c.Context, imageName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		pvcOption.ImageID = helper.BuildNamespacedName(imageNamespace, imageName)
		scName := vmimage.Status.StorageClassName
		if storageClassName == "" {
			storageClassName = scName
		} else if storageClassName != scName {
			return fmt.Errorf("the %s of an image can only be defined during image creation", constants.FieldVolumeStorageClassName)
		}
	} else if storageClassName == "" {
		storageClasses, err := c.Client.StorageClassClient.StorageClasses().List(c.Context, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, storageClass := range storageClasses.Items {
			if storageClass.Annotations[harvesterutil.AnnotationIsDefaultStorageClassName] == "true" {
				storageClassName = storageClass.Name
				break
			}
		}
	}
	pvcOption.StorageClassName = ptr.To(storageClassName)

	if volumeMode := r[constants.FieldVolumeMode].(string); volumeMode != "" {
		pvcOption.VolumeMode = corev1.PersistentVolumeMode(volumeMode)
	}
	if accessMode := r[constants.FieldVolumeAccessMode].(string); accessMode != "" {
		pvcOption.AccessMode = corev1.PersistentVolumeAccessMode(accessMode)
	}
	if autoDelete := r[constants.FieldDiskAutoDelete].(bool); autoDelete {
		pvcOption.Annotations = map[string]string{
			constants.AnnotationDiskAutoDelete: "true",
		}
	}

	_, err := resource.ParseQuantity(diskSize)
	if diskSize == "" {
		diskSize = builder.DefaultDiskSize
	} else if err != nil {
		return fmt.Errorf("\"%v\" is not a parsable quantity: %v", diskSize, err)
	}

	c.VMBuilder.PVCVolume(diskName, diskSize, volumeName, hotPlug, pvcOption)
	return nil
}

func (c *Constructor) cloudInitProcessor() util.Processor {
	return util.Processor{
		Field:  constants.FieldVirtualMachineCloudInit,
		Parser: c.parseCloudInit,
	}
}

func (c *Constructor) parseCloudInit(i interface{}) error {
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
	if cloudInitSource.UserDataBase64 == "" && cloudInitSource.UserDataSecretName == "" {
		c.enrichCloudInitUserData(&cloudInitSource)
	}
	diskName := builder.CloudInitDiskName
	c.VMBuilder.Disk(diskName, diskBus, isCDRom, 0)
	c.VMBuilder.CloudInit(diskName, cloudInitSource)
	return nil
}

func (c *Constructor) enrichCloudInitUserData(source *builder.CloudInitSource) {
	if c.VMBuilder.VirtualMachine.Labels != nil {
		if sshUsername, ok := c.VMBuilder.VirtualMachine.Labels[builder.LabelPrefixHarvesterTag+constants.LabelSSHUsername]; ok && sshUsername != "" {
			if source.UserData == "" {
				source.UserData = fmt.Sprintf("#cloud-config\nuser: %s\n", sshUsername)
			} else {
				appendUser := true
				for _, line := range strings.Split(source.UserData, "\n") {
					if strings.HasPrefix(line, "user: ") {
						appendUser = false
						break
					}
				}
				if appendUser {
					source.UserData += fmt.Sprintf("\nuser: %s\n", sshUsername)
				}
			}
		}
	}

	publicKeys := []string{}
	for _, sshName := range c.VMBuilder.SSHNames {
		_, keyPairName, err := helper.NamespacedNameParts(sshName)
		if err != nil {
			continue
		}
		keyPair, err := c.Client.HarvesterClient.HarvesterhciV1beta1().KeyPairs(c.VMBuilder.VirtualMachine.Namespace).Get(c.Context, keyPairName, metav1.GetOptions{})
		if err != nil {
			continue
		}
		publicKeys = append(publicKeys, keyPair.Spec.PublicKey)
	}
	appendPublicKeys := len(publicKeys) > 0
	for _, line := range strings.Split(source.UserData, "\n") {
		if strings.HasPrefix(line, "ssh_authorized_keys:") {
			appendPublicKeys = false
			break
		}
	}
	if appendPublicKeys {
		if source.UserData == "" {
			source.UserData = fmt.Sprintf("#cloud-config\nssh_authorized_keys:\n  - %s", strings.Join(publicKeys, "\n  - "))
		} else {
			source.UserData += fmt.Sprintf("\nssh_authorized_keys:\n  - %s", strings.Join(publicKeys, "\n  - "))
		}
	}
}

func (c *Constructor) deviceProcessors() []util.Processor {
	return []util.Processor{
		{
			Field: constants.FieldVirtualMachineInput,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				inputName := r[constants.FieldInputName].(string)
				inputType := kubevirtv1.InputType(r[constants.FieldInputType].(string))
				inputBus := kubevirtv1.InputBus(r[constants.FieldInputBus].(string))
				c.VMBuilder.Input(inputName, inputType, inputBus)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineTPM,
			Parser: func(i interface{}) error {
				c.VMBuilder.TPM()
				return nil
			},
		},
	}
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	vm, err := c.VMBuilder.VM()
	if err != nil {
		return nil, err
	}
	c.Version.Spec.VM = harvsterv1.VirtualMachineSourceSpec{
		ObjectMeta: vm.Spec.Template.ObjectMeta,
		Spec: kubevirtv1.VirtualMachineSpec{
			RunStrategy: vm.Spec.RunStrategy,
			Template:    vm.Spec.Template,
		},
	}
	// Store VolumeClaimTemplates on the version's own annotations (not nested VM ObjectMeta)
	// because the K8s API server strips annotations from nested metav1.ObjectMeta fields.
	// The importer reads this back to map PVC references to disk size/image.
	if vct, ok := vm.Annotations[harvesterutil.AnnotationVolumeClaimTemplates]; ok {
		c.Version.Annotations[harvesterutil.AnnotationVolumeClaimTemplates] = vct
	}
	return c.Version, nil
}

func Creator(c *client.Client, ctx context.Context, namespace, name string) util.Constructor {
	version := &harvsterv1.VirtualMachineTemplateVersion{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}
	if name != "" {
		version.Name = name
	} else {
		version.GenerateName = "template-"
	}

	vmBuilder := builder.NewVMBuilder(vmTemplateVersionCreator).
		Namespace(namespace).Name(name).
		EvictionStrategy(true).
		DefaultPodAntiAffinity()

	return &Constructor{
		Client:    c,
		Context:   ctx,
		Version:   version,
		VMBuilder: vmBuilder,
	}
}
