package cloudinitsecret

import (
	"encoding/base64"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	CloudInitSecret *corev1.Secret
}

func (c *Constructor) Setup() util.Processors {
	if c.CloudInitSecret.StringData == nil {
		c.CloudInitSecret.StringData = map[string]string{}
	}
	c.CloudInitSecret.Data = map[string][]byte{}

	processors := util.NewProcessors().Tags(&c.CloudInitSecret.Labels).Description(&c.CloudInitSecret.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldCloudInitSecretUserData,
			Parser: func(i interface{}) error {
				c.CloudInitSecret.StringData[constants.SecretDataKeyUserData] = i.(string)
				return nil
			},
		},
		{
			Field: constants.FieldCloudInitSecretUserDataBase64,
			Parser: func(i interface{}) error {
				value, err := base64.StdEncoding.DecodeString(i.(string))
				if err != nil {
					return fmt.Errorf("failed to decode %s string: %w", constants.FieldCloudInitSecretUserDataBase64, err)
				}
				c.CloudInitSecret.StringData[constants.SecretDataKeyUserData] = string(value)
				return nil
			},
		},
		{
			Field: constants.FieldCloudInitSecretNetworkData,
			Parser: func(i interface{}) error {
				c.CloudInitSecret.StringData[constants.SecretDataKeyNetworkData] = i.(string)
				return nil
			},
		},
		{
			Field: constants.FieldCloudInitSecretNetworkDataBase64,
			Parser: func(i interface{}) error {
				value, err := base64.StdEncoding.DecodeString(i.(string))
				if err != nil {
					return fmt.Errorf("failed to decode %s string: %w", constants.FieldCloudInitSecretNetworkDataBase64, err)
				}
				c.CloudInitSecret.StringData[constants.SecretDataKeyNetworkData] = string(value)
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
	return c.CloudInitSecret, nil
}

func newCloudInitSecretConstructor(cloudInitSecret *corev1.Secret) util.Constructor {
	return &Constructor{
		CloudInitSecret: cloudInitSecret,
	}
}

func Creator(namespace, name string) util.Constructor {
	cloudInitSecret := &corev1.Secret{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newCloudInitSecretConstructor(cloudInitSecret)
}

func Updater(cloudInitSecret *corev1.Secret) util.Constructor {
	return newCloudInitSecretConstructor(cloudInitSecret)
}
