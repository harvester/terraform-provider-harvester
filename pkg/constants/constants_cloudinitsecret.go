package constants

const (
	ResourceTypeCloudInitSecret = "harvester_cloudinit_secret"

	FieldCloudInitSecretUserData          = "user_data"
	FieldCloudInitSecretNetworkData       = "network_data"
	FieldCloudInitSecretUserDataBase64    = "user_data_base64" // #nosec G101
	FieldCloudInitSecretNetworkDataBase64 = "network_data_base64"

	SecretDataKeyUserData    = "userdata"
	SecretDataKeyNetworkData = "networkdata"
)
