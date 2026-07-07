package namespace

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Namespace *corev1.Namespace
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.Namespace.Labels).
		Labels(&c.Namespace.Labels).
		Description(&c.Namespace.Annotations)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Namespace, nil
}

func newNamespaceConstructor(ns *corev1.Namespace) util.Constructor {
	return &Constructor{
		Namespace: ns,
	}
}

func Creator(name string) util.Constructor {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}
	return newNamespaceConstructor(ns)
}

func Updater(ns *corev1.Namespace) util.Constructor {
	return newNamespaceConstructor(ns)
}
