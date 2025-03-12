package config

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
)

type Config struct {
	Bootstrap   bool
	APIURL      string
	KubeConfig  string
	KubeContext string
}

func (c *Config) K8sClient() (*client.Client, error) {
	return client.NewClient(c.KubeConfig, c.KubeContext)
}

func (c *Config) CheckVersion() error {
	client, err := c.K8sClient()
	if err != nil {
		return err
	}

	// check harvester version from settings
	serverVersion, err := client.HarvesterClient.HarvesterhciV1beta1().Settings().Get(context.Background(), "server-version", metav1.GetOptions{})
	if err != nil {
		return err
	}
	// harvester version v1.0-head, v1.0.2, v1.0.3 is not supported
	if strings.HasPrefix(serverVersion.Value, "v1.0") {
		return fmt.Errorf("current Harvester server version is %s, the minimum supported version is v1.1.0", serverVersion.Value)
	}
	return nil
}
