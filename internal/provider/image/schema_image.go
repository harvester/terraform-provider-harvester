package image

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	URLDescription = "supports the `raw` and `qcow2` image formats which are supported by [qemu](https://www.qemu.org/docs/master/system/images.html#disk-image-file-formats). Bootable ISO images can also be used and are treated like `raw` images."
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldImageDisplayName: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		constants.FieldImageURL: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  URLDescription,
		},
		constants.FieldImageSize: {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}
