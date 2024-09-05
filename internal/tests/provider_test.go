package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/harvester/terraform-provider-harvester/internal/provider"
)

var (
	testAccProviders         map[string]*schema.Provider
	testAccProviderFactories map[string]func() (*schema.Provider, error)
	testAccProvider          *schema.Provider
	testAccProviderConfigure sync.Once
)

const (
	ProviderNameHarvester = "harvester"

	testAccResourceStateRemoved = "removed"
	testAccResourceStateExist   = "exist"
	testAccResourceStateError   = "error"
)

func init() {
	testAccProvider = provider.Provider()

	testAccProviders = map[string]*schema.Provider{
		ProviderNameHarvester: testAccProvider,
	}

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderNameHarvester: func() (*schema.Provider, error) { return provider.Provider(), nil }, //nolint:unparam
	}
}

func testAccPreCheck(t *testing.T) {
	testAccProviderConfigure.Do(func() {
		err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func getStateChangeConf(refresh resource.StateRefreshFunc) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{testAccResourceStateExist},
		Target:     []string{testAccResourceStateRemoved},
		Refresh:    refresh,
		Timeout:    2 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
}

func getResourceStateRefreshFunc(getResourceFunc func() (interface{}, error)) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := getResourceFunc()
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, testAccResourceStateRemoved, nil
			}
			return nil, testAccResourceStateError, err
		}
		return obj, testAccResourceStateExist, nil
	}
}
