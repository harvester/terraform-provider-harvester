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

const (
	// sshKeysLinePrefix starts the ssh_authorized_keys section of user data.
	sshKeysLinePrefix = "ssh_authorized_keys:"
	// sshKeySeparator separates the public keys in the ssh_authorized_keys
	// section the constructor injects.
	sshKeySeparator = "\n  - "
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
	if len(publicKeys) > 0 && !hasLinePrefix(userData, sshKeysLinePrefix) {
		keys := strings.Join(publicKeys, sshKeySeparator)
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

// listKeyPairPublicKeys returns the public keys of the key pairs in namespace
// by name, read with a single List call.
func listKeyPairPublicKeys(ctx context.Context, c *client.Client, namespace string) (map[string]string, error) {
	keyPairs, err := c.HarvesterClient.HarvesterhciV1beta1().KeyPairs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	publicKeys := make(map[string]string, len(keyPairs.Items))
	for _, keyPair := range keyPairs.Items {
		publicKeys[keyPair.Name] = keyPair.Spec.PublicKey
	}
	return publicKeys, nil
}

// injectedKeys returns the public keys the constructor appended to the
// configured user data, taken from the stored user data itself. ok is true
// only when stored is exactly injectSSHUserAndKeys(configured, sshUser, keys)
// with keyCount keys, or, when the constructor injects no keys (no ssh names,
// or an ssh_authorized_keys section of its own), exactly configured with the
// user injected.
func injectedKeys(configured, stored, sshUser string, keyCount int) (keys []string, ok bool) {
	withUser := injectSSHUserAndKeys(configured, sshUser, nil)
	if keyCount == 0 || hasLinePrefix(withUser, sshKeysLinePrefix) {
		return nil, stored == withUser
	}
	// injecting a single empty key gives the text that precedes the keys
	rest, ok := strings.CutPrefix(stored, injectSSHUserAndKeys(withUser, "", []string{""}))
	if !ok {
		return nil, false
	}
	keys = strings.Split(rest, sshKeySeparator)
	if len(keys) != keyCount {
		return nil, false
	}
	return keys, true
}

// sameKeys reports whether keys are the public keys of the key pairs named by
// sshNames, in the same order.
func sameKeys(keys, sshNames []string, keyPairs func() (map[string]string, error)) bool {
	publicKeys, err := keyPairs()
	if err != nil {
		return false
	}
	for i, sshName := range sshNames {
		_, keyPairName, err := helper.NamespacedNameParts(sshName)
		if err != nil {
			return false
		}
		if publicKey, found := publicKeys[keyPairName]; !found || publicKey != keys[i] {
			return false
		}
	}
	return true
}

// keepConfiguredUserData sets the user data read from the VM back to the
// configured one when the VM holds exactly what the constructor builds from
// it, the ssh-user tag and as many keys as the ssh names annotation lists.
// Any other difference is reported, so changes made outside Terraform still
// show up in the plan.
//
// This structural check reads nothing but the VM. It cannot tell whether the
// injected keys are still those of the key pairs: when keyPairs is not nil,
// they are compared with the public keys it returns, so a key edited on the
// VM or changed in its key pair is reported, and so is the stored user data
// when keyPairs fails. Read passes it; the polling that follows a create or
// update does not, as the constructor has just injected the keys and the next
// read compares them.
func keepConfiguredUserData(d *schema.ResourceData, vm *kubevirtv1.VirtualMachine, getter *importer.StateGetter, keyPairs func() (map[string]string, error)) {
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
	sshUser := vm.Labels[builder.LabelPrefixHarvesterTag+constants.LabelSSHUsername]
	keys, ok := injectedKeys(configured, stored, sshUser, len(sshNames))
	if !ok {
		return
	}
	if len(keys) > 0 && keyPairs != nil && !sameKeys(keys, sshNames, keyPairs) {
		return
	}
	state[constants.FieldCloudInitUserData] = configured
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
