package virtualmachine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

// injectSSHUserAndKeys returns the user data the constructor stores on the VM:
// a `user:` line from the ssh-user tag when the user data sets no user, then
// the public keys of ssh_keys when it has no ssh_authorized_keys section.
func injectSSHUserAndKeys(userData, sshUser string, publicKeys []string) string {
	if sshUser != "" {
		if userData == "" {
			userData = fmt.Sprintf("#cloud-config\nuser: %s\n", sshUser)
		} else if !hasLinePrefix(userData, "user: ") {
			userData += fmt.Sprintf("\nuser: %s\n", sshUser)
		}
	}
	if len(publicKeys) > 0 && !hasLinePrefix(userData, "ssh_authorized_keys:") {
		keys := strings.Join(publicKeys, "\n  - ")
		if userData == "" {
			userData = fmt.Sprintf("#cloud-config\nssh_authorized_keys:\n  - %s", keys)
		} else {
			userData += fmt.Sprintf("\nssh_authorized_keys:\n  - %s", keys)
		}
	}
	return userData
}

func hasLinePrefix(text, prefix string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// keyPairPublicKeys returns the public keys of the given ssh_keys, in order.
func keyPairPublicKeys(ctx context.Context, c *client.Client, namespace string, sshNames []string) ([]string, error) {
	publicKeys := make([]string, 0, len(sshNames))
	for _, sshName := range sshNames {
		_, keyPairName, err := helper.NamespacedNameParts(sshName)
		if err != nil {
			return nil, err
		}
		keyPair, err := c.HarvesterClient.HarvesterhciV1beta1().KeyPairs(namespace).Get(ctx, keyPairName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, keyPair.Spec.PublicKey)
	}
	return publicKeys, nil
}

// keepConfiguredUserData sets the user data read from the VM back to the
// configured one when they only differ by what the constructor injected.
// Any other difference is reported, so changes made outside Terraform still
// show up in the plan. When the injection cannot be recomputed (for example a
// key pair that cannot be read), the stored user data is reported.
func keepConfiguredUserData(d *schema.ResourceData, vm *kubevirtv1.VirtualMachine, getter *importer.StateGetter, publicKeys func(sshNames []string) ([]string, error)) {
	configured, ok := configuredUserData(d)
	if !ok {
		return
	}
	cloudInit, ok := getter.States[constants.FieldVirtualMachineCloudInit].([]map[string]interface{})
	if !ok || len(cloudInit) != 1 {
		return
	}
	state := cloudInit[0]
	stored, _ := state[constants.FieldCloudInitUserData].(string)
	if stored == configured {
		return
	}
	var sshNames []string
	if annotation := vm.Spec.Template.ObjectMeta.Annotations[builder.AnnotationKeyVirtualMachineSSHNames]; annotation != "" {
		if err := json.Unmarshal([]byte(annotation), &sshNames); err != nil {
			return
		}
	}
	keys, err := publicKeys(sshNames)
	if err != nil {
		return
	}
	sshUser := vm.Labels[builder.LabelPrefixHarvesterTag+constants.LabelSSHUsername]
	if injectSSHUserAndKeys(configured, sshUser, keys) == stored {
		state[constants.FieldCloudInitUserData] = configured
	}
}

// configuredUserData returns the user_data held by d: the configuration during
// create and update, the previous state during refresh.
func configuredUserData(d *schema.ResourceData) (string, bool) {
	cloudInit, ok := d.Get(constants.FieldVirtualMachineCloudInit).([]interface{})
	if !ok || len(cloudInit) != 1 || cloudInit[0] == nil {
		return "", false
	}
	userData, ok := cloudInit[0].(map[string]interface{})[constants.FieldCloudInitUserData].(string)
	return userData, ok
}
