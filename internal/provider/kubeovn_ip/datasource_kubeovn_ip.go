package kubeovn_ip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIPPodName: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPNamespace: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPSubnet: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPIPAddress: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPMacAddress: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPNodeName: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}

func DataSourceKubeOVNIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubeOVNIPRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceKubeOVNIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.KubeOVNClient.KubeovnV1().IPs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(dataSourceKubeOVNIPImport(d, obj))
}

func dataSourceKubeOVNIPImport(d *schema.ResourceData, obj *kubeovnv1.IP) error {
	stateGetter, err := importer.ResourceKubeOVNIPStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
