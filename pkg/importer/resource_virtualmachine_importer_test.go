package importer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/harvester/pkg/builder"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestNetworkInterface(t *testing.T) {
	type testcase struct {
		importer    *VMImporter
		expectation []map[string]interface{}
		expectError error
	}

	properties := []string{
		constants.FieldNetworkInterfaceName,
		constants.FieldNetworkInterfaceType,
		constants.FieldNetworkInterfaceModel,
		constants.FieldNetworkInterfaceMACAddress,
		constants.FieldNetworkInterfaceNetworkName,
		constants.FieldNetworkInterfaceBootOrder,
		constants.FieldNetworkInterfaceIPAddress,
		constants.FieldNetworkInterfaceInterfaceName,
		constants.FieldNetworkInterfaceWaitForLease,
	}

	testcases := []testcase{
		{
			// a VM that doesn't have any network interface
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{},
							},
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									Devices: kubevirtv1.Devices{
										Interfaces: []kubevirtv1.Interface{},
									},
								},
							},
						},
					},
				},
				VirtualMachineInstance: &kubevirtv1.VirtualMachineInstance{},
			},
			expectation: []map[string]interface{}{},
			expectError: nil,
		},
		{
			// a VM that has a single minimal bridge network interface, but no IP
			// address
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{},
							},
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									Devices: kubevirtv1.Devices{
										Interfaces: []kubevirtv1.Interface{
											{
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{1}[0],
											},
										},
									},
								},
							},
						},
					},
				},
				VirtualMachineInstance: &kubevirtv1.VirtualMachineInstance{},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:         "",
					constants.FieldNetworkInterfaceType:         builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:        "",
					constants.FieldNetworkInterfaceMACAddress:   "",
					constants.FieldNetworkInterfaceNetworkName:  "",
					constants.FieldNetworkInterfaceBootOrder:    &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease: false,
				},
			},
			expectError: nil,
		},
		{
			// a VM that has a single minimal bridge network interface, and only
			// a link-local IP addresses
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{},
							},
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									Devices: kubevirtv1.Devices{
										Interfaces: []kubevirtv1.Interface{
											{
												Name: "net0",
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{1}[0],
											},
										},
									},
								},
							},
						},
					},
				},
				VirtualMachineInstance: &kubevirtv1.VirtualMachineInstance{
					Status: kubevirtv1.VirtualMachineInstanceStatus{
						Interfaces: []kubevirtv1.VirtualMachineInstanceNetworkInterface{
							{
								Name:          "net0",
								InterfaceName: "eth0",
								IPs:           []string{"169.254.10.140/24", "fe80::21f:bcff:fe13:405/64"},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:         "net0",
					constants.FieldNetworkInterfaceType:         builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:        "",
					constants.FieldNetworkInterfaceMACAddress:   "",
					constants.FieldNetworkInterfaceNetworkName:  "",
					constants.FieldNetworkInterfaceBootOrder:    &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease: false,
				},
			},
			expectError: nil,
		},
		{
			// a VM that has a single minimal bridge network interface with IP
			// addresses
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{},
							},
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									Devices: kubevirtv1.Devices{
										Interfaces: []kubevirtv1.Interface{
											{
												Name: "net0",
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{1}[0],
											},
										},
									},
								},
							},
						},
					},
				},
				VirtualMachineInstance: &kubevirtv1.VirtualMachineInstance{
					Status: kubevirtv1.VirtualMachineInstanceStatus{
						Interfaces: []kubevirtv1.VirtualMachineInstanceNetworkInterface{
							{
								Name:          "net0",
								InterfaceName: "eth0",
								IPs:           []string{"192.168.178.64/24", "fe80::21f:bcff:fe13:405/64"},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:          "net0",
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     "192.168.178.64/24",
					constants.FieldNetworkInterfaceInterfaceName: "eth0",
				},
			},
			expectError: nil,
		},
		{
			// a VM that has multiple minimal bridge network interfaces with several IP
			// addresses
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{},
							},
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									Devices: kubevirtv1.Devices{
										Interfaces: []kubevirtv1.Interface{
											{
												Name: "net0",
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{1}[0],
											},
											{
												Name: "net1",
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{2}[0],
											},
											{
												Name: "net2",
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{3}[0],
											},
										},
									},
								},
							},
						},
					},
				},
				VirtualMachineInstance: &kubevirtv1.VirtualMachineInstance{
					Status: kubevirtv1.VirtualMachineInstanceStatus{
						Interfaces: []kubevirtv1.VirtualMachineInstanceNetworkInterface{
							{
								Name:          "net0",
								InterfaceName: "eth0",
								IPs:           []string{"192.168.178.64/24", "fe80::21f:bcff:fe13:405/64"},
							},
							{
								Name:          "net1",
								InterfaceName: "eth1",
								IPs:           []string{"fe80::21f:bcff:fe13:406/64"},
							},
							{
								Name:          "net2",
								InterfaceName: "eth2",
								IPs:           []string{"192.168.180.64/24", "169.254.180.64/24", "201.168.180.64/24"},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:          "net0",
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     "192.168.178.64/24",
					constants.FieldNetworkInterfaceInterfaceName: "eth0",
				},
				{
					constants.FieldNetworkInterfaceName:         "net1",
					constants.FieldNetworkInterfaceType:         builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:        "",
					constants.FieldNetworkInterfaceMACAddress:   "",
					constants.FieldNetworkInterfaceNetworkName:  "",
					constants.FieldNetworkInterfaceBootOrder:    &[]uint{2}[0],
					constants.FieldNetworkInterfaceWaitForLease: false,
				},
				{
					constants.FieldNetworkInterfaceName:          "net2",
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{3}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     "192.168.180.64/24",
					constants.FieldNetworkInterfaceInterfaceName: "eth2",
				},
			},
			expectError: nil,
		},
		{
			// a VM that has a minimal bridge network interface with multiple IP
			// addresses in different order
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{},
							},
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									Devices: kubevirtv1.Devices{
										Interfaces: []kubevirtv1.Interface{
											{
												Name: "net0",
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{1}[0],
											},
										},
									},
								},
							},
						},
					},
				},
				VirtualMachineInstance: &kubevirtv1.VirtualMachineInstance{
					Status: kubevirtv1.VirtualMachineInstanceStatus{
						Interfaces: []kubevirtv1.VirtualMachineInstanceNetworkInterface{
							{
								Name:          "net0",
								InterfaceName: "eth0",
								IPs:           []string{"201.168.180.64/24", "169.254.180.64/24", "192.168.180.64/24"},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:          "net0",
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     "192.168.180.64/24",
					constants.FieldNetworkInterfaceInterfaceName: "eth0",
				},
			},
			expectError: nil,
		},
	}

	for _, tc := range testcases {
		outcome, err := tc.importer.NetworkInterface()

		if err != nil && tc.expectError == nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if err == nil && tc.expectError != nil {
			t.Errorf("Expected error %v, got nil", tc.expectError)
		}

		if len(outcome) != len(tc.expectation) {
			t.Errorf("Unexpected outcome length: %v, expected %v", len(outcome), len(tc.expectation))
		}

		for idx, out := range outcome {
			expect := tc.expectation[idx]

			for _, property := range properties {
				switch expect[property].(type) {
				case *uint:
					o := (out[property].(*uint))
					e := (expect[property].(*uint))
					if *o != *e {
						t.Errorf("Failed Importing NetworkInterface. Value for %v is %v, expeceted %v",
							property,
							*o,
							*e)
					}
				default:
					if out[property] != expect[property] {
						t.Errorf("Failed Importing NetworkInterface. Value for %v is %v, expeceted %v",
							property,
							out[property],
							expect[property])
					}
				}
			}
		}
	}
}

func TestCPU(t *testing.T) {
	type testcase struct {
		importer      *VMImporter
		expectedCores int
		expectedModel string
	}

	testcases := []testcase{
		{
			// VM with basic CPU configuration (no model specified)
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									CPU: &kubevirtv1.CPU{
										Cores: 2,
									},
								},
							},
						},
					},
				},
			},
			expectedCores: 2,
			expectedModel: "",
		},
		{
			// VM with CPU model set to specific Intel model
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									CPU: &kubevirtv1.CPU{
										Cores: 8,
										Model: "Skylake-Client-IBRS",
									},
								},
							},
						},
					},
				},
			},
			expectedCores: 8,
			expectedModel: "Skylake-Client-IBRS",
		},
	}

	for idx, tc := range testcases {
		cores := tc.importer.CPU()
		if cores != tc.expectedCores {
			t.Errorf("Test case %d: CPU() returned %d, expected %d", idx, cores, tc.expectedCores)
		}

		model := tc.importer.CPUModel()
		if model != tc.expectedModel {
			t.Errorf("Test case %d: CPUModel() returned %q, expected %q", idx, model, tc.expectedModel)
		}
	}
}

func TestResourceRequestsImport(t *testing.T) {
	// Test with explicit requests
	vm := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Resources: kubevirtv1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}

	reqs := importer.Requests()
	if len(reqs) != 1 {
		t.Fatalf("Requests() returned %d entries, want 1", len(reqs))
	}
	if got := reqs[0][constants.FieldRequestsCPU]; got != "500m" {
		t.Errorf("Requests() cpu = %q, want %q", got, "500m")
	}
	if got := reqs[0][constants.FieldRequestsMemory]; got != "512Mi" {
		t.Errorf("Requests() memory = %q, want %q", got, "512Mi")
	}

	// Test without requests (empty)
	vmNoReq := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Resources: kubevirtv1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
						},
					},
				},
			},
		},
	}
	importerNoReq := &VMImporter{VirtualMachine: vmNoReq}

	reqsNoReq := importerNoReq.Requests()
	if len(reqsNoReq) != 1 {
		t.Fatalf("Requests() no requests returned %d entries, want 1", len(reqsNoReq))
	}
	if got := reqsNoReq[0][constants.FieldRequestsCPU]; got != "" {
		t.Errorf("Requests() no requests cpu = %q, want empty", got)
	}
	if got := reqsNoReq[0][constants.FieldRequestsMemory]; got != "" {
		t.Errorf("Requests() no requests memory = %q, want empty", got)
	}

	// Test with nil Requests map
	vmNilReq := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Resources: kubevirtv1.ResourceRequirements{
							Requests: nil,
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
						},
					},
				},
			},
		},
	}
	importerNilReq := &VMImporter{VirtualMachine: vmNilReq}

	reqsNil := importerNilReq.Requests()
	if len(reqsNil) != 1 {
		t.Fatalf("Requests() nil returned %d entries, want 1", len(reqsNil))
	}
	if got := reqsNil[0][constants.FieldRequestsCPU]; got != "" {
		t.Errorf("Requests() nil cpu = %q, want empty", got)
	}
	if got := reqsNil[0][constants.FieldRequestsMemory]; got != "" {
		t.Errorf("Requests() nil memory = %q, want empty", got)
	}
}

func TestHypervImport(t *testing.T) {
	retries := uint32(8192)
	vm := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Features: &kubevirtv1.Features{
							Hyperv: &kubevirtv1.FeatureHyperv{
								Relaxed:         &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								VAPIC:           &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								VPIndex:         &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								Runtime:         &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								SyNIC:           &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								Frequencies:     &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								Reenlightenment: &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								Spinlocks: &kubevirtv1.FeatureSpinlocks{
									Enabled: ptr.To(true),
									Retries: &retries,
								},
								SyNICTimer: &kubevirtv1.SyNICTimer{
									Enabled: ptr.To(true),
									Direct:  &kubevirtv1.FeatureState{Enabled: ptr.To(true)},
								},
								VendorID: &kubevirtv1.FeatureVendorID{
									Enabled:  ptr.To(true),
									VendorID: "KVMKVMKVM",
								},
							},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}
	hvList := importer.Hyperv()

	if len(hvList) != 1 {
		t.Fatalf("Hyperv() returned %d entries, want 1", len(hvList))
	}
	hv := hvList[0]
	if hv[constants.FieldHypervRelaxed] != true {
		t.Error("relaxed should be true")
	}
	if hv[constants.FieldHypervVAPIC] != true {
		t.Error("vapic should be true")
	}
	if hv[constants.FieldHypervSpinlocks] != true {
		t.Error("spinlocks should be true")
	}
	if hv[constants.FieldHypervSpinlocksRetries] != 8192 {
		t.Errorf("spinlocks_retries = %v, want 8192", hv[constants.FieldHypervSpinlocksRetries])
	}
	if hv[constants.FieldHypervSyNICTimer] != true {
		t.Error("synictimer should be true")
	}
	if hv[constants.FieldHypervSyNICTimerDirect] != true {
		t.Error("synictimer_direct should be true")
	}
	if hv[constants.FieldHypervVendorID] != true {
		t.Error("vendorid should be true")
	}
	if hv[constants.FieldHypervVendorIDValue] != "KVMKVMKVM" {
		t.Errorf("vendorid_value = %q, want KVMKVMKVM", hv[constants.FieldHypervVendorIDValue])
	}

	// Test nil hyperv
	vmNil := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{},
				},
			},
		},
	}
	importerNil := &VMImporter{VirtualMachine: vmNil}
	if got := importerNil.Hyperv(); got != nil {
		t.Errorf("Hyperv() on nil features = %v, want nil", got)
	}
}

func TestHypervPassthroughImport(t *testing.T) {
	vm := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Features: &kubevirtv1.Features{
							HypervPassthrough: &kubevirtv1.HyperVPassthrough{
								Enabled: ptr.To(true),
							},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}
	if !importer.HypervPassthrough() {
		t.Error("HypervPassthrough() should be true")
	}

	vmNil := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{},
				},
			},
		},
	}
	importerNil := &VMImporter{VirtualMachine: vmNil}
	if importerNil.HypervPassthrough() {
		t.Error("HypervPassthrough() on nil should be false")
	}
}

func TestClockImport(t *testing.T) {
	tz := kubevirtv1.ClockOffsetTimezone("Europe/Paris")
	vm := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Clock: &kubevirtv1.Clock{
							ClockOffset: kubevirtv1.ClockOffset{
								Timezone: &tz,
							},
							Timer: &kubevirtv1.Timer{
								HPET: &kubevirtv1.HPETTimer{
									Enabled:    ptr.To(true),
									TickPolicy: kubevirtv1.HPETTickPolicyDelay,
								},
								KVM: &kubevirtv1.KVMTimer{
									Enabled: ptr.To(true),
								},
								PIT: &kubevirtv1.PITTimer{
									Enabled:    ptr.To(true),
									TickPolicy: kubevirtv1.PITTickPolicyCatchup,
								},
								RTC: &kubevirtv1.RTCTimer{
									Enabled:    ptr.To(true),
									TickPolicy: kubevirtv1.RTCTickPolicyCatchup,
									Track:      kubevirtv1.TrackGuest,
								},
								Hyperv: &kubevirtv1.HypervTimer{
									Enabled: ptr.To(true),
								},
							},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}
	clockList := importer.Clock()

	if len(clockList) != 1 {
		t.Fatalf("Clock() returned %d entries, want 1", len(clockList))
	}
	c := clockList[0]
	if c[constants.FieldClockTimezone] != "Europe/Paris" {
		t.Errorf("timezone = %q, want Europe/Paris", c[constants.FieldClockTimezone])
	}

	timerList := c[constants.FieldClockTimer].([]interface{})
	if len(timerList) != 1 {
		t.Fatalf("timer list has %d entries, want 1", len(timerList))
	}
	timer := timerList[0].(map[string]interface{})

	// Check HPET
	hpetList := timer[constants.FieldTimerHPET].([]interface{})
	if len(hpetList) != 1 {
		t.Fatalf("hpet has %d entries, want 1", len(hpetList))
	}
	hpet := hpetList[0].(map[string]interface{})
	if hpet[constants.FieldTimerEnabled] != true {
		t.Error("hpet enabled should be true")
	}
	if hpet[constants.FieldTimerTickPolicy] != "delay" {
		t.Errorf("hpet tick_policy = %q, want delay", hpet[constants.FieldTimerTickPolicy])
	}

	// Check RTC
	rtcList := timer[constants.FieldTimerRTC].([]interface{})
	if len(rtcList) != 1 {
		t.Fatalf("rtc has %d entries, want 1", len(rtcList))
	}
	rtc := rtcList[0].(map[string]interface{})
	if rtc[constants.FieldTimerTrack] != "guest" {
		t.Errorf("rtc track = %q, want guest", rtc[constants.FieldTimerTrack])
	}

	// Test nil clock
	vmNil := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{},
				},
			},
		},
	}
	importerNil := &VMImporter{VirtualMachine: vmNil}
	if got := importerNil.Clock(); got != nil {
		t.Errorf("Clock() on nil = %v, want nil", got)
	}
}

func TestSysprepDiskImport(t *testing.T) {
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{
								{
									Name: "sysprep-secret",
									DiskDevice: kubevirtv1.DiskDevice{
										CDRom: &kubevirtv1.CDRomTarget{Bus: "sata"},
									},
								},
								{
									Name: "sysprep-cm",
									DiskDevice: kubevirtv1.DiskDevice{
										CDRom: &kubevirtv1.CDRomTarget{Bus: "sata"},
									},
								},
							},
						},
					},
					Volumes: []kubevirtv1.Volume{
						{
							Name: "sysprep-secret",
							VolumeSource: kubevirtv1.VolumeSource{
								Sysprep: &kubevirtv1.SysprepSource{
									Secret: &corev1.LocalObjectReference{Name: "win-unattend"},
								},
							},
						},
						{
							Name: "sysprep-cm",
							VolumeSource: kubevirtv1.VolumeSource{
								Sysprep: &kubevirtv1.SysprepSource{
									ConfigMap: &corev1.LocalObjectReference{Name: "win-unattend-cm"},
								},
							},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}
	disks, _, err := importer.Volume()
	if err != nil {
		t.Fatalf("Volume() error: %v", err)
	}
	if len(disks) != 2 {
		t.Fatalf("Volume() returned %d disks, want 2", len(disks))
	}
	if disks[0][constants.FieldDiskSysprepSecretName] != "win-unattend" {
		t.Errorf("disk 0 sysprep_secret_name = %v, want win-unattend", disks[0][constants.FieldDiskSysprepSecretName])
	}
	if disks[1][constants.FieldDiskSysprepConfigMapName] != "win-unattend-cm" {
		t.Errorf("disk 1 sysprep_configmap_name = %v, want win-unattend-cm", disks[1][constants.FieldDiskSysprepConfigMapName])
	}
}
