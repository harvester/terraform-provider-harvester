package kubeovn_provider_network

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	ProviderNetwork *kubeovnv1.ProviderNetwork
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.ProviderNetwork.Labels).
		Labels(&c.ProviderNetwork.Labels).
		Description(&c.ProviderNetwork.Annotations).
		String(constants.FieldKubeOVNProviderNetDefaultInterface, &c.ProviderNetwork.Spec.DefaultInterface, true).
		Bool(constants.FieldKubeOVNProviderNetExchangeLinkName, &c.ProviderNetwork.Spec.ExchangeLinkName, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNProviderNetCustomInterfaces,
			Parser: func(i interface{}) error {
				ci := i.(map[string]interface{})
				nodesRaw := ci[constants.FieldKubeOVNCustomInterfaceNodes].([]interface{})
				nodes := make([]string, 0, len(nodesRaw))
				for _, n := range nodesRaw {
					nodes = append(nodes, n.(string))
				}
				c.ProviderNetwork.Spec.CustomInterfaces = append(c.ProviderNetwork.Spec.CustomInterfaces, kubeovnv1.CustomInterface{
					Interface: ci[constants.FieldKubeOVNCustomInterfaceInterface].(string),
					Nodes:     nodes,
				})
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNProviderNetExcludeNodes,
			Parser: func(i interface{}) error {
				c.ProviderNetwork.Spec.ExcludeNodes = append(c.ProviderNetwork.Spec.ExcludeNodes, i.(string))
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
	return c.ProviderNetwork, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		ProviderNetwork: &kubeovnv1.ProviderNetwork{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.ProviderNetwork) util.Constructor {
	obj.Spec.CustomInterfaces = nil
	obj.Spec.ExcludeNodes = nil
	return &Constructor{ProviderNetwork: obj}
}
