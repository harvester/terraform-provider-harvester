package vgpudevice

import (
	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	VGPUDevice *devicesv1.VGPUDevice
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.VGPUDevice.Labels).
		Labels(&c.VGPUDevice.Labels).
		Description(&c.VGPUDevice.Annotations)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldVGPUDeviceEnabled,
			Parser: func(i any) error {
				enabled := i.(bool)
				c.VGPUDevice.Spec.Enabled = enabled
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVGPUDeviceType,
			Parser: func(i any) error {
				t := i.(string)
				c.VGPUDevice.Spec.VGPUTypeName = t
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVGPUDeviceParentGPUDeviceAddress,
			Parser: func(i any) error {
				a := i.(string)
				c.VGPUDevice.Spec.ParentGPUDeviceAddress = a
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVGPUDeviceNodeName,
			Parser: func(i any) error {
				n := i.(string)
				c.VGPUDevice.Spec.NodeName = n
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
	return c.VGPUDevice, nil
}

func newVGPUDeviceConstructor(vGPUDevice *devicesv1.VGPUDevice) util.Constructor {
	return &Constructor{
		VGPUDevice: vGPUDevice,
	}
}

func Creator(name string) util.Constructor {
	vGPUDevice := &devicesv1.VGPUDevice{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newVGPUDeviceConstructor(vGPUDevice)
}

func Updater(vGPUDevice *devicesv1.VGPUDevice) util.Constructor {
	return newVGPUDeviceConstructor(vGPUDevice)
}
