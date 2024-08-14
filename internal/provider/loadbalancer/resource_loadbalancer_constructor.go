package loadbalancer

import (
	"fmt"

	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	LoadBalancer *loadbalancerv1.LoadBalancer
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.LoadBalancer.Labels).
		Description(&c.LoadBalancer.Annotations).
		String(constants.FieldLoadBalancerDescription, &c.LoadBalancer.Spec.Description, false)

	subresourceProcessors := []util.Processor{
		{
			Field:    constants.FieldLoadBalancerWorkloadType,
			Parser:   c.subresourceLoadBalancerWorkloadTypeParser,
			Required: false,
		},
		{
			Field:    constants.FieldLoadBalancerIPAM,
			Parser:   c.subresourceLoadBalancerIPAMParser,
			Required: false,
		},
		{
			Field:    constants.SubresourceTypeLoadBalancerListener,
			Parser:   c.subresourceLoadBalancerListenerParser,
			Required: true,
		},
		{
			Field:    constants.SubresourceTypeLoadBalancerHealthCheck,
			Parser:   c.subresourceLoadBalancerHealthCheckParser,
			Required: false,
		},
	}

	return append(processors, subresourceProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.LoadBalancer, nil
}

func newLoadBalancerConstructor(loadbalancer *loadbalancerv1.LoadBalancer) util.Constructor {
	return &Constructor{
		LoadBalancer: loadbalancer,
	}
}

func Creator(namespace, name string) util.Constructor {
	loadbalancer := &loadbalancerv1.LoadBalancer{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newLoadBalancerConstructor(loadbalancer)
}

func Updater(loadbalancer *loadbalancerv1.LoadBalancer) util.Constructor {
	return newLoadBalancerConstructor(loadbalancer)
}

func (c *Constructor) subresourceLoadBalancerWorkloadTypeParser(data interface{}) error {
	workloadType := data.(string)

	if workloadType != "vm" && workloadType != "cluster" {
		return fmt.Errorf("invalid value for workload type: %v", workloadType)
	}

	c.LoadBalancer.Spec.WorkloadType = loadbalancerv1.WorkloadType(workloadType)

	return nil
}

func (c *Constructor) subresourceLoadBalancerIPAMParser(data interface{}) error {
	ipam := data.(string)

	if ipam != "dhcp" && ipam != "cluster" {
		return fmt.Errorf("invalid value for IPAM: %v", ipam)
	}

	c.LoadBalancer.Spec.IPAM = loadbalancerv1.IPAM(ipam)
	return nil
}

func (c *Constructor) subresourceLoadBalancerListenerParser(data interface{}) error {
	listener := data.(map[string]interface{})

	name := listener[constants.FieldListenerName].(string)
	port := int32(listener[constants.FieldListenerPort].(int))
	protocol := corev1.Protocol(listener[constants.FieldListenerProtocol].(string))
	backendPort := int32(listener[constants.FieldListenerBackendPort].(int))

	c.LoadBalancer.Spec.Listeners = append(c.LoadBalancer.Spec.Listeners, loadbalancerv1.Listener{
		Name:        name,
		Port:        port,
		Protocol:    protocol,
		BackendPort: backendPort,
	})

	return nil
}

func (c *Constructor) subresourceLoadBalancerHealthCheckParser(data interface{}) error {
	return nil
}
