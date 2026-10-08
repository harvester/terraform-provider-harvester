package virtualmachine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

const (
	udSSHUser  = "sles"
	udKeyA     = "ssh-ed25519 AAAAkeyA a@example"
	udKeyB     = "ssh-ed25519 AAAAkeyB b@example"
	udKeyX     = "ssh-ed25519 AAAAkeyX x@example"
	udPackages = "#cloud-config\npackages:\n  - vim\n"
	udOwnUser  = "#cloud-config\nuser: admin\n"
	udOwnKeys  = "#cloud-config\nssh_authorized_keys:\n  - " + udKeyX + "\n"

	udKeyPairA   = "key-a"
	udKeyPairB   = "key-b"
	udKeyNamesA  = `["default/key-a"]`
	udKeyNamesAB = `["default/key-a","default/key-b"]`

	// what the constructor appends to udPackages with the ssh-user tag
	udUserLine = "\nuser: sles\n"
	udKeysHead = "\nssh_authorized_keys:\n  - "
)

var udKeyPairsAB = map[string]string{udKeyPairA: udKeyA, udKeyPairB: udKeyB}

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

// TestKeepConfiguredUserData runs every case as Read does, with the key pairs
// listed once when the keys need comparing, and as the polling after a create
// or update does, without reading the key pairs.
func TestKeepConfiguredUserData(t *testing.T) {
	testcases := []struct {
		name          string
		configured    *string // nil: no cloudinit block in d, as after an import
		stored        string
		sshUser       string
		sshNames      string            // ssh names annotation
		keyPairs      map[string]string // public keys of the key pairs by name
		listErr       error
		keep          bool // Read keeps the configured user data
		keepOnPolling bool // the polling keeps the configured user data
		wantList      bool // Read lists the key pairs
	}{
		{
			name:          "user line followed by injected keys",
			configured:    new(udPackages),
			stored:        udPackages + udUserLine + udKeysHead + udKeyA + "\n  - " + udKeyB,
			sshUser:       udSSHUser,
			sshNames:      udKeyNamesAB,
			keyPairs:      udKeyPairsAB,
			keep:          true,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "user line only",
			configured:    new(udPackages),
			stored:        udPackages + udUserLine,
			sshUser:       udSSHUser,
			keep:          true,
			keepOnPolling: true,
		},
		{
			name:          "keys injected into an empty user data",
			configured:    new(""),
			stored:        "#cloud-config\nssh_authorized_keys:\n  - " + udKeyA,
			sshNames:      udKeyNamesA,
			keyPairs:      udKeyPairsAB,
			keep:          true,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "user and keys injected into an empty user data",
			configured:    new(""),
			stored:        "#cloud-config\nuser: sles\n" + udKeysHead + udKeyA,
			sshUser:       udSSHUser,
			sshNames:      udKeyNamesA,
			keyPairs:      udKeyPairsAB,
			keep:          true,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "keys injected after a user line of its own",
			configured:    new(udOwnUser),
			stored:        udOwnUser + udKeysHead + udKeyA,
			sshUser:       udSSHUser,
			sshNames:      udKeyNamesA,
			keyPairs:      udKeyPairsAB,
			keep:          true,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "ssh_authorized_keys of its own, only the user injected",
			configured:    new(udOwnKeys),
			stored:        udOwnKeys + udUserLine,
			sshUser:       udSSHUser,
			sshNames:      udKeyNamesAB,
			keep:          true,
			keepOnPolling: true,
		},
		{
			name:          "nothing injected",
			configured:    new(udOwnUser),
			stored:        udOwnUser,
			sshUser:       udSSHUser,
			keep:          true,
			keepOnPolling: true,
		},
		{
			name:       "change made outside Terraform",
			configured: new(udPackages),
			stored:     udPackages + "\nuser: sles\nruncmd:\n  - reboot\n",
			sshUser:    udSSHUser,
		},
		{
			name:       "configured user data changed",
			configured: new("#cloud-config\npackages:\n  - git\n"),
			stored:     udPackages + udUserLine,
			sshUser:    udSSHUser,
		},
		{
			name:       "user line removed on the VM",
			configured: new(udPackages),
			stored:     udPackages + udKeysHead + udKeyA,
			sshUser:    udSSHUser,
			sshNames:   udKeyNamesA,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "user line edited on the VM",
			configured: new(udPackages),
			stored:     udPackages + "\nuser: admin\n" + udKeysHead + udKeyA,
			sshUser:    udSSHUser,
			sshNames:   udKeyNamesA,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "ssh-user tag removed from the VM",
			configured: new(udPackages),
			stored:     udPackages + udUserLine,
		},
		{
			name:       "keys injected but no ssh names on the VM",
			configured: new(udPackages),
			stored:     udPackages + udKeysHead + udKeyA,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "user data replaced on the VM",
			configured: new("#cloud-config\ntimezone: UTC\n"),
			stored:     "#cloud-config\nssh_authorized_keys:\n  - " + udKeyA,
			sshNames:   udKeyNamesAB,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "key removed on the VM",
			configured: new(udPackages),
			stored:     udPackages + udKeysHead + udKeyA,
			sshNames:   udKeyNamesAB,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "key added on the VM",
			configured: new(udPackages),
			stored:     udPackages + udKeysHead + udKeyA + "\n  - " + udKeyB + "\n  - " + udKeyX,
			sshNames:   udKeyNamesAB,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "section added after the keys",
			configured: new(udPackages),
			stored:     udPackages + udKeysHead + udKeyA + "\n  - " + udKeyB + "\nruncmd:\n  - reboot",
			sshNames:   udKeyNamesAB,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "keys indented differently",
			configured: new(udPackages),
			stored:     udPackages + "\nssh_authorized_keys:\n    - " + udKeyA,
			sshNames:   udKeyNamesA,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "keys separated differently",
			configured: new(udPackages),
			stored:     udPackages + udKeysHead + udKeyA + "\n- " + udKeyB,
			sshNames:   udKeyNamesAB,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:       "blank line before the keys removed",
			configured: new(udPackages),
			stored:     udPackages + udUserLine + "ssh_authorized_keys:\n  - " + udKeyA,
			sshUser:    udSSHUser,
			sshNames:   udKeyNamesA,
			keyPairs:   udKeyPairsAB,
		},
		{
			name:          "text appended to the last key",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyA + "\n  - " + udKeyB + "\nruncmd: [reboot]",
			sshNames:      udKeyNamesAB,
			keyPairs:      udKeyPairsAB,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "key edited on the VM",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyA + "\n  - " + udKeyX,
			sshNames:      udKeyNamesAB,
			keyPairs:      udKeyPairsAB,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "keys swapped on the VM",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyB + "\n  - " + udKeyA,
			sshNames:      udKeyNamesAB,
			keyPairs:      udKeyPairsAB,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "key pair changed since the VM was built",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyA,
			sshNames:      udKeyNamesA,
			keyPairs:      map[string]string{udKeyPairA: udKeyB},
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "key pair deleted",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyA,
			sshNames:      udKeyNamesA,
			keyPairs:      map[string]string{},
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "key emptied on the VM and key pair deleted",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead,
			sshNames:      udKeyNamesA,
			keyPairs:      map[string]string{},
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "key pairs cannot be listed",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyA,
			sshNames:      udKeyNamesA,
			keyPairs:      udKeyPairsAB,
			listErr:       errors.New("forbidden"),
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:          "malformed ssh name",
			configured:    new(udPackages),
			stored:        udPackages + udKeysHead + udKeyA,
			sshNames:      `["default/key-a/extra"]`,
			keyPairs:      udKeyPairsAB,
			keepOnPolling: true,
			wantList:      true,
		},
		{
			name:       "malformed ssh names annotation",
			configured: new(udPackages),
			stored:     udPackages + udUserLine,
			sshUser:    udSSHUser,
			sshNames:   `default/key-a`,
		},
		{
			name:    "no cloudinit block in Terraform",
			stored:  udPackages + udUserLine,
			sshUser: udSSHUser,
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
			run := func(keyPairs func() (map[string]string, error)) string {
				d := schema.TestResourceDataRaw(t, ResourceVirtualMachine().Schema, raw)
				getter := &importer.StateGetter{States: map[string]interface{}{
					constants.FieldVirtualMachineCloudInit: []map[string]interface{}{
						{constants.FieldCloudInitUserData: tc.stored},
					},
				}}
				keepConfiguredUserData(d, vm, getter, keyPairs)
				return getter.States[constants.FieldVirtualMachineCloudInit].([]map[string]interface{})[0][constants.FieldCloudInitUserData].(string)
			}
			expected := func(keep bool) string {
				if keep {
					return *tc.configured
				}
				return tc.stored
			}

			lists := 0
			got := run(func() (map[string]string, error) {
				lists++
				return tc.keyPairs, tc.listErr
			})
			if got != expected(tc.keep) {
				t.Errorf("read: user_data = %q, want %q", got, expected(tc.keep))
			}
			if wantLists := map[bool]int{true: 1}[tc.wantList]; lists != wantLists {
				t.Errorf("read: key pairs listed %d times, want %d", lists, wantLists)
			}
			if got := run(nil); got != expected(tc.keepOnPolling) {
				t.Errorf("polling: user_data = %q, want %q", got, expected(tc.keepOnPolling))
			}
			// what Read keeps is exactly what the constructor builds
			if tc.keep {
				if injected := injectSSHUserAndKeys(*tc.configured, tc.sshUser, testPublicKeys(t, tc.sshNames, tc.keyPairs)); injected != tc.stored {
					t.Errorf("kept user data the constructor builds as %q, stored %q", injected, tc.stored)
				}
			}
		})
	}
}

// testPublicKeys returns the public keys the constructor injects for the ssh
// names annotation.
func testPublicKeys(t *testing.T, sshNames string, keyPairs map[string]string) []string {
	if sshNames == "" {
		return nil
	}
	var names []string
	if err := json.Unmarshal([]byte(sshNames), &names); err != nil {
		t.Fatal(err)
	}
	publicKeys := make([]string, 0, len(names))
	for _, name := range names {
		_, keyPairName, err := helper.NamespacedNameParts(name)
		if err != nil {
			t.Fatal(err)
		}
		publicKeys = append(publicKeys, keyPairs[keyPairName])
	}
	return publicKeys
}

// TestVirtualMachineKeyPairRequests checks against a fake API server which
// key pair requests Read and the polling send for a VM with two injected
// keys: one List for Read, none for the polling.
func TestVirtualMachineKeyPairRequests(t *testing.T) {
	const (
		namespace   = "default"
		name        = "test-vm"
		keyPairPath = "/apis/harvesterhci.io/v1beta1/namespaces/default/keypairs"
	)
	stored := udPackages + udUserLine + udKeysHead + udKeyA + "\n  - " + udKeyB

	vm, err := builder.NewVMBuilder("test").Namespace(namespace).Name(name).
		CPU(1).Memory("1Gi").MachineType("q35").EvictionStrategy(true).RunStrategy(kubevirtv1.RunStrategyHalted).
		Labels(map[string]string{builder.LabelPrefixHarvesterTag + constants.LabelSSHUsername: udSSHUser}).
		SSHKey("default/key-a").SSHKey("default/key-b").
		CloudInitDisk(builder.CloudInitDiskName, builder.DiskBusVirtio, false, 0, builder.CloudInitSource{
			CloudInitType: builder.CloudInitTypeNoCloud,
			UserData:      stored,
		}).VM()
	if err != nil {
		t.Fatal(err)
	}
	vm.TypeMeta = metav1.TypeMeta{APIVersion: "kubevirt.io/v1", Kind: "VirtualMachine"}
	keyPairs := &harvsterv1.KeyPairList{
		TypeMeta: metav1.TypeMeta{APIVersion: "harvesterhci.io/v1beta1", Kind: "KeyPairList"},
		Items: []harvsterv1.KeyPair{
			{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: udKeyPairA}, Spec: harvsterv1.KeyPairSpec{PublicKey: udKeyA}},
			{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: udKeyPairB}, Spec: harvsterv1.KeyPairSpec{PublicKey: udKeyB}},
		},
	}

	var (
		mu       sync.Mutex
		requests []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apis/kubevirt.io/v1/namespaces/default/virtualmachines/" + name:
			_ = json.NewEncoder(w).Encode(vm)
		case keyPairPath:
			_ = json.NewEncoder(w).Encode(keyPairs)
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(metav1.Status{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Status"},
				Status:   metav1.StatusFailure,
				Reason:   metav1.StatusReasonNotFound,
				Code:     http.StatusNotFound,
			})
		}
	}))
	defer server.Close()
	kubeConfig := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: %s
users:
- name: test
contexts:
- name: test
  context:
    cluster: test
    user: test
current-context: test
`, server.URL)
	meta := &config.Config{KubeConfig: base64.StdEncoding.EncodeToString([]byte(kubeConfig))}

	testcases := []struct {
		name     string
		run      func(ctx context.Context, d *schema.ResourceData) error
		expected []string
	}{
		{
			name: "read",
			run: func(ctx context.Context, d *schema.ResourceData) error {
				if diags := resourceVirtualMachineRead(ctx, d, meta); diags.HasError() {
					return fmt.Errorf("%v", diags)
				}
				return nil
			},
			expected: []string{http.MethodGet + " " + keyPairPath},
		},
		{
			name: "polling",
			run: func(ctx context.Context, d *schema.ResourceData) error {
				_, _, err := resourceVirtualMachineRefresh(ctx, d, meta, namespace, name, "")()
				return err
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			requests = nil
			mu.Unlock()
			d := schema.TestResourceDataRaw(t, ResourceVirtualMachine().Schema, map[string]interface{}{
				constants.FieldVirtualMachineCloudInit: []interface{}{
					map[string]interface{}{constants.FieldCloudInitUserData: udPackages},
				},
			})
			d.SetId(helper.BuildID(namespace, name))
			if err := tc.run(context.Background(), d); err != nil {
				t.Fatal(err)
			}

			var keyPairRequests []string
			mu.Lock()
			for _, request := range requests {
				if strings.Contains(request, "/keypairs") {
					keyPairRequests = append(keyPairRequests, request)
				}
			}
			mu.Unlock()
			if strings.Join(keyPairRequests, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("key pair requests = %q, want %q", keyPairRequests, tc.expected)
			}
			if got := d.Get(constants.FieldVirtualMachineCloudInit + ".0." + constants.FieldCloudInitUserData); got != udPackages {
				t.Errorf("user_data = %q, want %q", got, udPackages)
			}
		})
	}
}
