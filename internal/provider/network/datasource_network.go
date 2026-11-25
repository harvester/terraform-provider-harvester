package network

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func DataSourceNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(dataSourceNetworkImport(d, obj))
}

func dataSourceNetworkImport(d *schema.ResourceData, obj *nadv1.NetworkAttachmentDefinition) error {
	stateGetter, err := importer.ResourceNetworkStateGetter(obj)
	if err != nil {
		return err
	}
	return util.DataSourceStatesSet(d, stateGetter)
}
