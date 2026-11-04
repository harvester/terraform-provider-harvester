package sriovgpudevice

import (
	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	SRIOVGPUDevice *devicesv1.SRIOVGPUDevice
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.SRIOVGPUDevice.Labels).
		Labels(&c.SRIOVGPUDevice.Labels).
		Description(&c.SRIOVGPUDevice.Annotations)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldSRIOVGPUDeviceEnabled,
			Parser: func(i any) error {
				enabled := i.(bool)
				c.SRIOVGPUDevice.Spec.Enabled = enabled
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

func (c *Constructor) Result() (any, error) {
	return c.SRIOVGPUDevice, nil
}

func newSRIOVGPUDeviceConstructor(SRIOVGPUDevice *devicesv1.SRIOVGPUDevice) util.Constructor {
	return &Constructor{
		SRIOVGPUDevice: SRIOVGPUDevice,
	}
}

func Creator(name string) util.Constructor {
	SRIOVGPUDevice := &devicesv1.SRIOVGPUDevice{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newSRIOVGPUDeviceConstructor(SRIOVGPUDevice)
}

func Updater(SRIOVGPUDevice *devicesv1.SRIOVGPUDevice) util.Constructor {
	return newSRIOVGPUDeviceConstructor(SRIOVGPUDevice)
}
