package pcidevice

import (
	"context"
	"fmt"

	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	PCIDevice      *devicesv1.PCIDevice
	PCIDeviceClaim *devicesv1.PCIDeviceClaim

	Enabled bool

	Error error
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Description(&c.PCIDevice.Annotations)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldPCIDevicePassthroughEnabled,
			Parser: func(i interface{}) error {
				enabled := i.(bool)
				c.Enabled = enabled
				return nil
			},
		},
	}

	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	if c.Enabled && c.PCIDeviceClaim == nil {
		return fmt.Errorf("Can't enable PCI passthrough for PCI device %s", c.PCIDevice.Name)
	}

	return c.Error
}

func (c *Constructor) Result() (interface{}, error) {
	return c.PCIDeviceClaim, c.Error
}

func Creator(ctx context.Context, client *client.Client, pcidevice *devicesv1.PCIDevice) util.Constructor {
	pciDeviceClaim, err := findPCIDeviceClaim(ctx, client, pcidevice)
	if pciDeviceClaim != nil {
		err = fmt.Errorf("PCIDeviceClaim %s for PCIDevice %s already exists", pciDeviceClaim.Name, pcidevice.Name)
	} else if err != nil {
		// do nothing, propagate the error through c.Error
	} else {
		pciDeviceClaim = claimPCIDevice(pcidevice)
	}

	return newPCIDeviceConstructor(pcidevice, pciDeviceClaim, err)
}

func Updater(ctx context.Context, client *client.Client, pcidevice *devicesv1.PCIDevice) util.Constructor {
	pciDeviceClaim, err := findPCIDeviceClaim(ctx, client, pcidevice)
	if pciDeviceClaim == nil {
		pciDeviceClaim = claimPCIDevice(pcidevice)
	}

	return newPCIDeviceConstructor(pcidevice, pciDeviceClaim, err)
}

func Deleter(ctx context.Context, client *client.Client, pcidevice *devicesv1.PCIDevice) util.Constructor {
	pciDeviceClaim, err := findPCIDeviceClaim(ctx, client, pcidevice)
	return newPCIDeviceConstructor(pcidevice, pciDeviceClaim, err)
}

func newPCIDeviceConstructor(pciDevice *devicesv1.PCIDevice, pciDeviceClaim *devicesv1.PCIDeviceClaim, err error) util.Constructor {
	return &Constructor{
		PCIDevice:      pciDevice,
		PCIDeviceClaim: pciDeviceClaim,

		Enabled: false,

		Error: err,
	}
}

func findPCIDeviceClaim(ctx context.Context, client *client.Client, pcidevice *devicesv1.PCIDevice) (*devicesv1.PCIDeviceClaim, error) {
	var pciDeviceClaim *devicesv1.PCIDeviceClaim
	pciDeviceClaim = nil

	claims, err := client.HarvesterDeviceClient.DevicesV1beta1().PCIDeviceClaims().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, claim := range claims.Items {
			for _, owner := range claim.OwnerReferences {
				if owner.Name == pcidevice.Name && owner.UID == pcidevice.UID {
					pciDeviceClaim = &claim
				}
			}
		}
	}
	return pciDeviceClaim, err
}

func claimPCIDevice(pcidevice *devicesv1.PCIDevice) *devicesv1.PCIDeviceClaim {
	PCIDeviceClaim := devicesv1.PCIDeviceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pcidevice.Name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "devices.harvesterhci.io/v1beta1",
					Kind:       "PCIDevice",
					Name:       pcidevice.Name,
					UID:        pcidevice.UID,
				},
			},
		},
		Spec: devicesv1.PCIDeviceClaimSpec{
			Address:  pcidevice.Status.Address,
			NodeName: pcidevice.Status.NodeName,
			UserName: "admin",
		},
	}
	return &PCIDeviceClaim
}
