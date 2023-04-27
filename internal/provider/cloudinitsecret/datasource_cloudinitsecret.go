package cloudinitsecret

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceCloudInitSecret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCloudInitSecretRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceCloudInitSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	secret, err := c.KubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceCloudInitSecretImport(d, secret))
}
