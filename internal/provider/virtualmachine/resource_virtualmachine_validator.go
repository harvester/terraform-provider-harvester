package virtualmachine

import (
	"encoding/base64"
	"fmt"
	"strings"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	userDataSecretKey = "userdata"
)

func (c *Constructor) getKeyPairs(sshNames []string, defaultNamespace string) ([]*harvsterv1.KeyPair, error) {
	keyPairs := make([]*harvsterv1.KeyPair, 0, len(sshNames))
	for _, keyPairNamespacedName := range sshNames {
		keyPairNamespace, keyPairName, err := helper.NamespacedNamePartsByDefault(keyPairNamespacedName, defaultNamespace)
		if err != nil {
			return nil, err
		}
		keyPair, err := c.Client.HarvesterClient.HarvesterhciV1beta1().KeyPairs(keyPairNamespace).Get(c.Context, keyPairName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		keyPairs = append(keyPairs, keyPair)
	}
	return keyPairs, nil
}

func (c *Constructor) checkKeyPairsInCloudInit(keyPairs []*harvsterv1.KeyPair) error {
	volumes := c.Builder.VirtualMachine.Spec.Template.Spec.Volumes
	for _, volume := range volumes {
		if volume.CloudInitNoCloud != nil {
			if volume.CloudInitNoCloud.UserDataSecretRef != nil {
				userDataSecretName := volume.CloudInitNoCloud.UserDataSecretRef.Name
				userDataSecretNamespace := c.Builder.VirtualMachine.Namespace
				return c.checkKeyPairsInUserDataSecret(userDataSecretNamespace, userDataSecretName, keyPairs)
			} else if volume.CloudInitNoCloud.UserDataBase64 != "" {
				return c.checkKeyPairsInUserDataBase64(volume.CloudInitNoCloud.UserDataBase64, keyPairs)
			} else {
				return checkKeyPairsInUserData([]byte(volume.CloudInitNoCloud.UserData), keyPairs)
			}
		} else if volume.CloudInitConfigDrive != nil {
			if volume.CloudInitConfigDrive.UserDataSecretRef != nil {
				userDataSecretName := volume.CloudInitConfigDrive.UserDataSecretRef.Name
				userDataSecretNamespace := c.Builder.VirtualMachine.Namespace
				return c.checkKeyPairsInUserDataSecret(userDataSecretNamespace, userDataSecretName, keyPairs)
			} else if volume.CloudInitConfigDrive.UserDataBase64 != "" {
				return c.checkKeyPairsInUserDataBase64(volume.CloudInitConfigDrive.UserDataBase64, keyPairs)
			} else {
				return checkKeyPairsInUserData([]byte(volume.CloudInitConfigDrive.UserData), keyPairs)
			}
		}
	}
	return checkKeyPairsInUserData([]byte{}, keyPairs)
}

func (c *Constructor) checkKeyPairsInUserDataSecret(userDataSecretNamespace, userDataSecretName string, keyPairs []*harvsterv1.KeyPair) error {
	userDataSecret, err := c.Client.KubeClient.CoreV1().Secrets(userDataSecretNamespace).Get(c.Context, userDataSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return checkKeyPairsInUserData(userDataSecret.Data[userDataSecretKey], keyPairs)
}

func (c *Constructor) checkKeyPairsInUserDataBase64(userdataBase64Content string, keyPairs []*harvsterv1.KeyPair) error {
	userDataContent, err := base64.StdEncoding.DecodeString(userdataBase64Content)
	if err != nil {
		return err
	}
	return checkKeyPairsInUserData(userDataContent, keyPairs)
}

func checkKeyPairsInUserData(userdataContent []byte, keyPairs []*harvsterv1.KeyPair) error {
	userData := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(userdataContent, &userData); err != nil {
		return fmt.Errorf("failed to parser cloud-int userdata, err: %+v", err)
	}
	var missingKeyPairs []string
	for _, keyPair := range keyPairs {
		if ok := inCloudConfig("", userData, "ssh_authorized_keys", keyPair.Spec.PublicKey); !ok {
			missingKeyPairs = append(missingKeyPairs, helper.BuildNamespacedName(keyPair.Namespace, keyPair.Name))
		}
	}
	if len(missingKeyPairs) > 0 {
		return fmt.Errorf(`missing ssh public keys in cloud-int userdata ssh_authorized_keys section, ssh_keys: [%s].
Either remove unused ssh keys from "ssh_keys" or add ssh public keys to cloud-int userdata ssh_authorized_keys section. `, strings.Join(missingKeyPairs, ","))
	}
	return nil
}

func inCloudConfig(parentKey, parent, key, value interface{}) bool {
	switch section := parent.(type) {
	case map[interface{}]interface{}:
		for k, v := range section {
			if inCloudConfig(k, v, key, value) {
				return true
			}
		}
	case []interface{}:
		for _, v := range section {
			if inCloudConfig(parentKey, v, key, value) {
				return true
			}
		}
	case interface{}:
		if parentKey == key && section == value {
			return true
		}
	}
	return false
}
