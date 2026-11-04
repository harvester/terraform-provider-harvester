package sriovnetworkdevice

import (
	"context"

	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	ctx context.Context

	SRIOVNetworkDevice *devicesv1.SRIOVNetworkDevice
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.SRIOVNetworkDevice.Labels).
		Labels(&c.SRIOVNetworkDevice.Labels).
		Description(&c.SRIOVNetworkDevice.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldSRIOVNetworkDeviceNumVFs,
			Parser: func(i interface{}) error {
				num := i.(int)
				c.SRIOVNetworkDevice.Spec.NumVFs = num
				return nil
			},
			Required: true,
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.SRIOVNetworkDevice, nil
}

func newSRIOVNetworkDeviceConstructor(ctx context.Context, SRIOVNetworkDevice *devicesv1.SRIOVNetworkDevice) util.Constructor {
	return &Constructor{
		ctx:                ctx,
		SRIOVNetworkDevice: SRIOVNetworkDevice,
	}
}

func Creator(ctx context.Context, name string) util.Constructor {
	SRIOVNetworkDevice := &devicesv1.SRIOVNetworkDevice{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newSRIOVNetworkDeviceConstructor(ctx, SRIOVNetworkDevice)
}

func Updater(ctx context.Context, SRIOVNetworkDevice *devicesv1.SRIOVNetworkDevice) util.Constructor {
	return newSRIOVNetworkDeviceConstructor(ctx, SRIOVNetworkDevice)
}
