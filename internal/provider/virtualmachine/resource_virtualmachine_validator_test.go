package virtualmachine

import (
	"fmt"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
)

func Test_checkKeyPairsInUserData(t *testing.T) {
	type args struct {
		userdataContent []byte
		keyPairs        []*harvsterv1.KeyPair
	}
	testSSHPublicKey := "ssh key content"
	testWrongSSHPublicKey := "ssh key wrong content"
	testKeyPairs := []*harvsterv1.KeyPair{
		{
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
				userdataContent: []byte(fmt.Sprintf(testRootKeyTemplate, testSSHPublicKey)),
			},
			wantErr: false,
		},
		{
			name: "wrong ssh_authorized_keys in root",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: []byte(fmt.Sprintf(testRootKeyTemplate, testWrongSSHPublicKey)),
			},
			wantErr: true,
		},
		{
			name: "correct ssh_authorized_keys in users",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: []byte(fmt.Sprintf(testUsersKeyTemplate, testSSHPublicKey)),
			},
			wantErr: false,
		},
		{
			name: "wrong ssh_authorized_keys in users",
			args: args{
				keyPairs:        testKeyPairs,
				userdataContent: []byte(fmt.Sprintf(testUsersKeyTemplate, testWrongSSHPublicKey)),
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
