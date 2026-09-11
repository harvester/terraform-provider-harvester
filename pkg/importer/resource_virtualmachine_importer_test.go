package importer

import (
	"reflect"
	"testing"

	assert "github.com/stretchr/testify/assert"

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

	const (
		networkName0   = "net0"
		networkName1   = "net1"
		networkName2   = "net2"
		interfaceName0 = "eth0"
		linkLocalIPv60 = "fe80::21f:bcff:fe13:405/64"
		ipv4Address0   = "192.168.178.64/24"
		ipv4Address1   = "192.168.180.64/24"
	)

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
												Name: networkName0,
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
								Name:          networkName0,
								InterfaceName: interfaceName0,
								IPs:           []string{"169.254.10.140/24", linkLocalIPv60},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:         networkName0,
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
												Name: networkName0,
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
								Name:          networkName0,
								InterfaceName: interfaceName0,
								IPs:           []string{ipv4Address0, linkLocalIPv60},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:          networkName0,
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     ipv4Address0,
					constants.FieldNetworkInterfaceInterfaceName: interfaceName0,
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
												Name: networkName0,
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{1}[0],
											},
											{
												Name: networkName1,
												InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
													Bridge: &kubevirtv1.InterfaceBridge{},
												},
												BootOrder: &[]uint{2}[0],
											},
											{
												Name: networkName2,
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
								Name:          networkName0,
								InterfaceName: interfaceName0,
								IPs:           []string{ipv4Address0, linkLocalIPv60},
							},
							{
								Name:          networkName1,
								InterfaceName: "eth1",
								IPs:           []string{"fe80::21f:bcff:fe13:406/64"},
							},
							{
								Name:          networkName2,
								InterfaceName: "eth2",
								IPs:           []string{ipv4Address1, "169.254.180.64/24", "201.168.180.64/24"},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:          networkName0,
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     ipv4Address0,
					constants.FieldNetworkInterfaceInterfaceName: interfaceName0,
				},
				{
					constants.FieldNetworkInterfaceName:         networkName1,
					constants.FieldNetworkInterfaceType:         builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:        "",
					constants.FieldNetworkInterfaceMACAddress:   "",
					constants.FieldNetworkInterfaceNetworkName:  "",
					constants.FieldNetworkInterfaceBootOrder:    &[]uint{2}[0],
					constants.FieldNetworkInterfaceWaitForLease: false,
				},
				{
					constants.FieldNetworkInterfaceName:          networkName2,
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{3}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     ipv4Address1,
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
												Name: networkName0,
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
								Name:          networkName0,
								InterfaceName: interfaceName0,
								IPs:           []string{"201.168.180.64/24", "169.254.180.64/24", ipv4Address1},
							},
						},
					},
				},
			},
			expectation: []map[string]interface{}{
				{
					constants.FieldNetworkInterfaceName:          networkName0,
					constants.FieldNetworkInterfaceType:          builder.NetworkInterfaceTypeBridge,
					constants.FieldNetworkInterfaceModel:         "",
					constants.FieldNetworkInterfaceMACAddress:    "",
					constants.FieldNetworkInterfaceNetworkName:   "",
					constants.FieldNetworkInterfaceBootOrder:     &[]uint{1}[0],
					constants.FieldNetworkInterfaceWaitForLease:  false,
					constants.FieldNetworkInterfaceIPAddress:     ipv4Address1,
					constants.FieldNetworkInterfaceInterfaceName: interfaceName0,
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

func TestStateRunningNetworkInterfaceNoIP(t *testing.T) {
	// Ensure that of `wait_for_lease` is set on a network interface, the state isn't reported as
	// `ready` as long as there is no IP address
	oldInstanceUID := "oldUid"
	networkInterfaces := []map[string]any{
		{
			constants.FieldNetworkInterfaceWaitForLease: true,
		},
	}
	vmi := kubevirtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			UID: "newUid",
		},
		Status: kubevirtv1.VirtualMachineInstanceStatus{
			Phase: "Running",
		},
	}

	vmImporter := NewVMImporter(nil, &vmi)
	state := vmImporter.State(networkInterfaces, oldInstanceUID)
	assert.Equal(t, state, constants.StateVirtualMachineRunning, "IP address not set results in state running")

	networkInterfaces[0][constants.FieldNetworkInterfaceIPAddress] = ""
	state = vmImporter.State(networkInterfaces, oldInstanceUID)
	assert.Equal(t, state, constants.StateVirtualMachineRunning, "IP address set to empty string results in state running")

	networkInterfaces[0][constants.FieldNetworkInterfaceIPAddress] = "123.123.123.123"
	state = vmImporter.State(networkInterfaces, oldInstanceUID)
	assert.Equal(t, state, constants.StateCommonReady, "IP address set to non-empty string results in state ready")
}

func TestVirtualMachineFeatures(t *testing.T) {
	make_importer := func(features *kubevirtv1.Features) *VMImporter {
		importer := VMImporter{
			VirtualMachine: &kubevirtv1.VirtualMachine{
				Spec: kubevirtv1.VirtualMachineSpec{
					Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
						Spec: kubevirtv1.VirtualMachineInstanceSpec{
							Domain: kubevirtv1.DomainSpec{
								Features: features,
							},
						},
					},
				},
			},
		}
		return &importer
	}

	type testcase struct {
		importer *VMImporter
		expected map[string]any
	}

	testcases := []testcase{
		{
			importer: make_importer(&kubevirtv1.Features{}),
			expected: map[string]any{},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					ACPI: kubevirtv1.FeatureState{
						Enabled: ptr.To(true),
					}}),
			expected: map[string]any{
				constants.FieldFeatureACPI: true,
			},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					APIC: &kubevirtv1.FeatureAPIC{
						Enabled:        ptr.To(true),
						EndOfInterrupt: true,
					}}),
			expected: map[string]any{
				constants.FieldFeatureAPIC: map[string]any{
					constants.FieldFeatureAPICEnabled:        true,
					constants.FieldFeatureAPICEndOfInterrupt: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						EVMCS: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVEVMCS: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						Frequencies: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVFrequencies: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						IPI: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVIPI: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						Reenlightenment: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVReenlightenment: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						Relaxed: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVRelaxed: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						Runtime: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVRuntime: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						Spinlocks: &kubevirtv1.FeatureSpinlocks{
							Enabled: ptr.To(true),
							Retries: func() *uint32 { v := uint32(8192); return &v }(),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVSpinlocks: map[string]any{
						constants.FieldFeatureHyperVSpinlocksEnabled: true,
						constants.FieldFeatureHyperVSpinlocksRetries: uint32(8192),
					}}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						SyNIC: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVSyNIC: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						SyNICTimer: &kubevirtv1.SyNICTimer{
							Enabled: ptr.To(true),
							Direct: &kubevirtv1.FeatureState{
								Enabled: ptr.To(true),
							}}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVSyNICTimer: map[string]any{
						constants.FieldFeatureHyperVSyNICTimerEnabled: true,
						constants.FieldFeatureHyperVSyNICTimerDirect:  true,
					}}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						TLBFlush: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVTLBFlush: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						VAPIC: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVVAPIC: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Hyperv: &kubevirtv1.FeatureHyperv{
						VPIndex: &kubevirtv1.FeatureState{
							Enabled: ptr.To(true),
						}}}),
			expected: map[string]any{
				constants.FieldFeatureHyperV: map[string]any{
					constants.FieldFeatureHyperVVPIndex: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					HypervPassthrough: &kubevirtv1.HyperVPassthrough{
						Enabled: ptr.To(true),
					}}),
			expected: map[string]any{
				constants.FieldFeatureHyperVPassthrough: true,
			},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					KVM: &kubevirtv1.FeatureKVM{
						Hidden: true,
					}}),
			expected: map[string]any{
				constants.FieldFeatureKVM: map[string]any{
					constants.FieldFeatureKVMHidden: true,
				}},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					Pvspinlock: &kubevirtv1.FeatureState{
						Enabled: ptr.To(true),
					}}),
			expected: map[string]any{
				constants.FieldFeaturePVSpinLock: true,
			},
		},
		{
			importer: make_importer(
				&kubevirtv1.Features{
					SMM: &kubevirtv1.FeatureState{
						Enabled: ptr.To(true),
					}}),
			expected: map[string]any{
				constants.FieldFeatureSMM: true,
			},
		},
	}

	for _, tc := range testcases {
		features, err := tc.importer.Features()
		if err != nil {
			t.Errorf("Error while importing VM features")
		}
		if !reflect.DeepEqual(features, tc.expected) {
			t.Errorf("Unexpected features returned: %v, expected: %v", features, tc.expected)
		}
	}
}
