package keypair

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	KeyPair *harvsterv1.KeyPair
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.KeyPair.Labels).
		Labels(&c.KeyPair.Labels).
		Description(&c.KeyPair.Annotations).
		String(constants.FieldKeyPairPublicKey, &c.KeyPair.Spec.PublicKey, true)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.KeyPair, nil
}

func newKeyPairConstructor(keyPair *harvsterv1.KeyPair) util.Constructor {
	return &Constructor{
		KeyPair: keyPair,
	}
}

func Creator(namespace, name string) util.Constructor {
	keyPair := &harvsterv1.KeyPair{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newKeyPairConstructor(keyPair)
}

func Updater(keyPair *harvsterv1.KeyPair) util.Constructor {
	return newKeyPairConstructor(keyPair)
}
