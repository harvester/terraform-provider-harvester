package virtualmachine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/harvester/pkg/builder"
	harvesterutil "github.com/harvester/harvester/pkg/util"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// vmAffinity returns the affinity of the VM being built, empty when it has
// none yet, so that each affinity block can complete it.
func vmAffinity(vmBuilder *builder.VMBuilder) *corev1.Affinity {
	if affinity := vmBuilder.VirtualMachine.Spec.Template.Spec.Affinity; affinity != nil {
		return affinity
	}
	return &corev1.Affinity{}
}

// blockMap returns the content of a block. Terraform passes nil for a block
// written without any attribute, read as an empty map here.
func blockMap(i interface{}) map[string]interface{} {
	r, _ := i.(map[string]interface{})
	return r
}

// requiredOrPreferred returns the required and preferred lists of an
// affinity block, and an error when the block sets neither.
func requiredOrPreferred(block string, r map[string]interface{}, requiredKey, preferredKey string) ([]interface{}, []interface{}, error) {
	required, _ := r[requiredKey].([]interface{})
	preferred, _ := r[preferredKey].([]interface{})
	if len(required) == 0 && len(preferred) == 0 {
		return nil, nil, fmt.Errorf("%s: set a %s or a %s block", block, requiredKey, preferredKey)
	}
	return required, preferred, nil
}

// parseNodeAffinity returns the node affinity of a node_affinity block.
func parseNodeAffinity(r map[string]interface{}) (*corev1.NodeAffinity, error) {
	required, preferred, err := requiredOrPreferred(constants.FieldVirtualMachineNodeAffinity, r, constants.FieldNodeAffinityRequired, constants.FieldNodeAffinityPreferred)
	if err != nil {
		return nil, err
	}
	nodeAffinity := &corev1.NodeAffinity{}
	if len(required) > 0 {
		data, _ := blockMap(required[0])[constants.FieldNodeSelectorTerm].([]interface{})
		terms, err := parseNodeSelectorTerms(data)
		if err != nil {
			return nil, err
		}
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{NodeSelectorTerms: terms}
	}
	for _, item := range preferred {
		p := blockMap(item)
		data, _ := p[constants.FieldPreferredPreference].([]interface{})
		terms, err := parseNodeSelectorTerms(data)
		if err != nil {
			return nil, err
		}
		for _, term := range terms {
			nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, corev1.PreferredSchedulingTerm{
				Weight:     int32(p[constants.FieldPreferredWeight].(int)), //nolint:gosec // weight is validated 1-100 by schema
				Preference: term,
			})
		}
	}
	return nodeAffinity, nil
}

// parsePodAffinityRules returns the required and preferred terms of a
// pod_affinity or pod_anti_affinity block.
func parsePodAffinityRules(block string, r map[string]interface{}) ([]corev1.PodAffinityTerm, []corev1.WeightedPodAffinityTerm, error) {
	required, preferred, err := requiredOrPreferred(block, r, constants.FieldPodAffinityRequired, constants.FieldPodAffinityPreferred)
	if err != nil {
		return nil, nil, err
	}
	requiredTerms, err := parsePodAffinityTerms(required)
	if err != nil {
		return nil, nil, err
	}
	preferredTerms, err := parseWeightedPodAffinityTerms(preferred)
	if err != nil {
		return nil, nil, err
	}
	return requiredTerms, preferredTerms, nil
}

// parseLabelSelectorRequirements parses a list of label selector requirements
func parseLabelSelectorRequirements(data []interface{}) []metav1.LabelSelectorRequirement {
	requirements := make([]metav1.LabelSelectorRequirement, 0, len(data))
	for _, item := range data {
		r := blockMap(item)
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

// parseSelector parses a label_selector or namespace_selector block.
func parseSelector(item interface{}) *metav1.LabelSelector {
	r := blockMap(item)
	selector := &metav1.LabelSelector{}
	if matchLabels, ok := r[constants.FieldMatchLabels].(map[string]interface{}); ok && len(matchLabels) > 0 {
		selector.MatchLabels = make(map[string]string, len(matchLabels))
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
		r := blockMap(item)
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

// parseNodeSelectorTerms parses node selector terms. A term needs at least one
// match expression: an empty term would match every node, and the Harvester
// VM webhook (makeAffinityFromVMTemplate) drops the terms left without match
// expressions, match_fields alone included.
func parseNodeSelectorTerms(data []interface{}) ([]corev1.NodeSelectorTerm, error) {
	terms := make([]corev1.NodeSelectorTerm, 0, len(data))
	for _, item := range data {
		r := blockMap(item)
		term := corev1.NodeSelectorTerm{}
		if matchExprs, ok := r[constants.FieldMatchExpressions].([]interface{}); ok && len(matchExprs) > 0 {
			term.MatchExpressions = parseNodeSelectorRequirements(matchExprs)
		}
		if matchFields, ok := r[constants.FieldMatchFields].([]interface{}); ok && len(matchFields) > 0 {
			term.MatchFields = parseNodeSelectorRequirements(matchFields)
		}
		if len(term.MatchExpressions) == 0 {
			return nil, fmt.Errorf("each %s and %s needs at least one %s block", constants.FieldNodeSelectorTerm, constants.FieldPreferredPreference, constants.FieldMatchExpressions)
		}
		terms = append(terms, term)
	}
	return terms, nil
}

// parsePodAffinityTerms parses pod affinity terms
func parsePodAffinityTerms(data []interface{}) ([]corev1.PodAffinityTerm, error) {
	var terms []corev1.PodAffinityTerm
	for _, item := range data {
		r := blockMap(item)
		term := corev1.PodAffinityTerm{
			TopologyKey: r[constants.FieldTopologyKey].(string),
		}
		if labelSelector, ok := r[constants.FieldLabelSelector].([]interface{}); ok && len(labelSelector) > 0 {
			// An empty label selector would match every pod.
			term.LabelSelector = parseSelector(labelSelector[0])
			if len(term.LabelSelector.MatchLabels) == 0 && len(term.LabelSelector.MatchExpressions) == 0 {
				return nil, fmt.Errorf("%s needs %s or %s", constants.FieldLabelSelector, constants.FieldMatchLabels, constants.FieldMatchExpressions)
			}
		}
		if namespaces, ok := r[constants.FieldNamespaces].([]interface{}); ok && len(namespaces) > 0 {
			for _, ns := range namespaces {
				term.Namespaces = append(term.Namespaces, ns.(string))
			}
		}
		if nsSelector, ok := r[constants.FieldNamespaceSelector].([]interface{}); ok && len(nsSelector) > 0 {
			// An empty namespace selector selects every namespace, like the
			// "all namespaces" option of the Harvester UI.
			term.NamespaceSelector = parseSelector(nsSelector[0])
		}
		terms = append(terms, term)
	}
	return terms, nil
}

// parseWeightedPodAffinityTerms parses weighted pod affinity terms
func parseWeightedPodAffinityTerms(data []interface{}) ([]corev1.WeightedPodAffinityTerm, error) {
	var terms []corev1.WeightedPodAffinityTerm
	for _, item := range data {
		r := blockMap(item)
		podAffinityTerm, _ := r[constants.FieldPodAffinityTerm].([]interface{})
		parsed, err := parsePodAffinityTerms(podAffinityTerm)
		if err != nil {
			return nil, err
		}
		for _, term := range parsed {
			terms = append(terms, corev1.WeightedPodAffinityTerm{
				Weight:          int32(r[constants.FieldPreferredWeight].(int)), //nolint:gosec // weight is validated 1-100 by schema
				PodAffinityTerm: term,
			})
		}
	}
	return terms, nil
}

// harvesterManagedNodeAffinity returns a node affinity with only the
// requirements that the Harvester webhook manages in the required terms of
// affinity, nil when there are none. Harvester refuses an update that changes
// the CPU manager requirements (checkCpuManagerAffinityTermChanged), then adds
// its requirements back to every term, so one term is enough to carry them.
func harvesterManagedNodeAffinity(affinity *corev1.Affinity) *corev1.NodeAffinity {
	if affinity == nil || affinity.NodeAffinity == nil || affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return nil
	}
	var managed []corev1.NodeSelectorRequirement
	for _, term := range affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, req := range term.MatchExpressions {
			if helper.IsHarvesterManagedNodeSelector(req) {
				managed = append(managed, req)
			}
		}
	}
	if len(managed) == 0 {
		return nil
	}
	return &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: managed}},
	}}
}

// nodeAffinityParser sets the node affinity of a node_affinity block, keeping
// the requirements that Harvester manages.
func nodeAffinityParser(vmBuilder *builder.VMBuilder) func(interface{}) error {
	return func(i interface{}) error {
		nodeAffinity, err := parseNodeAffinity(blockMap(i))
		if err != nil {
			return err
		}
		affinity := vmAffinity(vmBuilder)
		if managed := harvesterManagedNodeAffinity(affinity); managed != nil {
			if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
				nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = managed.RequiredDuringSchedulingIgnoredDuringExecution
			} else {
				terms := nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				terms[0].MatchExpressions = append(terms[0].MatchExpressions, managed.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions...)
			}
		}
		affinity.NodeAffinity = nodeAffinity
		vmBuilder.Affinity(affinity)
		return nil
	}
}

// podAffinityParser sets the pod affinity of a pod_affinity block.
func podAffinityParser(vmBuilder *builder.VMBuilder) func(interface{}) error {
	return func(i interface{}) error {
		required, preferred, err := parsePodAffinityRules(constants.FieldVirtualMachinePodAffinity, blockMap(i))
		if err != nil {
			return err
		}
		affinity := vmAffinity(vmBuilder)
		affinity.PodAffinity = &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  required,
			PreferredDuringSchedulingIgnoredDuringExecution: preferred,
		}
		vmBuilder.Affinity(affinity)
		return nil
	}
}

// podAntiAffinityParser adds the rules of a pod_anti_affinity block to the
// default anti-affinity set by Creator and Updater.
func podAntiAffinityParser(vmBuilder *builder.VMBuilder) func(interface{}) error {
	return func(i interface{}) error {
		required, preferred, err := parsePodAffinityRules(constants.FieldVirtualMachinePodAntiAffinity, blockMap(i))
		if err != nil {
			return err
		}
		if slices.ContainsFunc(preferred, helper.IsDefaultPodAntiAffinityTerm) {
			return fmt.Errorf("%s: remove the %s term on %s, the provider always sets it", constants.FieldVirtualMachinePodAntiAffinity, constants.FieldPodAffinityPreferred, builder.LabelKeyVirtualMachineCreator)
		}
		affinity := vmAffinity(vmBuilder)
		if affinity.PodAntiAffinity == nil {
			affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
		}
		affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, required...)
		affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, preferred...)
		vmBuilder.Affinity(affinity)
		return nil
	}
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
			Field: constants.FieldVirtualMachineRequests,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
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
					vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Resources.Requests = requests
				}
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
								SecureBoot: new(false),
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
				firmware.Bootloader.EFI.SecureBoot = new(true)
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Firmware = firmware

				features := vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features
				if features == nil {
					features = &kubevirtv1.Features{}
				}
				features.SMM = &kubevirtv1.FeatureState{
					Enabled: new(true),
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
				diskCacheMode := r[constants.FieldDiskCacheMode].(string)
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
				if diskCacheMode != "" {
					mode := kubevirtv1.DriverCache(diskCacheMode)
					vmBuilder.DiskCacheMode(diskName, mode)
					if vmBuilder.Error != nil {
						return vmBuilder.Error
					}
				}

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
					pvcOption.StorageClassName = new(storageClassName)

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
			Field: constants.FieldVirtualMachineHostDevice,
			Parser: func(i interface{}) error {
				v := i.(map[string]interface{})
				name := v[constants.FieldHostDeviceName].(string)
				deviceName := v[constants.FieldHostDeviceDeviceName].(string)
				vmBuilder.AddHostDevice(name, deviceName, "")
				return nil
			},
		},
		{
			Field:  constants.FieldVirtualMachineNodeAffinity,
			Parser: nodeAffinityParser(vmBuilder),
		},
		{
			Field:  constants.FieldVirtualMachinePodAffinity,
			Parser: podAffinityParser(vmBuilder),
		},
		{
			Field:  constants.FieldVirtualMachinePodAntiAffinity,
			Parser: podAntiAffinityParser(vmBuilder),
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
	// Rebuild the affinity like Creator does, then from the configuration,
	// keeping the node selector requirements that Harvester manages.
	vm.Spec.Template.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: harvesterManagedNodeAffinity(vm.Spec.Template.Spec.Affinity),
	}
	vm.Annotations[harvesterutil.AnnotationVolumeClaimTemplates] = "[]"
	vmBuilder := &builder.VMBuilder{
		VirtualMachine: vm,
	}
	return newVMConstructor(c, ctx, vmBuilder.DefaultPodAntiAffinity())
}
