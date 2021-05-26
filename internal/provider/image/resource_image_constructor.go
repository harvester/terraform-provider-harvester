package image

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Image *harvsterv1.VirtualMachineImage
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().Tags(&c.Image.Labels).Description(&c.Image.Annotations).
		String(constants.FieldImageDisplayName, &c.Image.Spec.DisplayName, true).
		String(constants.FieldImageURL, &c.Image.Spec.URL, true)
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Image, nil
}

func newImageConstructor(image *harvsterv1.VirtualMachineImage) util.Constructor {
	return &Constructor{
		Image: image,
	}
}

func Creator(namespace, name string) util.Constructor {
	image := &harvsterv1.VirtualMachineImage{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newImageConstructor(image)
}

func Updater(image *harvsterv1.VirtualMachineImage) util.Constructor {
	return newImageConstructor(image)
}
