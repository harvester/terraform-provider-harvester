package builder

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

type ServiceBuilder struct {
	vm       *kubevirtv1.VirtualMachine
	services map[string]*corev1.Service
}

func NewServiceBuilder(vm *kubevirtv1.VirtualMachine) *ServiceBuilder {
	return &ServiceBuilder{
		vm:       vm,
		services: make(map[string]*corev1.Service),
	}
}

func (s *ServiceBuilder) Expose(name string, serviceType corev1.ServiceType, ports ...int32) *ServiceBuilder {
	vm := s.vm
	objectMeta := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", vm.Name, name),
		Namespace: vm.Namespace,
		Labels:    vm.Spec.Template.ObjectMeta.Labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: vm.APIVersion,
				Kind:       vm.Kind,
				Name:       vm.Name,
				UID:        vm.UID,
			},
		},
	}
	servicePorts := make([]corev1.ServicePort, 0, len(ports))
	for _, port := range ports {
		servicePort := corev1.ServicePort{
			Name: fmt.Sprintf("%s-%d", name, port),
			Port: port,
			TargetPort: intstr.IntOrString{
				IntVal: port,
			},
		}
		servicePorts = append(servicePorts, servicePort)
	}
	svc := &corev1.Service{
		ObjectMeta: objectMeta,
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Ports:    servicePorts,
			Selector: vm.Spec.Template.ObjectMeta.Labels,
		},
	}
	s.services[name] = svc
	return s
}

func (s *ServiceBuilder) Services() map[string]*corev1.Service {
	return s.services
}
