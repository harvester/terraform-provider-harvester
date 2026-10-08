package virtualmachine

import (
	"context"
	"reflect"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

const (
	accessCredentialTestKeys      = "ssh-keys"
	accessCredentialTestPasswords = "passwords"
	accessCredentialTestUser      = "root"
)

// sshPublicKeyBlock returns an ssh_public_key block as Terraform passes it:
// the users list is always set, empty when not configured.
func sshPublicKeyBlock(secretName, method string, users ...string) map[string]interface{} {
	userList := make([]interface{}, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}
	return map[string]interface{}{
		constants.FieldAccessCredentialSecretName:        secretName,
		constants.FieldAccessCredentialPropagationMethod: method,
		constants.FieldAccessCredentialUsers:             userList,
	}
}

// accessCredentialBlock returns an access_credentials block with the given
// ssh_public_key and user_password blocks.
func accessCredentialBlock(sshPublicKey, userPassword []interface{}) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldAccessCredentialSSHPublicKey: sshPublicKey,
		constants.FieldAccessCredentialUserPassword: userPassword,
	}
}

func userPasswordBlock(secretName string) []interface{} {
	return []interface{}{map[string]interface{}{constants.FieldAccessCredentialSecretName: secretName}}
}

func sshPublicKeyCredential(secretName string, method kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod) *kubevirtv1.SSHPublicKeyAccessCredential {
	return &kubevirtv1.SSHPublicKeyAccessCredential{
		Source: kubevirtv1.SSHPublicKeyAccessCredentialSource{
			Secret: &kubevirtv1.AccessCredentialSecretSource{SecretName: secretName},
		},
		PropagationMethod: method,
	}
}

func userPasswordCredential(secretName string) *kubevirtv1.UserPasswordAccessCredential {
	return &kubevirtv1.UserPasswordAccessCredential{
		Source: kubevirtv1.UserPasswordAccessCredentialSource{
			Secret: &kubevirtv1.AccessCredentialSecretSource{SecretName: secretName},
		},
		PropagationMethod: kubevirtv1.UserPasswordAccessCredentialPropagationMethod{
			QemuGuestAgent: &kubevirtv1.QemuGuestAgentUserPasswordAccessCredentialPropagation{},
		},
	}
}

// TestParseAccessCredential also covers the blocks the KubeVirt webhook
// rejects (no type, both types, qemuGuestAgent without users): they must reach
// it unchanged, not be completed or silently reduced to one type.
func TestParseAccessCredential(t *testing.T) {
	testcases := []struct {
		name     string
		block    map[string]interface{}
		expected kubevirtv1.AccessCredential
	}{
		{
			name:     "empty block (nil from Terraform) sets no type",
			block:    nil,
			expected: kubevirtv1.AccessCredential{},
		},
		{
			name:     "empty ssh_public_key and user_password lists set no type",
			block:    accessCredentialBlock([]interface{}{}, []interface{}{}),
			expected: kubevirtv1.AccessCredential{},
		},
		{
			name:  "ssh public key over noCloud",
			block: accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeNoCloud)}, []interface{}{}),
			expected: kubevirtv1.AccessCredential{SSHPublicKey: sshPublicKeyCredential(accessCredentialTestKeys, kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
				NoCloud: &kubevirtv1.NoCloudSSHPublicKeyAccessCredentialPropagation{},
			})},
		},
		{
			name:  "ssh public key over configDrive",
			block: accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeConfigDrive)}, []interface{}{}),
			expected: kubevirtv1.AccessCredential{SSHPublicKey: sshPublicKeyCredential(accessCredentialTestKeys, kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
				ConfigDrive: &kubevirtv1.ConfigDriveSSHPublicKeyAccessCredentialPropagation{},
			})},
		},
		{
			name:  "ssh public key over qemuGuestAgent keeps the users in order",
			block: accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, constants.AccessCredentialPropagationQemuGuestAgent, accessCredentialTestUser, "admin", "opensuse")}, []interface{}{}),
			expected: kubevirtv1.AccessCredential{SSHPublicKey: sshPublicKeyCredential(accessCredentialTestKeys, kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
				QemuGuestAgent: &kubevirtv1.QemuGuestAgentSSHPublicKeyAccessCredentialPropagation{Users: []string{accessCredentialTestUser, "admin", "opensuse"}},
			})},
		},
		{
			name:  "qemuGuestAgent without users is not completed",
			block: accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, constants.AccessCredentialPropagationQemuGuestAgent)}, []interface{}{}),
			expected: kubevirtv1.AccessCredential{SSHPublicKey: sshPublicKeyCredential(accessCredentialTestKeys, kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
				QemuGuestAgent: &kubevirtv1.QemuGuestAgentSSHPublicKeyAccessCredentialPropagation{Users: []string{}},
			})},
		},
		{
			name:     "user password over qemuGuestAgent",
			block:    accessCredentialBlock([]interface{}{}, userPasswordBlock(accessCredentialTestPasswords)),
			expected: kubevirtv1.AccessCredential{UserPassword: userPasswordCredential(accessCredentialTestPasswords)},
		},
		{
			name:  "both types are kept",
			block: accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeNoCloud)}, userPasswordBlock(accessCredentialTestPasswords)),
			expected: kubevirtv1.AccessCredential{
				SSHPublicKey: sshPublicKeyCredential(accessCredentialTestKeys, kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
					NoCloud: &kubevirtv1.NoCloudSSHPublicKeyAccessCredentialPropagation{},
				}),
				UserPassword: userPasswordCredential(accessCredentialTestPasswords),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseAccessCredential(tc.block); !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("parseAccessCredential() = %+v, want %+v", got, tc.expected)
			}
		})
	}
}

// TestAccessCredentialsSchema covers the invalid blocks that the schema
// rejects at plan time, which parseAccessCredential relies on.
func TestAccessCredentialsSchema(t *testing.T) {
	sshPublicKey := func(blocks ...interface{}) map[string]interface{} {
		return map[string]interface{}{constants.FieldAccessCredentialSSHPublicKey: blocks}
	}
	userPassword := func(blocks ...interface{}) map[string]interface{} {
		return map[string]interface{}{constants.FieldAccessCredentialUserPassword: blocks}
	}
	testcases := []struct {
		name        string
		block       map[string]interface{}
		expectError bool
	}{
		{
			name:  "valid ssh public key",
			block: sshPublicKey(sshPublicKeyBlock(accessCredentialTestKeys, constants.AccessCredentialPropagationQemuGuestAgent, accessCredentialTestUser)),
		},
		{
			name:  "valid user password",
			block: userPassword(userPasswordBlock(accessCredentialTestPasswords)...),
		},
		{
			name:        "unknown propagation method",
			block:       sshPublicKey(sshPublicKeyBlock(accessCredentialTestKeys, "cloudInit")),
			expectError: true,
		},
		{
			name: "ssh public key without secret name",
			block: sshPublicKey(map[string]interface{}{
				constants.FieldAccessCredentialPropagationMethod: builder.CloudInitTypeNoCloud,
			}),
			expectError: true,
		},
		{
			name: "ssh public key without propagation method",
			block: sshPublicKey(map[string]interface{}{
				constants.FieldAccessCredentialSecretName: accessCredentialTestKeys,
			}),
			expectError: true,
		},
		{
			name:        "user password without secret name",
			block:       userPassword(map[string]interface{}{}),
			expectError: true,
		},
		{
			name: "two ssh public keys in one block",
			block: sshPublicKey(
				sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeNoCloud),
				sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeConfigDrive),
			),
			expectError: true,
		},
		{
			name:        "two user passwords in one block",
			block:       userPassword(append(userPasswordBlock(accessCredentialTestPasswords), userPasswordBlock(accessCredentialTestPasswords)...)...),
			expectError: true,
		},
	}

	resource := &schema.Resource{Schema: map[string]*schema.Schema{
		constants.FieldVirtualMachineAccessCredentials: Schema()[constants.FieldVirtualMachineAccessCredentials],
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			diags := resource.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
				constants.FieldVirtualMachineAccessCredentials: []interface{}{tc.block},
			}))
			if diags.HasError() != tc.expectError {
				t.Errorf("Validate() = %v, expectError = %v", diags, tc.expectError)
			}
		})
	}
}

// TestAccessCredentialsParser checks that each access_credentials block adds
// one credential, in order, and that an empty block does not crash.
func TestAccessCredentialsParser(t *testing.T) {
	testcases := []struct {
		name     string
		blocks   []interface{}
		expected []kubevirtv1.AccessCredential
	}{
		{
			name:     "empty block (nil from Terraform) adds a credential without type",
			blocks:   []interface{}{nil},
			expected: []kubevirtv1.AccessCredential{{}},
		},
		{
			name:   "one block adds one credential",
			blocks: []interface{}{accessCredentialBlock([]interface{}{}, userPasswordBlock(accessCredentialTestPasswords))},
			expected: []kubevirtv1.AccessCredential{
				{UserPassword: userPasswordCredential(accessCredentialTestPasswords)},
			},
		},
		{
			name: "several blocks add their credentials in order",
			blocks: []interface{}{
				accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeNoCloud)}, []interface{}{}),
				accessCredentialBlock([]interface{}{}, userPasswordBlock(accessCredentialTestPasswords)),
				accessCredentialBlock([]interface{}{sshPublicKeyBlock("other-keys", builder.CloudInitTypeConfigDrive)}, []interface{}{}),
			},
			expected: []kubevirtv1.AccessCredential{
				{SSHPublicKey: sshPublicKeyCredential(accessCredentialTestKeys, kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
					NoCloud: &kubevirtv1.NoCloudSSHPublicKeyAccessCredentialPropagation{},
				})},
				{UserPassword: userPasswordCredential(accessCredentialTestPasswords)},
				{SSHPublicKey: sshPublicKeyCredential("other-keys", kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
					ConfigDrive: &kubevirtv1.ConfigDriveSSHPublicKeyAccessCredentialPropagation{},
				})},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			c := Creator(nil, ctx, "default", "access-credentials").(*Constructor)
			for _, processor := range c.Setup() {
				if processor.Field != constants.FieldVirtualMachineAccessCredentials {
					continue
				}
				for _, block := range tc.blocks {
					if err := processor.Parser(block); err != nil {
						t.Fatal(err)
					}
				}
			}
			if got := c.Builder.VirtualMachine.Spec.Template.Spec.AccessCredentials; !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("access credentials = %+v, want %+v", got, tc.expected)
			}
		})
	}
}

// TestUpdaterResetsAccessCredentials checks that an update starts from no
// credential, so that a removed access_credentials block is removed from the VM.
func TestUpdaterResetsAccessCredentials(t *testing.T) {
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
		Spec: kubevirtv1.VirtualMachineSpec{Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
			Spec: kubevirtv1.VirtualMachineInstanceSpec{AccessCredentials: []kubevirtv1.AccessCredential{
				{UserPassword: userPasswordCredential(accessCredentialTestPasswords)},
			}},
		}},
	}
	updater := Updater(nil, context.Background(), vm).(*Constructor)
	if got := updater.Builder.VirtualMachine.Spec.Template.Spec.AccessCredentials; got != nil {
		t.Errorf("access credentials after Updater() = %+v, want nil", got)
	}
}

// TestAccessCredentialsRoundTrip builds a VM from several access_credentials
// blocks, reads it back with the importer and checks that the state equals the
// configuration, so that no block is lost, reordered or changed on the way.
func TestAccessCredentialsRoundTrip(t *testing.T) {
	raw := map[string]interface{}{
		constants.FieldCommonName:      "round-trip",
		constants.FieldCommonNamespace: "default",
		constants.FieldVirtualMachineAccessCredentials: []interface{}{
			accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, constants.AccessCredentialPropagationQemuGuestAgent, accessCredentialTestUser, "admin")}, []interface{}{}),
			accessCredentialBlock([]interface{}{}, userPasswordBlock(accessCredentialTestPasswords)),
			accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeNoCloud)}, []interface{}{}),
			accessCredentialBlock([]interface{}{sshPublicKeyBlock(accessCredentialTestKeys, builder.CloudInitTypeConfigDrive)}, []interface{}{}),
		},
	}
	ctx := context.Background()
	config := schema.TestResourceDataRaw(t, Schema(), raw)
	obj, err := util.ResourceConstruct(ctx, config, Creator(nil, ctx, "default", "round-trip"))
	if err != nil {
		t.Fatal(err)
	}
	state := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{})
	exported := importer.NewVMImporter(obj.(*kubevirtv1.VirtualMachine), nil).AccessCredentials()
	if err := state.Set(constants.FieldVirtualMachineAccessCredentials, exported); err != nil {
		t.Fatal(err)
	}
	field := constants.FieldVirtualMachineAccessCredentials
	if !reflect.DeepEqual(state.Get(field), config.Get(field)) {
		t.Errorf("%s read back as %v, want %v", field, state.Get(field), config.Get(field))
	}
}
