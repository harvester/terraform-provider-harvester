package importer

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			// VM with CPU model specified
			importer: &VMImporter{
				VirtualMachine: &kubevirtv1.VirtualMachine{
					Spec: kubevirtv1.VirtualMachineSpec{
						Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
							Spec: kubevirtv1.VirtualMachineInstanceSpec{
								Domain: kubevirtv1.DomainSpec{
									CPU: &kubevirtv1.CPU{
										Cores: 4,
										Model: "host-model",
									},
								},
							},
						},
					},
				},
			},
			expectedCores: 4,
			expectedModel: "host-model",
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
