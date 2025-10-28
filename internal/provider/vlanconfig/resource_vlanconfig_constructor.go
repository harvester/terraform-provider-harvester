package vlanconfig

import (
	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	VLANConfig *harvsternetworkv1.VlanConfig
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.VLANConfig.Labels).
		Labels(&c.VLANConfig.Labels).
		Description(&c.VLANConfig.Annotations).
		String(constants.FieldVLANConfigClusterNetworkName, &c.VLANConfig.Spec.ClusterNetwork, true)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldVLANConfigNodeSelector,
			Parser: func(i interface{}) error {
				c.VLANConfig.Spec.NodeSelector = util.MapMerge(nil, "", i.(map[string]interface{}))
				return nil
			},
		},
		{
			Field: constants.FieldVLANConfigUplink,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})

				// nics
				nicNames := r[constants.FieldUplinkNICs].([]interface{})
				for _, nicName := range nicNames {
					c.VLANConfig.Spec.Uplink.NICs = append(c.VLANConfig.Spec.Uplink.NICs, nicName.(string))
				}

				// bond options
				if c.VLANConfig.Spec.Uplink.BondOptions == nil {
					c.VLANConfig.Spec.Uplink.BondOptions = &harvsternetworkv1.BondOptions{}
				}
				c.VLANConfig.Spec.Uplink.BondOptions.Mode = harvsternetworkv1.BondMode(r[constants.FieldUplinkBondMode].(string))
				c.VLANConfig.Spec.Uplink.BondOptions.Miimon = r[constants.FieldUplinkBondMiimon].(int)

				// link attrs
				if c.VLANConfig.Spec.Uplink.LinkAttrs == nil {
					c.VLANConfig.Spec.Uplink.LinkAttrs = &harvsternetworkv1.LinkAttrs{}
				}
				c.VLANConfig.Spec.Uplink.LinkAttrs.MTU = r[constants.FieldUplinkMTU].(int)
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
	return c.VLANConfig, nil
}

func newVLANConfigConstructor(vlanConfig *harvsternetworkv1.VlanConfig) util.Constructor {
	return &Constructor{
		VLANConfig: vlanConfig,
	}
}

func Creator(name string) util.Constructor {
	vlanConfig := &harvsternetworkv1.VlanConfig{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newVLANConfigConstructor(vlanConfig)
}

func Updater(vlanConfig *harvsternetworkv1.VlanConfig) util.Constructor {
	vlanConfig.Spec.Uplink = harvsternetworkv1.Uplink{}
	return newVLANConfigConstructor(vlanConfig)
}
