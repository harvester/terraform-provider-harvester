package util

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceStatesSet(d *schema.ResourceData, getter *importer.StateGetter) error {
	for key, value := range getter.States {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	d.SetId(getter.ID)
	return nil
}
