package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	bootstrapDefaultUser        = "admin"
	bootstrapDefaultTTL         = "60000"
	bootstrapDefaultSessionDesc = "Terraform bootstrap admin session"
)

func ResourceBootstrap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBootstrapCreate,
		ReadContext:   resourceBootstrapRead,
		DeleteContext: resourceBootstrapDelete,
		Schema:        Schema(),
	}
}

func resourceBootstrapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	if !c.Bootstrap {
		return diag.FromErr(fmt.Errorf("harvester_bootstrap just available on bootstrap mode"))
	}

	kubeConfig, err := homedir.Expand(d.Get(constants.FieldProviderKubeConfig).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// login to get token
	tokenID, token, err := bootstrapLogin(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tokenID)

	// change password
	log.Printf("Doing change password")
	if d.Get(constants.FieldShouldUpdatePassword).(bool) {
		initialPassword := d.Get(constants.FieldBootstrapInitialPassword).(string)
		password := d.Get(constants.FieldBootstrapPassword).(string)
		changePasswordURL := fmt.Sprintf("%s/%s", c.APIURL, "v3/users?action=changepassword")
		changePasswordData := `{"currentPassword":"` + initialPassword + `","newPassword":"` + password + `"}`
		changePasswordResp, err := util.DoPost(changePasswordURL, changePasswordData, "", true, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
		if err != nil {
			return diag.FromErr(err)
		}
		if changePasswordResp.StatusCode != http.StatusOK {
			return diag.Errorf("failed to change password, status code %d", changePasswordResp.StatusCode)
		}
	}

	// get kubeconfig
	log.Printf("Doing generate kubeconfig")
	genKubeConfigURL := fmt.Sprintf("%s/%s", c.APIURL, "v1/management.cattle.io.clusters/local?action=generateKubeconfig")
	genKubeConfigResp, err := util.DoPost(genKubeConfigURL, "", "", true, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return diag.FromErr(err)
	}
	if genKubeConfigResp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to generate kubeconfig, status code %d", genKubeConfigResp.StatusCode)
	}

	genKubeConfigBody, err := util.GetJSONBody(genKubeConfigResp)
	if err != nil {
		return diag.FromErr(err)
	}
	if genKubeConfigBody["config"] == nil {
		return diag.FromErr(fmt.Errorf("failed to generate kubeconfig"))
	}
	kubeConfigContent := genKubeConfigBody["config"].(string)

	// write kubeconfig
	if err = os.WriteFile(kubeConfig, []byte(kubeConfigContent), 0600); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBootstrapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	if !c.Bootstrap {
		return diag.FromErr(fmt.Errorf("[ERROR] harvester_bootstrap just available on bootstrap mode"))
	}

	kubeConfig, err := homedir.Expand(d.Get(constants.FieldProviderKubeConfig).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// login to get token
	_, token, err := bootstrapLogin(d, c)
	if err != nil {
		log.Printf("[INFO] Bootstrap is unable to login to Harvester")
		d.SetId("")
		return diag.FromErr(err)
	}

	log.Printf("Doing generate kubeconfig")
	genKubeConfigURL := fmt.Sprintf("%s/%s", c.APIURL, "v1/management.cattle.io.clusters/local?action=generateKubeconfig")
	genKubeConfigResp, err := util.DoPost(genKubeConfigURL, "", "", true, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return diag.FromErr(err)
	}
	if genKubeConfigResp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to generate kubeconfig, status code %d", genKubeConfigResp.StatusCode)
	}

	genKubeConfigBody, err := util.GetJSONBody(genKubeConfigResp)
	if err != nil {
		return diag.FromErr(err)
	}
	if genKubeConfigBody["config"] == nil {
		return diag.FromErr(fmt.Errorf("failed to generate kubeconfig"))
	}
	kubeConfigContent := genKubeConfigBody["config"].(string)

	// write kubeconfig
	if err = os.WriteFile(kubeConfig, []byte(kubeConfigContent), 0600); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceBootstrapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func bootstrapLogin(d *schema.ResourceData, c *config.Config) (string, string, error) {
	initialPassword := d.Get(constants.FieldBootstrapInitialPassword).(string)

	log.Printf("Doing login with initial password")
	tokenID, token, err := doUserLogin(c.APIURL, bootstrapDefaultUser, initialPassword, bootstrapDefaultTTL, bootstrapDefaultSessionDesc, "", true)
	if err == nil {
		err = d.Set(constants.FieldShouldUpdatePassword, true)
		return tokenID, token, err
	}

	log.Printf("Doing login with password")
	password := d.Get(constants.FieldBootstrapPassword).(string)
	tokenID, token, err = doUserLogin(c.APIURL, bootstrapDefaultUser, password, bootstrapDefaultTTL, bootstrapDefaultSessionDesc, "", true)
	if err == nil {
		err = d.Set(constants.FieldShouldUpdatePassword, false)
		return tokenID, token, err
	}
	return "", "", err
}

func doUserLogin(url, user, pass, ttl, desc, cacert string, insecure bool) (string, string, error) {
	loginURL := url + "/v3-public/localProviders/local?action=login"
	loginData := `{"username": "` + user + `", "password": "` + pass + `", "ttl": ` + ttl + `, "description": "` + desc + `"}`
	loginHead := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	// Login with user and pass
	loginResp, err := util.DoPost(loginURL, loginData, cacert, insecure, loginHead)
	if err != nil {
		return "", "", err
	}
	if loginResp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("can't login successfully, status code %d", loginResp.StatusCode)
	}

	loginBody, err := util.GetJSONBody(loginResp)
	if err != nil {
		return "", "", err
	}

	if loginBody["type"].(string) != "token" || loginBody["token"] == nil {
		return "", "", fmt.Errorf("doing  user logging: %s %s", loginBody["type"].(string), loginBody["code"].(string))
	}

	return loginBody["id"].(string), loginBody["token"].(string), nil
}
