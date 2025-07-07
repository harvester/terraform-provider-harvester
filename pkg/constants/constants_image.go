package constants

const (
	ResourceTypeImage = "harvester_image"

	FieldImageDisplayName            = "display_name"
	FieldImageURL                    = "url"
	FieldImagePVCNamespace           = "pvc_namespace"
	FieldImagePVCName                = "pvc_name"
	FieldImageSourceType             = "source_type"
	FieldImageProgress               = "progress"
	FieldImageSize                   = "size"
	FieldImageStorageClassName       = "storage_class_name"
	FieldImageStorageClassParameters = "storage_class_parameters"
	FieldImageVolumeStorageClassName = "volume_storage_class_name"
	FieldImageSecurityParameters     = "security_parameters"
	FieldImageCryptoOperation        = "crypto_operation"
	FieldImageSourceImageName        = "source_image_name"
	FieldImageSourceImageNamespace   = "source_image_namespace"

	StateImageUploading    = "Uploading"
	StateImageDownloading  = "Downloading"
	StateImageExporting    = "Exporting"
	StateImageInitializing = "Initializing"
	StateImageTerminating  = "Terminating"
)
