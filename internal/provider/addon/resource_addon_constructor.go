package addon

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Addon *harvsterv1.Addon
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.Addon.Labels).
		Labels(&c.Addon.Labels).
		Description(&c.Addon.Annotations).
		Bool(constants.FieldAddonEnabled, &c.Addon.Spec.Enabled, false).
		String(constants.FieldAddonValuesContent, &c.Addon.Spec.ValuesContent, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Addon, nil
}

func Updater(addon *harvsterv1.Addon) util.Constructor {
	return &Constructor{
		Addon: addon,
	}
}
