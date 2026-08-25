package addon

import (
	"fmt"
	"strings"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvsterutil "github.com/harvester/harvester/pkg/util"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const enabledLabelValue = "true"

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Addon *harvsterv1.Addon
	// Create is set when the addon does not exist on the cluster and a custom
	// addon object is being created.
	Create bool

	// originalChart backs the client-side immutability guard: the Harvester
	// webhook rejects chart changes on existing addons.
	originalChart string
	// preservedLabels keeps the addon.harvesterhci.io/* labels (notably the
	// experimental label required to delete custom addons) that the shared
	// Labels processor would otherwise wipe.
	preservedLabels map[string]string
}

func newConstructor(addon *harvsterv1.Addon, create bool) *Constructor {
	c := &Constructor{
		Addon:           addon,
		Create:          create,
		originalChart:   addon.Spec.Chart,
		preservedLabels: map[string]string{},
	}
	for key, value := range addon.Labels {
		if strings.HasPrefix(key, harvsterutil.AddonPrefix+"/") {
			c.preservedLabels[key] = value
		}
	}
	return c
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.Addon.Labels).
		Labels(&c.Addon.Labels).
		Description(&c.Addon.Annotations).
		// enabled must be a required processor: non-required processors are read
		// with GetOk, which skips zero values, so `enabled = false` would never
		// be applied on update and the addon could not be disabled in place.
		Bool(constants.FieldAddonEnabled, &c.Addon.Spec.Enabled, true).
		String(constants.FieldAddonValuesContent, &c.Addon.Spec.ValuesContent, false).
		String(constants.FieldAddonRepo, &c.Addon.Spec.Repo, false).
		String(constants.FieldAddonChart, &c.Addon.Spec.Chart, false).
		String(constants.FieldAddonVersion, &c.Addon.Spec.Version, false)
}

func (c *Constructor) Validate() error {
	if c.Create {
		if c.Addon.Spec.Repo == "" || c.Addon.Spec.Chart == "" || c.Addon.Spec.Version == "" {
			return fmt.Errorf("addon %s/%s does not exist on the cluster; "+
				"creating a custom addon requires repo, chart and version to be set. "+
				"Built-in addons already exist and are adopted automatically, "+
				"check the name and namespace if you expected adoption",
				c.Addon.Namespace, c.Addon.Name)
		}
		return nil
	}
	if c.originalChart != "" && c.Addon.Spec.Chart != c.originalChart {
		return fmt.Errorf("chart is immutable on an existing addon "+
			"(the Harvester webhook rejects chart changes); destroy and recreate "+
			"the addon under a new name to change its chart (current: %q)",
			c.originalChart)
	}
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	// Restore the addon.harvesterhci.io/* labels wiped by the shared Labels
	// processor; without the experimental label a custom addon could never be
	// deleted again.
	if len(c.preservedLabels) > 0 {
		if c.Addon.Labels == nil {
			c.Addon.Labels = map[string]string{}
		}
		for key, value := range c.preservedLabels {
			c.Addon.Labels[key] = value
		}
	}
	return c.Addon, nil
}

func Creator(namespace, name string) util.Constructor {
	addon := &harvsterv1.Addon{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	// Mark the object as owned by this provider (destroy deletes it) and as
	// experimental: the Harvester webhook only allows deleting addons that
	// carry the experimental label.
	addon.Annotations[constants.AnnotationAddonAutoDelete] = enabledLabelValue
	addon.Labels[harvsterutil.AddonExperimentalLabel] = enabledLabelValue
	return newConstructor(addon, true)
}

func Updater(addon *harvsterv1.Addon) util.Constructor {
	return newConstructor(addon, false)
}
