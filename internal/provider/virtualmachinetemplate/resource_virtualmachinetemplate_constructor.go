package virtualmachinetemplate

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Template *harvsterv1.VirtualMachineTemplate
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.Template.Labels).
		Labels(&c.Template.Labels).
		Description(&c.Template.Annotations).
		String(constants.FieldVirtualMachineTemplateDefaultVersionID, &c.Template.Spec.DefaultVersionID, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Template, nil
}

func newConstructor(template *harvsterv1.VirtualMachineTemplate) util.Constructor {
	return &Constructor{
		Template: template,
	}
}

func Creator(namespace, name string) util.Constructor {
	template := &harvsterv1.VirtualMachineTemplate{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newConstructor(template)
}

func Updater(template *harvsterv1.VirtualMachineTemplate) util.Constructor {
	return newConstructor(template)
}
