package image

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
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
		constants.FieldImageBackend: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(harvsterv1.VMIBackendBackingImage),
			ValidateFunc: validation.StringInSlice([]string{
				string(harvsterv1.VMIBackendBackingImage),
				string(harvsterv1.VMIBackendCDI),
			}, false),
			Description: "The backend type of the image, either 'backing-image' or 'cdi'.",
		},
		constants.FieldImageDisplayName: {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		constants.FieldImageURL: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  URLDescription,
		},
		constants.FieldImagePVCNamespace: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: util.IsValidName,
		},
		constants.FieldImagePVCName: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: util.IsValidName,
		},
		constants.FieldImageSourceType: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				string(harvsterv1.VirtualMachineImageSourceTypeDownload),
				string(harvsterv1.VirtualMachineImageSourceTypeUpload),
				string(harvsterv1.VirtualMachineImageSourceTypeExportVolume),
				string(harvsterv1.VirtualMachineImageSourceTypeClone),
			}, false),
		},
		constants.FieldImageProgress: {
			Type:     schema.TypeInt,
			Computed: true,
		},
		constants.FieldImageSize: {
			Type:     schema.TypeInt,
			Computed: true,
		},
		constants.FieldImageStorageClassName: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: util.IsValidName,
		},
		constants.FieldImageStorageClassParameters: {
			Type:     schema.TypeMap,
			Computed: true,
		},
		constants.FieldImageVolumeStorageClassName: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldImageSecurityParameters: {
			Type:         schema.TypeMap,
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			Description:  "Security parameters for encryption/decryption operations. When specified, source_type must be 'clone'. Required keys: crypto_operation, source_image_name, source_image_namespace",
			ValidateFunc: validateSecurityParameters,
		},
		constants.FieldImageChecksum: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "SHA-512 checksum of the image",
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	s := util.DataSourceSchemaWrap(Schema())
	s[constants.FieldCommonName].Required = false
	s[constants.FieldCommonName].Optional = true
	s[constants.FieldImageDisplayName].Computed = false
	s[constants.FieldImageDisplayName].Optional = true
	return s
}
