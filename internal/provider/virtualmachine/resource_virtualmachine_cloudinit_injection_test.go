package virtualmachine

import (
	"errors"
	"reflect"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

const (
	udSSHUser  = "sles"
	udKeyA     = "ssh-ed25519 AAAAkeyA a@example"
	udKeyB     = "ssh-ed25519 AAAAkeyB b@example"
	udPackages = "#cloud-config\npackages:\n  - vim\n"
	udOwnUser  = "#cloud-config\nuser: admin\n"

	udKeyNameA   = "default/key-a"
	udKeyNameB   = "default/key-b"
	udKeyNamesA  = `["default/key-a"]`
	udKeyNamesAB = `["default/key-a","default/key-b"]`
)

func TestInjectSSHUserAndKeys(t *testing.T) {
	testcases := []struct {
		name       string
		userData   string
		sshUser    string
		publicKeys []string
		expected   string
	}{
		{name: "nothing to inject", userData: udPackages, expected: udPackages},
		{name: "user into empty user data", sshUser: udSSHUser, expected: "#cloud-config\nuser: sles\n"},
		{name: "user appended", userData: udPackages, sshUser: udSSHUser, expected: udPackages + "\nuser: sles\n"},
		{name: "user appended without a trailing newline", userData: "#cloud-config\ntimezone: UTC", sshUser: udSSHUser, expected: "#cloud-config\ntimezone: UTC\nuser: sles\n"},
		{name: "user already set", userData: udOwnUser, sshUser: udSSHUser, expected: udOwnUser},
		{name: "keys into empty user data", publicKeys: []string{udKeyA, udKeyB}, expected: "#cloud-config\nssh_authorized_keys:\n  - " + udKeyA + "\n  - " + udKeyB},
		{name: "keys already set", userData: "#cloud-config\nssh_authorized_keys:\n  - " + udKeyB + "\n", publicKeys: []string{udKeyA}, expected: "#cloud-config\nssh_authorized_keys:\n  - " + udKeyB + "\n"},
		{name: "user and keys into empty user data", sshUser: udSSHUser, publicKeys: []string{udKeyA}, expected: "#cloud-config\nuser: sles\n\nssh_authorized_keys:\n  - " + udKeyA},
		{name: "user and keys appended", userData: udPackages, sshUser: udSSHUser, publicKeys: []string{udKeyA}, expected: udPackages + "\nuser: sles\n\nssh_authorized_keys:\n  - " + udKeyA},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := injectSSHUserAndKeys(tc.userData, tc.sshUser, tc.publicKeys); got != tc.expected {
				t.Errorf("injectSSHUserAndKeys() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestKeepConfiguredUserData(t *testing.T) {
	errLookup := errors.New("key pair not found")
	testcases := []struct {
		name         string
		configured   *string // nil: no cloudinit block in d, as after an import
		stored       map[string]interface{}
		sshUser      string
		sshNames     string // ssh names annotation
		keys         []string
		lookupErr    error
		expected     string
		wantLookupOf []string // nil: the key pairs must not be read
	}{
		{
			name:         "user line followed by injected keys",
			configured:   new(udPackages),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nuser: sles\n\nssh_authorized_keys:\n  - " + udKeyA + "\n  - " + udKeyB},
			sshUser:      udSSHUser,
			sshNames:     udKeyNamesAB,
			keys:         []string{udKeyA, udKeyB},
			expected:     udPackages,
			wantLookupOf: []string{udKeyNameA, udKeyNameB},
		},
		{
			name:         "user line only",
			configured:   new(udPackages),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nuser: sles\n"},
			sshUser:      udSSHUser,
			expected:     udPackages,
			wantLookupOf: []string{},
		},
		{
			name:         "keys injected into an empty user data",
			configured:   new(""),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: "#cloud-config\nssh_authorized_keys:\n  - " + udKeyA},
			sshNames:     udKeyNamesA,
			keys:         []string{udKeyA},
			expected:     "",
			wantLookupOf: []string{udKeyNameA},
		},
		{
			name:       "nothing injected",
			configured: new(udOwnUser),
			stored:     map[string]interface{}{constants.FieldCloudInitUserData: udOwnUser},
			sshUser:    udSSHUser,
			expected:   udOwnUser,
		},
		{
			name:         "change made outside Terraform",
			configured:   new(udPackages),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nuser: sles\nruncmd:\n  - reboot\n"},
			sshUser:      udSSHUser,
			expected:     udPackages + "\nuser: sles\nruncmd:\n  - reboot\n",
			wantLookupOf: []string{},
		},
		{
			name:         "configured user data changed",
			configured:   new("#cloud-config\npackages:\n  - git\n"),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nuser: sles\n"},
			sshUser:      udSSHUser,
			expected:     udPackages + "\nuser: sles\n",
			wantLookupOf: []string{},
		},
		{
			name:         "key pair changed since the VM was built",
			configured:   new(udPackages),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nssh_authorized_keys:\n  - " + udKeyA},
			sshNames:     udKeyNamesA,
			keys:         []string{udKeyB},
			expected:     udPackages + "\nssh_authorized_keys:\n  - " + udKeyA,
			wantLookupOf: []string{udKeyNameA},
		},
		{
			name:         "key pair cannot be read",
			configured:   new(udPackages),
			stored:       map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nssh_authorized_keys:\n  - " + udKeyA},
			sshNames:     udKeyNamesA,
			keys:         []string{udKeyA},
			lookupErr:    errLookup,
			expected:     udPackages + "\nssh_authorized_keys:\n  - " + udKeyA,
			wantLookupOf: []string{udKeyNameA},
		},
		{
			name:       "malformed ssh names annotation",
			configured: new(udPackages),
			stored:     map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nuser: sles\n"},
			sshUser:    udSSHUser,
			sshNames:   `default/key-a`,
			expected:   udPackages + "\nuser: sles\n",
		},
		{
			name:     "no cloudinit block in Terraform",
			stored:   map[string]interface{}{constants.FieldCloudInitUserData: udPackages + "\nuser: sles\n"},
			sshUser:  udSSHUser,
			expected: udPackages + "\nuser: sles\n",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := map[string]interface{}{}
			if tc.configured != nil {
				raw[constants.FieldVirtualMachineCloudInit] = []interface{}{
					map[string]interface{}{constants.FieldCloudInitUserData: *tc.configured},
				}
			}
			d := schema.TestResourceDataRaw(t, ResourceVirtualMachine().Schema, raw)

			vm := &kubevirtv1.VirtualMachine{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
				Spec: kubevirtv1.VirtualMachineSpec{
					Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
					},
				},
			}
			if tc.sshUser != "" {
				vm.Labels[builder.LabelPrefixHarvesterTag+constants.LabelSSHUsername] = tc.sshUser
			}
			if tc.sshNames != "" {
				vm.Spec.Template.ObjectMeta.Annotations[builder.AnnotationKeyVirtualMachineSSHNames] = tc.sshNames
			}

			getter := &importer.StateGetter{States: map[string]interface{}{
				constants.FieldVirtualMachineCloudInit: []map[string]interface{}{tc.stored},
			}}

			var lookupOf []string
			lookedUp := false
			keepConfiguredUserData(d, vm, getter, func(sshNames []string) ([]string, error) {
				lookedUp = true
				lookupOf = append([]string{}, sshNames...)
				return tc.keys, tc.lookupErr
			})

			got := getter.States[constants.FieldVirtualMachineCloudInit].([]map[string]interface{})[0][constants.FieldCloudInitUserData]
			if got != tc.expected {
				t.Errorf("user_data = %q, want %q", got, tc.expected)
			}
			if tc.wantLookupOf == nil {
				if lookedUp {
					t.Errorf("key pairs read with %v, want no read", lookupOf)
				}
			} else if !reflect.DeepEqual(lookupOf, tc.wantLookupOf) {
				t.Errorf("key pairs read with %v, want %v", lookupOf, tc.wantLookupOf)
			}
		})
	}
}
