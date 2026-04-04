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
				sysprepSecretName := r[constants.FieldDiskSysprepSecretName].(string)
				sysprepConfigMapName := r[constants.FieldDiskSysprepConfigMapName].(string)
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
				if sysprepSecretName != "" {
					vmBuilder.Volume(diskName, kubevirtv1.Volume{
						Name: diskName,
						VolumeSource: kubevirtv1.VolumeSource{
							Sysprep: &kubevirtv1.SysprepSource{
								Secret: &corev1.LocalObjectReference{Name: sysprepSecretName},
							},
						},
					})
				} else if sysprepConfigMapName != "" {
					vmBuilder.Volume(diskName, kubevirtv1.Volume{
						Name: diskName,
						VolumeSource: kubevirtv1.VolumeSource{
							Sysprep: &kubevirtv1.SysprepSource{
								ConfigMap: &corev1.LocalObjectReference{Name: sysprepConfigMapName},
							},
						},
					})
				} else if existingVolumeName != "" {
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
			Field: constants.FieldVirtualMachineHyperv,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				hv := parseHyperv(r)
				features := vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features
				if features == nil {
					features = &kubevirtv1.Features{}
				}
				features.Hyperv = hv
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features = features
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineHypervPassthrough,
			Parser: func(i interface{}) error {
				if i.(bool) {
					features := vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features
					if features == nil {
						features = &kubevirtv1.Features{}
					}
					features.HypervPassthrough = &kubevirtv1.HyperVPassthrough{
						Enabled: ptr.To(true),
					}
					vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Features = features
				}
				return nil
			},
		},
		{
			Field: constants.FieldVirtualMachineClock,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				clock, err := parseClock(r)
				if err != nil {
					return err
				}
				vmBuilder.VirtualMachine.Spec.Template.Spec.Domain.Clock = clock
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

func parseHyperv(r map[string]interface{}) *kubevirtv1.FeatureHyperv {
	hv := &kubevirtv1.FeatureHyperv{}
	setBool := func(field string) *kubevirtv1.FeatureState {
		if v, ok := r[field].(bool); ok && v {
			return &kubevirtv1.FeatureState{Enabled: ptr.To(true)}
		}
		return nil
	}
	hv.Relaxed = setBool(constants.FieldHypervRelaxed)
	hv.VAPIC = setBool(constants.FieldHypervVAPIC)
	hv.VPIndex = setBool(constants.FieldHypervVPIndex)
	hv.Runtime = setBool(constants.FieldHypervRuntime)
	hv.SyNIC = setBool(constants.FieldHypervSyNIC)
	hv.Reset = setBool(constants.FieldHypervReset)
	hv.Frequencies = setBool(constants.FieldHypervFrequencies)
	hv.Reenlightenment = setBool(constants.FieldHypervReenlightenment)
	hv.TLBFlush = setBool(constants.FieldHypervTLBFlush)
	hv.IPI = setBool(constants.FieldHypervIPI)
	hv.EVMCS = setBool(constants.FieldHypervEVMCS)

	if v, ok := r[constants.FieldHypervSpinlocks].(bool); ok && v {
		retries := uint32(r[constants.FieldHypervSpinlocksRetries].(int)) // nolint: gosec
		hv.Spinlocks = &kubevirtv1.FeatureSpinlocks{
			Enabled: ptr.To(true),
			Retries: &retries,
		}
	}

	if v, ok := r[constants.FieldHypervSyNICTimer].(bool); ok && v {
		hv.SyNICTimer = &kubevirtv1.SyNICTimer{
			Enabled: ptr.To(true),
		}
		if direct, ok := r[constants.FieldHypervSyNICTimerDirect].(bool); ok && direct {
			hv.SyNICTimer.Direct = &kubevirtv1.FeatureState{Enabled: ptr.To(true)}
		}
	}

	if v, ok := r[constants.FieldHypervVendorID].(bool); ok && v {
		hv.VendorID = &kubevirtv1.FeatureVendorID{
			Enabled:  ptr.To(true),
			VendorID: r[constants.FieldHypervVendorIDValue].(string),
		}
	}
	return hv
}

func parseClock(r map[string]interface{}) (*kubevirtv1.Clock, error) {
	clock := &kubevirtv1.Clock{}
	tz, hasTZ := r[constants.FieldClockTimezone].(string)
	offset, hasOffset := r[constants.FieldClockUTCOffsetSeconds].(int)
	if hasTZ && tz != "" && hasOffset && offset != 0 {
		return nil, errors.New("clock: timezone and utc_offset_seconds are mutually exclusive")
	}
	if hasTZ && tz != "" {
		timezone := kubevirtv1.ClockOffsetTimezone(tz)
		clock.Timezone = &timezone
	} else if hasOffset && offset != 0 {
		clock.UTC = &kubevirtv1.ClockOffsetUTC{OffsetSeconds: &offset}
	}

	if timerList, ok := r[constants.FieldClockTimer].([]interface{}); ok && len(timerList) > 0 {
		clock.Timer = parseClockTimers(timerList[0].(map[string]interface{}))
	}
	return clock, nil
}

// getTimerBlock extracts the first element of a TypeList timer sub-block.
func getTimerBlock(t map[string]interface{}, key string) (map[string]interface{}, bool) {
	list, ok := t[key].([]interface{})
	if !ok || len(list) == 0 {
		return nil, false
	}
	return list[0].(map[string]interface{}), true
}

func parseClockTimers(t map[string]interface{}) *kubevirtv1.Timer {
	timer := &kubevirtv1.Timer{}

	if h, ok := getTimerBlock(t, constants.FieldTimerHPET); ok {
		enabled := h[constants.FieldTimerEnabled].(bool)
		timer.HPET = &kubevirtv1.HPETTimer{Enabled: &enabled}
		if tp, _ := h[constants.FieldTimerTickPolicy].(string); tp != "" {
			timer.HPET.TickPolicy = kubevirtv1.HPETTickPolicy(tp)
		}
	}

	if k, ok := getTimerBlock(t, constants.FieldTimerKVM); ok {
		enabled := k[constants.FieldTimerEnabled].(bool)
		timer.KVM = &kubevirtv1.KVMTimer{Enabled: &enabled}
	}

	if p, ok := getTimerBlock(t, constants.FieldTimerPIT); ok {
		enabled := p[constants.FieldTimerEnabled].(bool)
		timer.PIT = &kubevirtv1.PITTimer{Enabled: &enabled}
		if tp, _ := p[constants.FieldTimerTickPolicy].(string); tp != "" {
			timer.PIT.TickPolicy = kubevirtv1.PITTickPolicy(tp)
		}
	}

	if r, ok := getTimerBlock(t, constants.FieldTimerRTC); ok {
		enabled := r[constants.FieldTimerEnabled].(bool)
		timer.RTC = &kubevirtv1.RTCTimer{Enabled: &enabled}
		if tp, _ := r[constants.FieldTimerTickPolicy].(string); tp != "" {
			timer.RTC.TickPolicy = kubevirtv1.RTCTickPolicy(tp)
		}
		if track, _ := r[constants.FieldTimerTrack].(string); track != "" {
			timer.RTC.Track = kubevirtv1.RTCTimerTrack(track)
		}
	}

	if hv, ok := getTimerBlock(t, constants.FieldTimerHyperv); ok {
		enabled := hv[constants.FieldTimerEnabled].(bool)
		timer.Hyperv = &kubevirtv1.HypervTimer{Enabled: &enabled}
	}

	return timer
}

func Updater(c *client.Client, ctx context.Context, vm *kubevirtv1.VirtualMachine) util.Constructor {
	vm.Spec.Template.Spec.Networks = []kubevirtv1.Network{}
	vm.Spec.Template.Spec.Domain.Devices.TPM = nil
	vm.Spec.Template.Spec.Domain.Devices.Interfaces = []kubevirtv1.Interface{}
	vm.Spec.Template.Spec.Domain.Devices.Disks = []kubevirtv1.Disk{}
	vm.Spec.Template.Spec.Domain.Devices.Inputs = []kubevirtv1.Input{}
	vm.Spec.Template.Spec.Volumes = []kubevirtv1.Volume{}
	if vm.Spec.Template.Spec.Domain.Features != nil {
		vm.Spec.Template.Spec.Domain.Features.Hyperv = nil
		vm.Spec.Template.Spec.Domain.Features.HypervPassthrough = nil
	}
	vm.Spec.Template.Spec.Domain.Clock = nil
	vm.Annotations[harvesterutil.AnnotationVolumeClaimTemplates] = "[]"
	return newVMConstructor(c, ctx, &builder.VMBuilder{
		VirtualMachine: vm,
	})
}
