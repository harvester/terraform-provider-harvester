package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	bootstrapDefaultTTL         = 60000
	bootstrapDefaultSessionDesc = "Terraform bootstrap admin session"
)

type loginRequestPayload struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ResponseType string `json:"responseType"`
	TTL          int    `json:"ttl"`
	Description  string `json:"description"`
}

type loginResponsePayload struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Token string `json:"token"`
	Code  string `json:"code"`
}

type changePasswordPayload struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type generateKubeConfigResponsePayload struct {
	Config string `json:"config"`
}

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

	u, err := url.Parse(d.Get(constants.FieldBootstrapAPIURL).(string))
	if err != nil {
		return diag.FromErr(err)
	}
	apiURL := u.String()

	kubeConfig, err := homedir.Expand(d.Get(constants.FieldProviderKubeConfig).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// login to get token
	tokenID, token, err := bootstrapLogin(apiURL, d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tokenID)

	// change password
	log.Printf("Doing change password")
	if d.Get(constants.FieldShouldUpdatePassword).(bool) {
		initialPassword := d.Get(constants.FieldBootstrapInitialPassword).(string)
		password := d.Get(constants.FieldBootstrapPassword).(string)
		changePasswordURL := fmt.Sprintf("%s/%s", apiURL, "v3/users?action=changepassword")
		changePasswordData, err := json.Marshal(changePasswordPayload{
			CurrentPassword: initialPassword,
			NewPassword:     password,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to marshal change password data: %v", err))
		}
		changePasswordResp, err := util.DoPost(changePasswordURL, string(changePasswordData), "", true, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
		if err != nil {
			return diag.FromErr(err)
		}
		if changePasswordResp.StatusCode != http.StatusOK {
			return diag.Errorf("failed to change password, status code %d", changePasswordResp.StatusCode)
		}
	}

	// get kubeconfig
	log.Printf("Doing generate kubeconfig")
	genKubeConfigURL := fmt.Sprintf("%s/%s", apiURL, "v1/management.cattle.io.clusters/local?action=generateKubeconfig")
	genKubeConfigResp, err := util.DoPost(genKubeConfigURL, "", "", true, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return diag.FromErr(err)
	}
	if genKubeConfigResp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to generate kubeconfig, status code %d", genKubeConfigResp.StatusCode)
	}

	var generateKubeConfigResponseData generateKubeConfigResponsePayload
	if err := json.NewDecoder(genKubeConfigResp.Body).Decode(&generateKubeConfigResponseData); err != nil {
		return diag.FromErr(fmt.Errorf("failed to decode generate kubeconfig response: %v", err))
	}
	if generateKubeConfigResponseData.Config == "" {
		return diag.FromErr(fmt.Errorf("failed to generate kubeconfig"))
	}

	// write kubeconfig
	if err = os.WriteFile(kubeConfig, []byte(generateKubeConfigResponseData.Config), 0600); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBootstrapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	if !c.Bootstrap {
		return diag.FromErr(fmt.Errorf("[ERROR] harvester_bootstrap just available on bootstrap mode"))
	}

	u, err := url.Parse(d.Get(constants.FieldBootstrapAPIURL).(string))
	if err != nil {
		return diag.FromErr(err)
	}
	apiURL := u.String()

	kubeConfig, err := homedir.Expand(d.Get(constants.FieldProviderKubeConfig).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// login to get token
	_, token, err := bootstrapLogin(apiURL, d, c)
	if err != nil {
		log.Printf("[INFO] Bootstrap is unable to login to Harvester")
		d.SetId("")
		return diag.FromErr(err)
	}

	log.Printf("Doing generate kubeconfig")
	genKubeConfigURL := fmt.Sprintf("%s/%s", apiURL, "v1/management.cattle.io.clusters/local?action=generateKubeconfig")
	genKubeConfigResp, err := util.DoPost(genKubeConfigURL, "", "", true, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return diag.FromErr(err)
	}
	if genKubeConfigResp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to generate kubeconfig, status code %d", genKubeConfigResp.StatusCode)
	}

	var generateKubeConfigResponseData generateKubeConfigResponsePayload
	if err := json.NewDecoder(genKubeConfigResp.Body).Decode(&generateKubeConfigResponseData); err != nil {
		return diag.FromErr(fmt.Errorf("failed to decode generate kubeconfig response: %v", err))
	}
	if generateKubeConfigResponseData.Config == "" {
		return diag.FromErr(fmt.Errorf("failed to generate kubeconfig"))
	}

	// write kubeconfig
	if err = os.WriteFile(kubeConfig, []byte(generateKubeConfigResponseData.Config), 0600); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceBootstrapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func bootstrapLogin(apiURL string, d *schema.ResourceData, c *config.Config) (string, string, error) {
	initialPassword := d.Get(constants.FieldBootstrapInitialPassword).(string)

	log.Printf("Doing login with initial password")
	tokenID, token, err := DoUserLogin(apiURL, bootstrapDefaultUser, initialPassword, bootstrapDefaultTTL, bootstrapDefaultSessionDesc, "", true)
	if err == nil {
		err = d.Set(constants.FieldShouldUpdatePassword, true)
		return tokenID, token, err
	}

	log.Printf("Doing login with password")
	password := d.Get(constants.FieldBootstrapPassword).(string)
	tokenID, token, err = DoUserLogin(apiURL, bootstrapDefaultUser, password, bootstrapDefaultTTL, bootstrapDefaultSessionDesc, "", true)
	if err == nil {
		err = d.Set(constants.FieldShouldUpdatePassword, false)
		return tokenID, token, err
	}
	return "", "", err
}

func DoUserLogin(url, user, pass string, ttl int, desc, cacert string, insecure bool) (string, string, error) {
	loginURL := url + "/v3-public/localProviders/local?action=login"
	loginData, err := json.Marshal(loginRequestPayload{
		Username:     user,
		Password:     pass,
		ResponseType: "token",
		TTL:          ttl,
		Description:  desc,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal login data: %v", err)
	}

	loginHead := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	// Login with user and pass
	loginResp, err := util.DoPost(loginURL, string(loginData), cacert, insecure, loginHead)
	if err != nil {
		return "", "", err
	}
	if loginResp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("can't login successfully, status code %d", loginResp.StatusCode)
	}

	var loginResponseData loginResponsePayload
	if err = json.NewDecoder(loginResp.Body).Decode(&loginResponseData); err != nil {
		return "", "", fmt.Errorf("failed to decode login response: %v", err)
	}

	if loginResponseData.Type != "token" || loginResponseData.Token == "" {
		return "", "", fmt.Errorf("doing user login: %s %s", loginResponseData.Type, loginResponseData.Code)
	}

	return loginResponseData.ID, loginResponseData.Token, nil
}
