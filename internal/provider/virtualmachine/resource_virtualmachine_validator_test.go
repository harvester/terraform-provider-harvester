package virtualmachine

import (
	"fmt"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_inCloudConfig(t *testing.T) {
	//nolint:goconst
	tests := []struct {
		name        string
		parentKey   any
		parent      any
		key         any
		value       string
		expectNotOk bool
	}{
		{
			name:      "complex case 1",
			parentKey: "",
			parent: map[string]any{
				"ssh_pwauth": true,
				"users": []any{
					map[string]any{
						"name": "root",
						"ssh_authorized_keys": []any{
							"ssh key content",
						},
					},
				},
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "complex case 2",
			parentKey: "users",
			parent: []any{
				map[string]any{
					"name": "root",
					"ssh_authorized_keys": []any{
						"ssh key content",
					},
				},
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "complex case 3",
			parentKey: "users",
			parent: map[string]any{
				"name": "root",
				"ssh_authorized_keys": []any{
					"ssh key content",
				},
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "complex case 4",
			parentKey: "ssh_authorized_keys",
			parent: []any{
				"ssh key content",
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:        "complex case 5",
			parentKey:   "ssh_authorized_keys",
			parent:      "ssh key content",
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "case map[string]any simple valie",
			parentKey: "ssh_authorized_keys",
			parent: map[string]any{
				"ssh_authorized_keys": "ssh key content",
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "case map[string]any list value",
			parentKey: "ssh_authorized_keys",
			parent: map[string]any{
				"ssh_authorized_keys": []any{
					"ssh key content",
				},
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "case []any simple value",
			parentKey: "list",
			parent: []any{
				map[string]any{
					"ssh_authorized_keys": "ssh key content",
				},
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:      "case []any list value",
			parentKey: "list",
			parent: []any{
				map[string]any{
					"ssh_authorized_keys": []any{
						"ssh key content",
					},
				},
			},
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:        "case any simple value found",
			parentKey:   "ssh_authorized_keys",
			parent:      "ssh key content",
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: false,
		},
		{
			name:        "case any simple value wrong value",
			parentKey:   "ssh_authorized_keys",
			parent:      "not ssh key content",
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: true,
		},
		{
			name:        "case any simple value wrong parent key",
			parentKey:   "some_parent",
			parent:      "ssh key content",
			key:         "ssh_authorized_keys",
			value:       "ssh key content",
			expectNotOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := inCloudConfig(tt.parentKey, tt.parent, tt.key, tt.value); (ok != true) != tt.expectNotOk {
				t.Errorf("inCloudConfig() ok = %v, expectNotOk = %v", ok, tt.expectNotOk)
			}
		})
	}
}

func Test_checkKeyPairsInUserData(t *testing.T) {
	type args struct {
		userdataContent []byte
		keyPairs        []*harvsterv1.KeyPair
	}
	testSSHPublicKey := "ssh key content"
	testWrongSSHPublicKey := "ssh key wrong content"
	testKeyPairs := []*harvsterv1.KeyPair{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-key",
				Namespace: "test-namespace",
			},
			Spec: harvsterv1.KeyPairSpec{
				PublicKey: testSSHPublicKey,
			},
		},
	}
	testRootKeyTemplate := `
ssh_authorized_keys:
  - %s
package_update: true
packages:
  - qemu-guest-agent
runcmd:
  - - systemctl
    - enable
    - '--now'
    - qemu-ga
`
	testUsersKeyTemplate := `
chpasswd:
  list: |
    root:linux
  expire: false
ssh_pwauth: true
users:
  - name: root
    ssh_authorized_keys:
      - %s
package_update: true
packages:
  - qemu-guest-agent
runcmd:
  - - systemctl
    - enable
    - '--now'
    - qemu-ga
`
	testNoKeyContent := `
chpasswd:
  list: |
    root:linux
  expire: false
ssh_pwauth: true
users:
  - name: root
package_update: true
packages:
  - qemu-guest-agent
runcmd:
  - - systemctl
    - enable
    - '--now'
    - qemu-ga
`
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "correct ssh_authorized_keys in root",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: fmt.Appendf([]byte{}, testRootKeyTemplate, testSSHPublicKey),
			},
			wantErr: false,
		},
		{
			name: "wrong ssh_authorized_keys in root",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: fmt.Appendf([]byte{}, testRootKeyTemplate, testWrongSSHPublicKey),
			},
			wantErr: true,
		},
		{
			name: "correct ssh_authorized_keys in users",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: fmt.Appendf([]byte{}, testUsersKeyTemplate, testSSHPublicKey),
			},
			wantErr: false,
		},
		{
			name: "wrong ssh_authorized_keys in users",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: fmt.Appendf([]byte{}, testUsersKeyTemplate, testWrongSSHPublicKey),
			},
			wantErr: true,
		},
		{
			name: "no ssh_authorized_keys",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: []byte(testNoKeyContent),
			},
			wantErr: true,
		},
		{
			name: "empty content",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: []byte{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkKeyPairsInUserData(tt.args.userdataContent, tt.args.keyPairs); (err != nil) != tt.wantErr {
				t.Errorf("checkKeyPairsInUserData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
