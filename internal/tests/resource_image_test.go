package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccImageName         = "test-acc-foo"
	testAccImageResourceName = constants.ResourceTypeImage + "." + testAccImageName
	testAccImageDescription  = "Terraform Harvester image acceptance test"
	testAccImageDisplayName  = "foo"
	testAccImageSourceType   = "download"

	testAccImageURL = "https://dl-cdn.alpinelinux.org/alpine/v3.23/releases/cloud/nocloud_alpine-3.23.0-x86_64-bios-tiny-r0.qcow2"

	// Crypto test constants
	testAccImageCryptoName         = "test-acc-crypto-foo"
	testAccImageCryptoResourceName = constants.ResourceTypeImage + "." + testAccImageCryptoName
	testAccImageCryptoDisplayName  = "crypto-foo"
	testAccImageCryptoDescription  = "Terraform Harvester crypto image acceptance test"
	testAccImageCryptoSourceType   = "clone"
	testAccImageCryptoOp           = "encrypt"

	testAccCryptoSourceImageName         = "test-acc-crypto-source"
	testAccCryptoSourceImageResourceName = constants.ResourceTypeImage + "." + testAccCryptoSourceImageName
	testAccCryptoSourceImageDisplayName  = "test-acc-crypto-source"
	testAccCryptoSourceImageNamespace    = "default"
	testAccCryptoSecretName              = "crypto"

	testAccImageConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
	storage_class_name = "harvester-longhorn"
}
`

	testAccImageCryptoConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
	storage_class_name = "encryption"
	%s = {
		%s = "%s"
		%s = "%s"
		%s = "%s"
	}
	%s
}
`
)

func buildImageConfig(name, description, displayName, sourceType, url string) string {
	return fmt.Sprintf(testAccImageConfigTemplate, constants.ResourceTypeImage, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldImageDisplayName, displayName,
		constants.FieldImageSourceType, sourceType,
		constants.FieldImageURL, url)
}

func buildImageCryptoConfig(name, description, displayName, sourceType, cryptoOp, sourceImageName, sourceImageNamespace string) string {
	if sourceType == "download" {
		return fmt.Sprintf(testAccImageCryptoConfigTemplate, constants.ResourceTypeImage, name,
			constants.FieldCommonName, name,
			constants.FieldCommonDescription, description,
			constants.FieldImageDisplayName, displayName,
			constants.FieldImageSourceType, sourceType,
			constants.FieldImageSecurityParameters,
			constants.FieldImageCryptoOperation, cryptoOp,
			constants.FieldImageSourceImageName, sourceImageName,
			constants.FieldImageSourceImageNamespace, sourceImageNamespace,
			fmt.Sprintf("  %s = \"%s\"", constants.FieldImageURL, testAccImageURL),
		)
	}
	return fmt.Sprintf(testAccImageCryptoConfigTemplate, constants.ResourceTypeImage, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldImageDisplayName, displayName,
		constants.FieldImageSourceType, sourceType,
		constants.FieldImageSecurityParameters,
		constants.FieldImageCryptoOperation, cryptoOp,
		constants.FieldImageSourceImageName, sourceImageName,
		constants.FieldImageSourceImageNamespace, sourceImageNamespace,
		"",
	)
}

func buildCryptoStorageClass() string {
	return fmt.Sprintf(`
resource "harvester_storageclass" "encryption" {
  name = "encryption"

  parameters = {
    "migratable"                                       = "true"
    "numberOfReplicas"                                 = "1"
    "staleReplicaTimeout"                              = "30"
    "encrypted"                                        = "true"
    "csi.storage.k8s.io/node-publish-secret-name"      = "%s"
    "csi.storage.k8s.io/node-publish-secret-namespace" = "default"
    "csi.storage.k8s.io/node-stage-secret-name"        = "%s"
    "csi.storage.k8s.io/node-stage-secret-namespace"   = "default"
    "csi.storage.k8s.io/provisioner-secret-name"       = "%s"
    "csi.storage.k8s.io/provisioner-secret-namespace"  = "default"
  }
}
`,
		testAccCryptoSecretName,
		testAccCryptoSecretName,
		testAccCryptoSecretName,
	)
}

func buildCryptoSourceImage() string {
	return fmt.Sprintf(`
resource "harvester_image" "%s" {
  name = "%s"
	namespace = "%s"

	display_name = "%s"

	source_type = "download"
	url = "%s"
}
`,
		testAccCryptoSourceImageName,
		testAccCryptoSourceImageName,
		testAccCryptoSourceImageNamespace,
		testAccCryptoSourceImageDisplayName,
		testAccImageURL,
	)
}

func buildCryptoStorageClassConfigAndSourceImage() string {
	storageClass := buildCryptoStorageClass()
	sourceImage := buildCryptoSourceImage()
	return fmt.Sprintf("%s\n\n%s", storageClass, sourceImage)
}

func TestAccImage_basic(t *testing.T) {
	var (
		image *harvsterv1.VirtualMachineImage
		ctx   = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      buildImageConfig(testAccImageName, testAccImageDescription, "", testAccImageSourceType, testAccImageURL),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`%s must not be empty`, constants.FieldImageDisplayName)),
			},
			{
				Config:      buildImageConfig(testAccImageName, testAccImageDescription, testAccImageDisplayName, testAccImageSourceType, ""),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected "%s" url to not be empty`, constants.FieldImageURL)),
			},
			{
				Config: buildImageConfig(testAccImageName, testAccImageDescription, testAccImageDisplayName, testAccImageSourceType, testAccImageURL),
				Check: resource.ComposeTestCheckFunc(
					testAccImageExists(ctx, testAccImageResourceName, image),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonName, testAccImageName),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonDescription, testAccImageDescription),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldImageSourceType, testAccImageSourceType),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldImageURL, testAccImageURL),
				),
			},
		},
	})
}

func testAccImageExists(ctx context.Context, n string, image *harvsterv1.VirtualMachineImage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource %s not found. ", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource %s ID not set. ", n)
		}

		id := rs.Primary.ID
		c, err := testAccProvider.Meta().(*config.Config).K8sClient()
		if err != nil {
			return err
		}

		namespace, name, err := helper.IDParts(id)
		if err != nil {
			return err
		}
		foundImage, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		image = foundImage
		return nil
	}
}

func testAccCheckImageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccProvider.Meta().(*config.Config).K8sClient()
		if err != nil {
			return err
		}
		for _, rs := range s.RootModule().Resources {
			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			switch t := rs.Type; t {
			case constants.ResourceTypeImage:
				imageStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
					return c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
				})
				stateConf := getStateChangeConf(imageStateRefreshFunc)
				if _, err = stateConf.WaitForStateContext(ctx); err != nil {
					return fmt.Errorf(
						"[ERROR] waiting for image (%s) to be removed: %s", rs.Primary.ID, err)
				}
			case constants.ResourceTypeStorageClass:
				scStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
					return c.KubeClient.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
				})
				stateConf := getStateChangeConf(scStateRefreshFunc)
				if _, err = stateConf.WaitForStateContext(ctx); err != nil {
					return fmt.Errorf(
						"[ERROR] waiting for storage class (%s) to be removed: %s", rs.Primary.ID, err)
				}
			}
		}

		err = c.KubeClient.
			CoreV1().
			Secrets(testAccCryptoSourceImageNamespace).
			Delete(ctx, testAccCryptoSecretName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		return nil
	}
}

func testAccCryptoCreateSecret(t *testing.T) {
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal(err)
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAccCryptoSecretName,
			Namespace: testAccCryptoSourceImageNamespace,
		},
		StringData: map[string]string{
			"CRYPTO_KEY_VALUE":    "your-encryption-passphrase-here",
			"CRYPTO_KEY_CIPHER":   "aes-xts-plain64",
			"CRYPTO_KEY_HASH":     "sha256",
			"CRYPTO_KEY_PROVIDER": "secret",
			"CRYPTO_KEY_SIZE":     "256",
			"CRYPTO_PBKDF":        "argon2i",
		},
	}

	_, err = c.KubeClient.
		CoreV1().
		Secrets(testAccCryptoSourceImageNamespace).
		Create(context.Background(), &secret, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatal(err)
	}
}

func TestAccImage_upload_no_filepath(t *testing.T) {
	var (
		ctx = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource %s "test-upload-no-path" {
	%s = "test-upload-no-path"
	%s = "test-upload-no-path"
	%s = "upload"
}
`, constants.ResourceTypeImage,
					constants.FieldCommonName,
					constants.FieldImageDisplayName,
					constants.FieldImageSourceType),
				ExpectError: regexp.MustCompile(`must specify file_path when source_type is 'upload'`),
			},
		},
	})
}

func TestAccImage_upload_wrong_sourcetype(t *testing.T) {
	var (
		ctx = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource %s "test-upload-wrong-type" {
	%s = "test-upload-wrong-type"
	%s = "test-upload-wrong-type"
	%s = "download"
	%s = "/tmp/some-file.qcow2"
	%s = "https://example.com/image.qcow2"
}
`, constants.ResourceTypeImage,
					constants.FieldCommonName,
					constants.FieldImageDisplayName,
					constants.FieldImageSourceType,
					constants.FieldImageFilePath,
					constants.FieldImageURL),
				ExpectError: regexp.MustCompile(`file_path must not be set when source_type is`),
			},
		},
	})
}

func TestAccImage_crypto_invalid(t *testing.T) {
	var (
		ctx = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildImageCryptoConfig(
					testAccImageCryptoName,
					testAccImageDescription,
					testAccImageCryptoDisplayName,
					"download",
					"encrypt",
					testAccCryptoSourceImageName,
					testAccCryptoSourceImageNamespace,
				),
				ExpectError: regexp.MustCompile(`security parameters can only be set when source type is 'clone'`),
			},
		},
	},
	)
}

func TestAccImage_crypto(t *testing.T) {
	var (
		sourceImage *harvsterv1.VirtualMachineImage
		image       *harvsterv1.VirtualMachineImage
		ctx         = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCryptoCreateSecret(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildCryptoStorageClassConfigAndSourceImage(),
				Check: resource.ComposeTestCheckFunc(
					testAccImageExists(ctx, testAccCryptoSourceImageResourceName, sourceImage),
				),
			},
			{
				Config: fmt.Sprintf("%s\n\n%s", buildCryptoStorageClassConfigAndSourceImage(), buildImageCryptoConfig(
					testAccImageCryptoName,
					testAccImageCryptoDescription,
					testAccImageCryptoDisplayName,
					testAccImageCryptoSourceType,
					testAccImageCryptoOp,
					testAccCryptoSourceImageName,
					testAccCryptoSourceImageNamespace,
				)),
				Check: resource.ComposeTestCheckFunc(
					testAccImageExists(ctx, testAccImageCryptoResourceName, image),
					resource.TestCheckResourceAttr(testAccImageCryptoResourceName, constants.FieldCommonName, testAccImageCryptoName),
					resource.TestCheckResourceAttr(testAccImageCryptoResourceName, constants.FieldCommonDescription, testAccImageCryptoDescription),
					resource.TestCheckResourceAttr(testAccImageCryptoResourceName, constants.FieldImageSourceType, testAccImageCryptoSourceType),
					resource.TestCheckResourceAttr(testAccImageCryptoResourceName, fmt.Sprintf("%s.%s", constants.FieldImageSecurityParameters, constants.FieldImageCryptoOperation), "encrypt"),
					resource.TestCheckResourceAttr(testAccImageCryptoResourceName, fmt.Sprintf("%s.%s", constants.FieldImageSecurityParameters, constants.FieldImageSourceImageName), testAccCryptoSourceImageName),
					resource.TestCheckResourceAttr(testAccImageCryptoResourceName, fmt.Sprintf("%s.%s", constants.FieldImageSecurityParameters, constants.FieldImageSourceImageNamespace), testAccCryptoSourceImageNamespace),
				),
			},
			{
				Config: buildCryptoStorageClassConfigAndSourceImage(),
				Check: resource.ComposeTestCheckFunc(
					testAccImageExists(ctx, testAccCryptoSourceImageResourceName, sourceImage),
				),
			},
		},
	})
}
