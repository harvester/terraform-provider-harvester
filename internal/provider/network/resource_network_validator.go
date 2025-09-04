package network

import (
	"time"

	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func (c *Constructor) waitForClusterNetworkReady(name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonReady},
		Refresh:    c.clusterNetworkStateRefresh(name),
		Timeout:    timeout,
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(c.Context)
	return err
}

func (c *Constructor) clusterNetworkStateRefresh(name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := c.Client.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Get(c.Context, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		state := constants.StateCommonActive
		for _, condition := range obj.Status.Conditions {
			if condition.Type == harvsternetworkv1.Ready && condition.Status == corev1.ConditionTrue {
				state = constants.StateCommonReady
				break
			}
		}
		return obj, state, nil
	}
}
