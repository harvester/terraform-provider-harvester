package image

import (
	"errors"
	"fmt"
	"os"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvsterutil "github.com/harvester/harvester/pkg/util"

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
	processors := util.NewProcessors().
		Tags(&c.Image.Labels).
		Labels(&c.Image.Labels).
		Description(&c.Image.Annotations).
		String(constants.FieldImageDisplayName, &c.Image.Spec.DisplayName, true).
		String(constants.FieldImageSourceType, (*string)(&c.Image.Spec.SourceType), true)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldImageBackend,
			Parser: func(i interface{}) error {
				c.Image.Spec.Backend = harvsterv1.VMIBackend(i.(string))
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldImageURL,
			Parser: func(i interface{}) error {
				imageURL := i.(string)
				if imageURL == "" && c.Image.Spec.SourceType == harvsterv1.VirtualMachineImageSourceTypeDownload {
					return errors.New("must specify image url if image source type is download")
				}
				c.Image.Spec.URL = imageURL
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldImagePVCNamespace,
			Parser: func(i interface{}) error {
				pvcNamespace := i.(string)
				if pvcNamespace == "" && c.Image.Spec.SourceType == harvsterv1.VirtualMachineImageSourceTypeExportVolume {
					return errors.New("must specify image pvc_namespace if image source type is export-from-volume")
				}
				c.Image.Spec.PVCNamespace = pvcNamespace
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldImagePVCName,
			Parser: func(i interface{}) error {
				pvcName := i.(string)
				if pvcName == "" && c.Image.Spec.SourceType == harvsterv1.VirtualMachineImageSourceTypeExportVolume {
					return errors.New("must specify image pvc_name if image source type is export-from-volume")
				}
				c.Image.Spec.PVCName = pvcName
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldImageChecksum,
			Parser: func(i interface{}) error {
				checksum := i.(string)
				c.Image.Spec.Checksum = checksum
				return nil
			},
		},
		{
			Field: constants.FieldImageFilePath,
			Parser: func(i interface{}) error {
				filePath := i.(string)
				if c.Image.Spec.SourceType == harvsterv1.VirtualMachineImageSourceTypeUpload {
					if filePath == "" {
						return errors.New("must specify file_path when source_type is 'upload'")
					}
					info, err := os.Stat(filePath)
					if err != nil {
						return fmt.Errorf("file_path %q is not accessible: %w", filePath, err)
					}
					if !info.Mode().IsRegular() {
						return fmt.Errorf("file_path %q must be a regular file", filePath)
					}
				} else if filePath != "" {
					return fmt.Errorf("file_path must not be set when source_type is %q", c.Image.Spec.SourceType)
				}
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldImageStorageClassName,
			Parser: func(i interface{}) error {
				storageClassName := i.(string)
				if storageClassName == "" && c.Image.Spec.Backend == harvsterv1.VMIBackendCDI {
					return errors.New("must specify image storage_class_name if image backend is cdi")
				}
				c.Image.Annotations[harvsterutil.AnnotationStorageClassName] = storageClassName
				return nil
			},
		}, {
			Field: constants.FieldImageSecurityParameters,
			Parser: func(i interface{}) error {
				if c.Image.Spec.SourceType != harvsterv1.VirtualMachineImageSourceTypeClone {
					return errors.New("security parameters can only be set when source type is 'clone'")
				}

				securityParamsMap := i.(map[string]interface{})

				// Extract values from map
				cryptoOp := securityParamsMap[constants.FieldImageCryptoOperation]
				sourceImageName := securityParamsMap[constants.FieldImageSourceImageName]
				sourceImageNamespace := securityParamsMap[constants.FieldImageSourceImageNamespace]

				cryptoOpStr := cryptoOp.(string)
				sourceImageNameStr := sourceImageName.(string)
				sourceImageNamespaceStr := sourceImageNamespace.(string)

				c.Image.Spec.SecurityParameters = &harvsterv1.VirtualMachineImageSecurityParameters{
					CryptoOperation:      harvsterv1.VirtualMachineImageCryptoOperationType(cryptoOpStr),
					SourceImageName:      sourceImageNameStr,
					SourceImageNamespace: sourceImageNamespaceStr,
				}

				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
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
