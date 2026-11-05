package tests

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	testAccKubeOVNProviderNetworkName = "test-acc-ko-pn"
	testAccKubeOVNVlanName            = "test-acc-ko-vlan"

	// Host interface the test provider network is created on. It must exist
	// on every node and carry no traffic the cluster needs.
	testAccKubeOVNNICEnv = "HARVESTER_KUBEOVN_NIC"
	// Existing provider network to put the VLAN on instead of creating one.
	testAccKubeOVNProviderNetworkEnv = "HARVESTER_KUBEOVN_PROVIDER_NETWORK"
	// VLAN ID free on that provider network; the test also uses the next one.
	testAccKubeOVNVlanIDEnv = "HARVESTER_KUBEOVN_VLAN_ID"
)

// testAccKubeOVNUnderlaySetup reads the environment. It returns the
// provider network the VLAN goes on, whether the test creates it, and the
// two VLAN IDs the test applies.
func testAccKubeOVNUnderlaySetup(t *testing.T) (providerNetwork string, create bool, vlanIDs [2]int) {
	providerNetwork = os.Getenv(testAccKubeOVNProviderNetworkEnv)
	create = providerNetwork == ""
	if create {
		if os.Getenv(testAccKubeOVNNICEnv) == "" {
			t.Skipf("set %s to a spare host interface, or %s to an existing provider network", testAccKubeOVNNICEnv, testAccKubeOVNProviderNetworkEnv)
		}
		providerNetwork = testAccKubeOVNProviderNetworkName
	}
	id, err := strconv.Atoi(os.Getenv(testAccKubeOVNVlanIDEnv))
	if err != nil || id <= constants.KubeOVNVlanIDMin || id > constants.KubeOVNVlanIDMax {
		t.Skipf("set %s to a free VLAN ID between %d and %d", testAccKubeOVNVlanIDEnv, constants.KubeOVNVlanIDMin+1, constants.KubeOVNVlanIDMax)
	}
	next := id + 1
	if id == constants.KubeOVNVlanIDMax {
		next = id - 1
	}
	return providerNetwork, create, [2]int{id, next}
}

func testAccKubeOVNUnderlayConfig(providerNetwork string, create bool, vlanID int) string {
	config := ""
	providerRef := strconv.Quote(providerNetwork)
	if create {
		config = fmt.Sprintf(`
resource "harvester_kubeovn_provider_network" "test" {
	name              = %q
	default_interface = %q
}

data "harvester_kubeovn_provider_network" "test" {
	name = harvester_kubeovn_provider_network.test.name
}
`, providerNetwork, os.Getenv(testAccKubeOVNNICEnv))
		providerRef = "harvester_kubeovn_provider_network.test.name"
	}
	return config + fmt.Sprintf(`
resource "harvester_kubeovn_vlan" "test" {
	name             = %q
	vlan_id          = %d
	network_provider = %s
}

data "harvester_kubeovn_vlan" "test" {
	name = harvester_kubeovn_vlan.test.name
}
`, testAccKubeOVNVlanName, vlanID, providerRef)
}

// testAccKubeOVNVlanSpec checks the VLAN stored by kube-ovn and that it is
// still the object with the given UID; an empty UID records it instead.
func testAccKubeOVNVlanSpec(ctx context.Context, t *testing.T, uid *types.UID, vlanID int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		vlan, err := testAccKubeOVNClient(t).KubeovnV1().Vlans().Get(ctx, testAccKubeOVNVlanName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if vlan.Spec.ID != vlanID {
			return fmt.Errorf("vlan %s: spec.id = %d, want %d", vlan.Name, vlan.Spec.ID, vlanID)
		}
		if *uid == "" {
			*uid = vlan.UID
		} else if vlan.UID != *uid {
			return fmt.Errorf("vlan %s was replaced (uid %s, was %s)", vlan.Name, vlan.UID, *uid)
		}
		return nil
	}
}

// testAccCheckKubeOVNUnderlayDestroy checks, without waiting, that the VLAN
// and the provider network the test created are gone.
func testAccCheckKubeOVNUnderlayDestroy(ctx context.Context, t *testing.T, create bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccKubeOVNClient(t).KubeovnV1()
		if _, err := client.Vlans().Get(ctx, testAccKubeOVNVlanName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
			return fmt.Errorf("vlan %s still exists after destroy (err = %v)", testAccKubeOVNVlanName, err)
		}
		if !create {
			return nil
		}
		if _, err := client.ProviderNetworks().Get(ctx, testAccKubeOVNProviderNetworkName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
			return fmt.Errorf("provider network %s still exists after destroy (err = %v)", testAccKubeOVNProviderNetworkName, err)
		}
		return nil
	}
}

// TestAccKubeOVNUnderlay_basic creates a provider network on the interface
// named by HARVESTER_KUBEOVN_NIC, or reuses the one named by
// HARVESTER_KUBEOVN_PROVIDER_NETWORK, and a VLAN on it, then changes the
// VLAN ID, which must update the VLAN in place, then imports both.
func TestAccKubeOVNUnderlay_basic(t *testing.T) {
	const (
		pnResourceName   = constants.ResourceTypeKubeOVNProviderNetwork + ".test"
		vlanResourceName = constants.ResourceTypeKubeOVNVlan + ".test"
	)
	providerNetwork, create, vlanIDs := testAccKubeOVNUnderlaySetup(t)
	ctx := context.Background()
	var vlanUID types.UID

	createChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(vlanResourceName, constants.FieldKubeOVNVlanID, strconv.Itoa(vlanIDs[0])),
		resource.TestCheckResourceAttr(vlanResourceName, constants.FieldKubeOVNVlanProvider, providerNetwork),
		resource.TestCheckResourceAttr(vlanResourceName, constants.FieldKubeOVNVlanStatusConflict, "false"),
		resource.TestCheckResourceAttrPair("data."+vlanResourceName, constants.FieldKubeOVNVlanID, vlanResourceName, constants.FieldKubeOVNVlanID),
		testAccKubeOVNVlanSpec(ctx, t, &vlanUID, vlanIDs[0]),
	}
	steps := []resource.TestStep{}
	if create {
		createChecks = append(createChecks,
			resource.TestCheckResourceAttr(pnResourceName, constants.FieldCommonState, constants.StateCommonReady),
			resource.TestCheckResourceAttr(pnResourceName, constants.FieldKubeOVNProviderNetStatusReady, "true"),
			resource.TestCheckResourceAttr(pnResourceName, constants.FieldKubeOVNProviderNetStatusNotReadyNodes+".#", "0"),
			resource.TestCheckResourceAttrPair("data."+pnResourceName, constants.FieldKubeOVNProviderNetDefaultInterface, pnResourceName, constants.FieldKubeOVNProviderNetDefaultInterface),
		)
	}
	steps = append(steps,
		resource.TestStep{
			Config: testAccKubeOVNUnderlayConfig(providerNetwork, create, vlanIDs[0]),
			Check:  resource.ComposeAggregateTestCheckFunc(createChecks...),
		},
		resource.TestStep{
			Config: testAccKubeOVNUnderlayConfig(providerNetwork, create, vlanIDs[1]),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(vlanResourceName, constants.FieldKubeOVNVlanID, strconv.Itoa(vlanIDs[1])),
				testAccKubeOVNVlanSpec(ctx, t, &vlanUID, vlanIDs[1]),
			),
		},
		resource.TestStep{
			ResourceName:      vlanResourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
	)
	if create {
		steps = append(steps, resource.TestStep{
			ResourceName:      pnResourceName,
			ImportState:       true,
			ImportStateVerify: true,
			// kube-ovn adds the VLAN to status.vlans after the provider
			// network was last read.
			ImportStateVerifyIgnore: []string{constants.FieldKubeOVNProviderNetStatusVlans},
		})
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if _, err := testAccKubeOVNClient(t).KubeovnV1().ProviderNetworks().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
				t.Skipf("KubeOVN is not available: %v", err)
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubeOVNUnderlayDestroy(ctx, t, create),
		Steps:        steps,
	})
}
