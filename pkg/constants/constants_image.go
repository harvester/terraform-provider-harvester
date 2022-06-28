package constants

const (
	ResourceTypeImage = "harvester_image"

	FieldImageDisplayName      = "display_name"
	FieldImageURL              = "url"
	FieldImagePVCNamespace     = "pvc_namespace"
	FieldImagePVCName          = "pvc_name"
	FieldImageSourceType       = "source_type"
	FieldImageProgress         = "progress"
	FieldImageSize             = "size"
	FieldImageStorageClassName = "storage_class_name"

	StateImageUploading    = "Uploading"
	StateImageDownloading  = "Downloading"
	StateImageExporting    = "Exporting"
	StateImageInitializing = "Initializing"
)
