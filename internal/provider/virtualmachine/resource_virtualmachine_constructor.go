package virtualmachine

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

	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// ensureAffinity initializes the Affinity struct if nil
func ensureAffinity(vmBuilder *builder.VMBuilder) {
	if vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity == nil {
		vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity = &corev1.Affinity{}
	}
}

// parseLabelSelectorRequirements parses a list of label selector requirements
func parseLabelSelectorRequirements(data []interface{}) []metav1.LabelSelectorRequirement {
	requirements := make([]metav1.LabelSelectorRequirement, 0, len(data))
	for _, item := range data {
		r := item.(map[string]interface{})
		req := metav1.LabelSelectorRequirement{
			Key:      r[constants.FieldExpressionKey].(string),
			Operator: metav1.LabelSelectorOperator(r[constants.FieldExpressionOperator].(string)),
		}
		if values, ok := r[constants.FieldExpressionValues].([]interface{}); ok {
			for _, v := range values {
				req.Values = append(req.Values, v.(string))
			}
		}
		requirements = append(requirements, req)
	}
	return requirements
}

// parseLabelSelector parses a label selector from Terraform state
func parseLabelSelector(data []interface{}) *metav1.LabelSelector {
	if len(data) == 0 {
		return nil
	}
	r := data[0].(map[string]interface{})
	selector := &metav1.LabelSelector{}

	if matchLabels, ok := r[constants.FieldMatchLabels].(map[string]interface{}); ok && len(matchLabels) > 0 {
		selector.MatchLabels = make(map[string]string)
		for k, v := range matchLabels {
			selector.MatchLabels[k] = v.(string)
		}
	}

	if matchExprs, ok := r[constants.FieldMatchExpressions].([]interface{}); ok && len(matchExprs) > 0 {
		selector.MatchExpressions = parseLabelSelectorRequirements(matchExprs)
	}

	return selector
}

// parseNodeSelectorRequirements parses node selector requirements
func parseNodeSelectorRequirements(data []interface{}) []corev1.NodeSelectorRequirement {
	requirements := make([]corev1.NodeSelectorRequirement, 0, len(data))
	for _, item := range data {
		r := item.(map[string]interface{})
		req := corev1.NodeSelectorRequirement{
			Key:      r[constants.FieldExpressionKey].(string),
			Operator: corev1.NodeSelectorOperator(r[constants.FieldExpressionOperator].(string)),
		}
		if values, ok := r[constants.FieldExpressionValues].([]interface{}); ok {
			for _, v := range values {
				req.Values = append(req.Values, v.(string))
			}
		}
		requirements = append(requirements, req)
	}
	return requirements
}

// parseNodeSelectorTerms parses node selector terms
func parseNodeSelectorTerms(data []interface{}) []corev1.NodeSelectorTerm {
	terms := make([]corev1.NodeSelectorTerm, 0, len(data))
	for _, item := range data {
		r := item.(map[string]interface{})
		term := corev1.NodeSelectorTerm{}
		if matchExprs, ok := r[constants.FieldMatchExpressions].([]interface{}); ok && len(matchExprs) > 0 {
			term.MatchExpressions = parseNodeSelectorRequirements(matchExprs)
		}
		if matchFields, ok := r[constants.FieldMatchFields].([]interface{}); ok && len(matchFields) > 0 {
			term.MatchFields = parseNodeSelectorRequirements(matchFields)
		}
		terms = append(terms, term)
	}
	return terms
}

// parsePodAffinityTerms parses pod affinity terms
func parsePodAffinityTerms(data []interface{}) []corev1.PodAffinityTerm {
	terms := make([]corev1.PodAffinityTerm, 0, len(data))
	for _, item := range data {
		r := item.(map[string]interface{})
		term := corev1.PodAffinityTerm{
			TopologyKey: r[constants.FieldTopologyKey].(string),
		}
		if labelSelector, ok := r[constants.FieldLabelSelector].([]interface{}); ok && len(labelSelector) > 0 {
			term.LabelSelector = parseLabelSelector(labelSelector)
		}
		if namespaces, ok := r[constants.FieldNamespaces].([]interface{}); ok && len(namespaces) > 0 {
			for _, ns := range namespaces {
				term.Namespaces = append(term.Namespaces, ns.(string))
			}
		}
		if nsSelector, ok := r[constants.FieldNamespaceSelector].([]interface{}); ok && len(nsSelector) > 0 {
			term.NamespaceSelector = parseLabelSelector(nsSelector)
		}
		terms = append(terms, term)
	}
	return terms
}

// parseWeightedPodAffinityTerms parses weighted pod affinity terms
func parseWeightedPodAffinityTerms(data []interface{}) []corev1.WeightedPodAffinityTerm {
	terms := make([]corev1.WeightedPodAffinityTerm, 0, len(data))
	for _, item := range data {
		r := item.(map[string]interface{})
		term := corev1.WeightedPodAffinityTerm{
			Weight: int32(r[constants.FieldPreferredWeight].(int)), //nolint:gosec // weight is validated 1-100 by schema
		}
		if podAffinityTerm, ok := r[constants.FieldPodAffinityTerm].([]interface{}); ok && len(podAffinityTerm) > 0 {
			parsed := parsePodAffinityTerms(podAffinityTerm)
			if len(parsed) > 0 {
				term.PodAffinityTerm = parsed[0]
			}
		}
		terms = append(terms, term)
	}
	return terms
}

const (
	vmCreator = "terraform-provider-harvester"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Client  *client.Client
	Context context.Context

	Builder *builder.VMBuilder
}

func (c *Constructor) Setup() util.Processors {
	vmBuilder := c.Builder
	if vmBuilder == nil {
		return nil
	}
	processors := util.NewProcessors().
		Tags(&c.Builder.VirtualMachine.Labels).
		Labels(&c.Builder.VirtualMachine.Labels).
		Description(&c.Builder.VirtualMachine.Annotations)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldVirtualMachineCPU,
			Parser: func(i interface{}) error {
				vmBuilder.CPU(i.(int))
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineCPUModel,
			Parser: func(i interface{}) error {
				cpuModel := i.(string)
				if cpuModel != "" {
					vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.CPU.Model = cpuModel
				}
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
			Field: constants.FieldVirtualMachineEFI,
			Parser: func(i interface{}) error {
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
				if oldFirmware := vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware; oldFirmware != nil {
					if firmware == nil {
						firmware = &kubevirtv1.Firmware{}
					}
					firmware.UUID = oldFirmware.UUID
					firmware.Serial = oldFirmware.Serial
				}
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware = firmware
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVirtualMachineSecureBoot,
			Parser: func(i interface{}) error {
				firmware := vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware
				if firmware == nil || firmware.Bootloader == nil || firmware.Bootloader.EFI == nil {
					return errors.New("EFI must be enabled to use Secure Boot. ")
				}
				firmware.Bootloader.EFI.SecureBoot = ptr.To(true)
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware = firmware

				features := vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features
				if features == nil {
					features = &kubevirtv1.Features{}
				}
				features.SMM = &kubevirtv1.FeatureState{
					Enabled: ptr.To(true),
				}
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features = features
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
			Field: constants.FieldVirtualMachineReservedMemory,
			Parser: func(i interface{}) error {
				reservedMemory := i.(string)
				if reservedMemory != "" {
					vmBuilder.Annotations(map[string]string{
						harvesterutil.AnnotationReservedMemory: reservedMemory,
					})
				} else {
					delete(vmBuilder.VirtualMachine.Annotations, harvesterutil.AnnotationReservedMemory)
				}
				return nil
			},
			Required: true,
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
				interfaceName := r[constants.FieldNetworkInterfaceName].(string)
				interfaceType := r[constants.FieldNetworkInterfaceType].(string)
				interfaceModel := r[constants.FieldNetworkInterfaceModel].(string)
				interfaceMACAddress := r[constants.FieldNetworkInterfaceMACAddress].(string)
				interfaceWaitForLease := r[constants.FieldNetworkInterfaceWaitForLease].(bool)
				networkName := r[constants.FieldNetworkInterfaceNetworkName].(string)
				bootOrder := r[constants.FieldNetworkInterfaceBootOrder].(int)

				if interfaceType == "" {
					if networkName == "" {
						interfaceType = builder.NetworkInterfaceTypeMasquerade
					} else {
						interfaceType = builder.NetworkInterfaceTypeBridge
					}
				}
				if interfaceWaitForLease {
					vmBuilder.WaitForLease(interfaceName)
				}
				vmBuilder.NetworkInterface(interfaceName, interfaceModel, interfaceMACAddress, interfaceType, networkName)
				if bootOrder != 0 {
					vmBuilder.SetNetworkInterfaceBootOrder(interfaceName, uint(bootOrder)) // nolint: gosec
				}
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

				vmBuilder.Disk(diskName, diskBus, isCDRom, uint(bootOrder)) // nolint: gosec
				if existingVolumeName != "" {
					vmBuilder.ExistingPVCVolume(diskName, existingVolumeName, hotPlug)
				} else if containerImageName != "" {
					vmBuilder.ContainerDiskVolume(diskName, containerImageName, builder.DefaultImagePullPolicy)
				} else if isCDRom && imageNamespacedName == "" {
					// Empty CDRom: don't prepare volume
				} else {
					pvcOption := &builder.PersistentVolumeClaimOption{
						VolumeMode: corev1.PersistentVolumeBlock,
						AccessMode: corev1.ReadWriteMany,
					}
					// storageClass
					storageClassName := r[constants.FieldVolumeStorageClassName].(string)
					if imageNamespacedName != "" {
						imageNamespace, imageName, err := helper.NamespacedNamePartsByDefault(imageNamespacedName, c.Builder.VirtualMachine.Namespace)
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
					} else {
						if storageClassName == "" {
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
				// only apply ssh username and ssh keys to cloud-init if UserDataBase64 and UserDataSecretName are not set
				if cloudInitSource.UserDataBase64 == "" && cloudInitSource.UserDataSecretName == "" {
					if vmBuilder.VirtualMachine.Labels != nil {
						if sshUsername, ok := vmBuilder.VirtualMachine.Labels[builder.LabelPrefixHarvesterTag+constants.LabelSSHUsername]; ok && sshUsername != "" {
							if cloudInitSource.UserData == "" {
								cloudInitSource.UserData = fmt.Sprintf("#cloud-config\nuser: %s\n", sshUsername)
							} else {
								appendUser := true
								for _, line := range strings.Split(cloudInitSource.UserData, "\n") {
									if strings.HasPrefix(line, "user: ") {
										appendUser = false
										break
									}
								}
								if appendUser {
									cloudInitSource.UserData += fmt.Sprintf("\nuser: %s\n", sshUsername)
								}
							}
						}
					}

					publicKeys := []string{}
					for _, sshName := range vmBuilder.SSHNames {
						_, keyPairName, err := helper.NamespacedNameParts(sshName)
						if err != nil {
							return err
						}
						keyPair, err := c.Client.HarvesterClient.HarvesterhciV1beta1().KeyPairs(c.Builder.VirtualMachine.Namespace).Get(c.Context, keyPairName, metav1.GetOptions{})
						if err != nil {
							return err
						}
						publicKeys = append(publicKeys, keyPair.Spec.PublicKey)
					}
					appendPublicKeys := len(publicKeys) > 0
					for _, line := range strings.Split(cloudInitSource.UserData, "\n") {
						if strings.HasPrefix(line, "ssh_authorized_keys:") {
							appendPublicKeys = false
							break
						}
					}
					if appendPublicKeys {
						if cloudInitSource.UserData == "" {
							cloudInitSource.UserData = fmt.Sprintf("#cloud-config\nssh_authorized_keys:\n  - %s", strings.Join(publicKeys, "\n  - "))
						} else {
							cloudInitSource.UserData += fmt.Sprintf("\nssh_authorized_keys:\n  - %s", strings.Join(publicKeys, "\n  - "))
						}
					}
				}
				diskName := builder.CloudInitDiskName
				vmBuilder.Disk(diskName, diskBus, isCDRom, 0)
				vmBuilder.CloudInit(diskName, cloudInitSource)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineInput,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				inputName := r[constants.FieldInputName].(string)
				inputType := kubevirtv1.InputType(r[constants.FieldInputType].(string))
				inputBus := kubevirtv1.InputBus(r[constants.FieldInputBus].(string))
				vmBuilder.Input(inputName, inputType, inputBus)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineTPM,
			Parser: func(i interface{}) error {
				vmBuilder.TPM()
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineCPUPinning,
			Parser: func(i interface{}) error {
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.CPU.DedicatedCPUPlacement = i.(bool)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineIsolateEmulatorThread,
			Parser: func(i interface{}) error {
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.CPU.IsolateEmulatorThread = i.(bool)
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineNodeSelector,
			Parser: func(i interface{}) error {
				v := i.(map[string]interface{})
				vmBuilder.VirtualMachine.Spec.Template.Spec.NodeSelector = make(map[string]string)
				for k, val := range v {
					vmBuilder.VirtualMachine.Spec.Template.Spec.NodeSelector[k] = val.(string)
				}
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineNodeAffinity,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				ensureAffinity(vmBuilder)
				nodeAffinity := &corev1.NodeAffinity{}

				if required, ok := r[constants.FieldNodeAffinityRequired].([]interface{}); ok && len(required) > 0 {
					reqData := required[0].(map[string]interface{})
					if terms, ok := reqData[constants.FieldNodeSelectorTerm].([]interface{}); ok && len(terms) > 0 {
						nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
							NodeSelectorTerms: parseNodeSelectorTerms(terms),
						}
					}
				}

				if preferred, ok := r[constants.FieldNodeAffinityPreferred].([]interface{}); ok && len(preferred) > 0 {
					for _, item := range preferred {
						p := item.(map[string]interface{})
						prefTerm := corev1.PreferredSchedulingTerm{
							Weight: int32(p[constants.FieldPreferredWeight].(int)), //nolint:gosec // weight is validated 1-100 by schema
						}
						if pref, ok := p[constants.FieldPreferredPreference].([]interface{}); ok && len(pref) > 0 {
							terms := parseNodeSelectorTerms(pref)
							if len(terms) > 0 {
								prefTerm.Preference = terms[0]
							}
						}
						nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
							nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, prefTerm)
					}
				}

				vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity.NodeAffinity = nodeAffinity
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachinePodAffinity,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				ensureAffinity(vmBuilder)
				podAffinity := &corev1.PodAffinity{}

				if required, ok := r[constants.FieldPodAffinityRequired].([]interface{}); ok && len(required) > 0 {
					podAffinity.RequiredDuringSchedulingIgnoredDuringExecution = parsePodAffinityTerms(required)
				}

				if preferred, ok := r[constants.FieldPodAffinityPreferred].([]interface{}); ok && len(preferred) > 0 {
					podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = parseWeightedPodAffinityTerms(preferred)
				}

				vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity.PodAffinity = podAffinity
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachinePodAntiAffinity,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				ensureAffinity(vmBuilder)
				podAntiAffinity := &corev1.PodAntiAffinity{}

				if required, ok := r[constants.FieldPodAffinityRequired].([]interface{}); ok && len(required) > 0 {
					podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = parsePodAffinityTerms(required)
				}

				if preferred, ok := r[constants.FieldPodAffinityPreferred].([]interface{}); ok && len(preferred) > 0 {
					podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = parseWeightedPodAffinityTerms(preferred)
				}

				vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity.PodAntiAffinity = podAntiAffinity
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

func newVMConstructor(c *client.Client, ctx context.Context, vmBuilder *builder.VMBuilder) util.Constructor {
	return &Constructor{
		Client:  c,
		Context: ctx,
		Builder: vmBuilder,
	}
}

func Creator(c *client.Client, ctx context.Context, namespace, name string) util.Constructor {
	vmBuilder := builder.NewVMBuilder(vmCreator).
		Namespace(namespace).Name(name).
		EvictionStrategy(true).
		DefaultPodAntiAffinity()
	return newVMConstructor(c, ctx, vmBuilder)
}

func Updater(c *client.Client, ctx context.Context, vm *kubevirtv1.VirtualMachine) util.Constructor {
	vm.Spec.Template.Spec.Networks = []kubevirtv1.Network{}
	vm.Spec.Template.Spec.Domain.Devices.TPM = nil
	vm.Spec.Template.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{}
	vm.Spec.Template.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{}
	vm.Spec.Template.Spec.Domain.Devices.Inputs = []kubevirtv1.Input{}
	vm.Spec.Template.Spec.Volumes = []kubevirtv1.Volume{}
	vm.Spec.Template.Spec.Affinity = nil // Clear affinity to allow complete replacement
	vm.Annotations[harvesterutil.AnnotationVolumeClaimTemplates] = "[]"
	return newVMConstructor(c, ctx, &builder.VMBuilder{
		VirtualMachine: vm,
	})
}
