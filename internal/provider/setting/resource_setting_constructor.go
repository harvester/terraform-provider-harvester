package setting

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Setting *harvsterv1.Setting
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.Setting.Labels).
		Labels(&c.Setting.Labels).
		Description(&c.Setting.Annotations).
		String(constants.FieldSettingValue, &c.Setting.Value, true)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Setting, nil
}

func newSettingConstructor(setting *harvsterv1.Setting) util.Constructor {
	return &Constructor{
		Setting: setting,
	}
}

func Creator(name string) util.Constructor {
	setting := &harvsterv1.Setting{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}
	return newSettingConstructor(setting)
}

func Updater(setting *harvsterv1.Setting) util.Constructor {
	return newSettingConstructor(setting)
}
