package client

import (
	"encoding/base64"

	"github.com/mitchellh/go-homedir"

	harvloadbalancerclient "github.com/harvester/harvester-load-balancer/pkg/generated/clientset/versioned"
	harvnetworkclient "github.com/harvester/harvester-network-controller/pkg/generated/clientset/versioned"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/harvester/harvester/pkg/generated/clientset/versioned/scheme"
	"github.com/rancher/wrangler/v3/pkg/kubeconfig"
	kubeschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	storageclient "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	RestConfig                  *rest.Config
	KubeVirtSubresourceClient   *rest.RESTClient
	KubeClient                  *kubernetes.Clientset
	StorageClassClient          *storageclient.StorageV1Client
	HarvesterClient             *harvclient.Clientset
	HarvesterNetworkClient      *harvnetworkclient.Clientset
	HarvesterLoadbalancerClient *harvloadbalancerclient.Clientset
}

func NewClient(kubeConfig, kubeContext string) (*Client, error) {
	var (
		restConfig *rest.Config
		err        error
	)

	if restConfig, err = restConfigFromBase64(kubeConfig); err != nil {
		if restConfig, err = restConfigFromFile(kubeConfig, kubeContext); err != nil {
			return nil, err
		}
	}

	copyConfig := rest.CopyConfig(restConfig)
	copyConfig.GroupVersion = &kubeschema.GroupVersion{Group: "subresources.kubevirt.io", Version: "v1"}
	copyConfig.APIPath = "/apis"
	copyConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restClient, err := rest.RESTClientFor(copyConfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	storageClassClient, err := storageclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	harvClient, err := harvclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	harvNetworkClient, err := harvnetworkclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	harvLoadbalancerClient, err := harvloadbalancerclient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		RestConfig:                  restConfig,
		KubeVirtSubresourceClient:   restClient,
		KubeClient:                  kubeClient,
		StorageClassClient:          storageClassClient,
		HarvesterClient:             harvClient,
		HarvesterNetworkClient:      harvNetworkClient,
		HarvesterLoadbalancerClient: harvLoadbalancerClient,
	}, nil
}

func restConfigFromFile(kubeConfig, kubeContext string) (*rest.Config, error) {
	clientConfigPath, err := homedir.Expand(kubeConfig)
	if err != nil {
		return nil, err
	}

	clientConfig := kubeconfig.GetNonInteractiveClientConfigWithContext(clientConfigPath, kubeContext)
	return clientConfig.ClientConfig()
}

func restConfigFromBase64(kubeConfigBase64 string) (*rest.Config, error) {
	bytes, err := base64.StdEncoding.DecodeString(kubeConfigBase64)
	if err != nil {
		return nil, err
	}
	return clientcmd.RESTConfigFromKubeConfig(bytes)
}
