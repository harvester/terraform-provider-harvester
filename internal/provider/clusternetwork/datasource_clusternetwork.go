package clusternetwork

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceClusterNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterNetworkRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceClusterNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	name := d.Get(constants.FieldCommonName).(string)
	clusterNetwork, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceClusterNetworkImport(d, clusterNetwork))
}
