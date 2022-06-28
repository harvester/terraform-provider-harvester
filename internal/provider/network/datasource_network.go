package network

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceNetworkImport(d, obj))
}
