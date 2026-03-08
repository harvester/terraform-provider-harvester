package importer

import (
	"testing"

	"reflect"

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

func TestAccessCredentialsImport(t *testing.T) {
	// VM with SSH public key via qemuGuestAgent
	vm := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					AccessCredentials: []kubevirtv1.AccessCredential{
						{
							SSHPublicKey: &kubevirtv1.SSHPublicKeyAccessCredential{
								Source: kubevirtv1.SSHPublicKeyAccessCredentialSource{
									Secret: &kubevirtv1.AccessCredentialSecretSource{SecretName: "my-ssh-keys"}, // #nosec G101
								},
								PropagationMethod: kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
									QemuGuestAgent: &kubevirtv1.QemuGuestAgentSSHPublicKeyAccessCredentialPropagation{
										Users: []string{"root", "admin"},
									},
								},
							},
						},
						{
							UserPassword: &kubevirtv1.UserPasswordAccessCredential{
								Source: kubevirtv1.UserPasswordAccessCredentialSource{
									Secret: &kubevirtv1.AccessCredentialSecretSource{SecretName: "my-passwords"},
								},
								PropagationMethod: kubevirtv1.UserPasswordAccessCredentialPropagationMethod{
									QemuGuestAgent: &kubevirtv1.QemuGuestAgentUserPasswordAccessCredentialPropagation{},
								},
							},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}
	acs := importer.AccessCredentials()

	if len(acs) != 2 {
		t.Fatalf("AccessCredentials() returned %d entries, want 2", len(acs))
	}

	// Check SSH entry
	sshList := acs[0][constants.FieldAccessCredentialSSHPublicKey].([]interface{})
	if len(sshList) != 1 {
		t.Fatalf("SSH entry has %d items, want 1", len(sshList))
	}
	ssh := sshList[0].(map[string]interface{})
	if ssh[constants.FieldAccessCredentialSecretName] != "my-ssh-keys" {
		t.Errorf("SSH secret_name = %q, want %q", ssh[constants.FieldAccessCredentialSecretName], "my-ssh-keys")
	}
	if ssh[constants.FieldAccessCredentialPropagationMethod] != "qemuGuestAgent" {
		t.Errorf("SSH propagation_method = %q, want %q", ssh[constants.FieldAccessCredentialPropagationMethod], "qemuGuestAgent")
	}
	users := ssh[constants.FieldAccessCredentialUsers].([]string)
	if !reflect.DeepEqual(users, []string{"root", "admin"}) {
		t.Errorf("SSH users = %v, want [root admin]", users)
	}

	// Check UserPassword entry
	pwList := acs[1][constants.FieldAccessCredentialUserPassword].([]interface{})
	if len(pwList) != 1 {
		t.Fatalf("UserPassword entry has %d items, want 1", len(pwList))
	}
	pw := pwList[0].(map[string]interface{})
	if pw[constants.FieldAccessCredentialSecretName] != "my-passwords" {
		t.Errorf("UserPassword secret_name = %q, want %q", pw[constants.FieldAccessCredentialSecretName], "my-passwords")
	}

	// Test empty access credentials
	vmEmpty := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{},
			},
		},
	}
	importerEmpty := &VMImporter{VirtualMachine: vmEmpty}
	if got := importerEmpty.AccessCredentials(); len(got) != 0 {
		t.Errorf("AccessCredentials() empty VM returned %d entries, want 0", len(got))
	}
}

func TestDNSImport(t *testing.T) {
	// VM with DNS policy and config
	vm := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					DNSPolicy: corev1.DNSNone,
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{"8.8.8.8", "8.8.4.4"},
						Searches:    []string{"example.com"},
						Options: []corev1.PodDNSConfigOption{
							{Name: "ndots", Value: ptr.To("5")},
							{Name: "single-request"},
						},
					},
				},
			},
		},
	}
	importer := &VMImporter{VirtualMachine: vm}

	if policy := importer.DNSPolicy(); policy != "None" {
		t.Errorf("DNSPolicy() = %q, want %q", policy, "None")
	}

	dc := importer.DNSConfig()
	if len(dc) != 1 {
		t.Fatalf("DNSConfig() returned %d entries, want 1", len(dc))
	}
	ns := dc[0][constants.FieldDNSConfigNameservers].([]string)
	if !reflect.DeepEqual(ns, []string{"8.8.8.8", "8.8.4.4"}) {
		t.Errorf("DNSConfig nameservers = %v, want [8.8.8.8 8.8.4.4]", ns)
	}
	searches := dc[0][constants.FieldDNSConfigSearches].([]string)
	if !reflect.DeepEqual(searches, []string{"example.com"}) {
		t.Errorf("DNSConfig searches = %v, want [example.com]", searches)
	}
	opts := dc[0][constants.FieldDNSConfigOptions].([]map[string]interface{})
	if len(opts) != 2 {
		t.Fatalf("DNSConfig options has %d items, want 2", len(opts))
	}
	if opts[0][constants.FieldDNSOptionName] != "ndots" || opts[0][constants.FieldDNSOptionValue] != "5" {
		t.Errorf("DNS option 0 = %v, want {name:ndots, value:5}", opts[0])
	}
	if opts[1][constants.FieldDNSOptionName] != "single-request" || opts[1][constants.FieldDNSOptionValue] != "" {
		t.Errorf("DNS option 1 = %v, want {name:single-request, value:\"\"}", opts[1])
	}

	// VM without DNS config
	vmNoDNS := &kubevirtv1.VirtualMachine{
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{},
			},
		},
	}
	importerNoDNS := &VMImporter{VirtualMachine: vmNoDNS}
	if policy := importerNoDNS.DNSPolicy(); policy != "" {
		t.Errorf("DNSPolicy() empty = %q, want empty", policy)
	}
	if dc := importerNoDNS.DNSConfig(); dc != nil {
		t.Errorf("DNSConfig() empty = %v, want nil", dc)
	}
}

func TestConfigMapSecretDiskImport(t *testing.T) {
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{
								{
									Name: "cm-disk",
									DiskDevice: kubevirtv1.DiskDevice{
										Disk: &kubevirtv1.DiskTarget{Bus: "virtio"},
									},
								},
								{
									Name: "sec-disk",
									DiskDevice: kubevirtv1.DiskDevice{
										Disk: &kubevirtv1.DiskTarget{Bus: "virtio"},
									},
								},
							},
						},
					},
					Volumes: []kubevirtv1.Volume{
						{
							Name: "cm-disk",
							VolumeSource: kubevirtv1.VolumeSource{
								ConfigMap: &kubevirtv1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
								},
							},
						},
						{
							Name: "sec-disk",
							VolumeSource: kubevirtv1.VolumeSource{
								Secret: &kubevirtv1.SecretVolumeSource{
									SecretName: "my-secret",
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
	if disks[0][constants.FieldDiskConfigMapName] != "my-config" {
		t.Errorf("disk 0 configmap_name = %v, want my-config", disks[0][constants.FieldDiskConfigMapName])
	}
	if disks[1][constants.FieldDiskSecretName] != "my-secret" {
		t.Errorf("disk 1 secret_name = %v, want my-secret", disks[1][constants.FieldDiskSecretName])
	}
}
